package biz

import (
	pb "base-server/api/gen/go/base_api/v1"
	"base-server/internal/data/ent"
	"base-server/internal/tools"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/emptypb"
	"strconv"
	"strings"
	"time"
)

// Base is a Base model.
type Base struct {
	Hello string
}

// BaseRepo is a Base repo.
type BaseRepo interface {
	Login(context.Context, *pb.LoginRequest) (string, error)
	FindUserByID(ctx context.Context, id *uuid.UUID) (*ent.User, error)

	GetDeptLeafsChildren(ctx context.Context, id int64) ([]*ent.Dept, error)
	GetDeptChildren(ctx context.Context, id int64) ([]*ent.Dept, error)
	GetDeptById(ctx context.Context, id int64) (*ent.Dept, error)
	GetRolesByDept(ctx context.Context, id int64) ([]*ent.Role, error)
	GetRolesFromUser(ctx context.Context, user1 *ent.User) ([]*ent.Role, error)
	GetUsersByDept(ctx context.Context, id int64) ([]*ent.User, error)

	ChangePassword(ctx context.Context, uid *uuid.UUID, passwordOld, passwordNew string) error

	GetUserList(ctx context.Context, deptId int64, req *pb.GetUserParams) ([]*ent.User, int64, error)
	AddUser(ctx context.Context, req *pb.UserListItem) (*ent.User, error)
	UpdateUser(ctx context.Context, id *uuid.UUID, req *pb.UserListItem) (*ent.User, error)
	DeleteByID(ctx context.Context, id *uuid.UUID) error

	GetAllRoleList(ctx context.Context, req *pb.RolePageParams) ([]*ent.Role, error)
	GetRole(ctx context.Context, id int64) (*ent.Role, error)
	AddRole(ctx context.Context, req *pb.RoleListItem) (*ent.Role, error)
	UpdateRole(ctx context.Context, deptId int64, req *pb.RoleListItem) (*ent.Role, error)
	DelRole(ctx context.Context, id int64) error

	GetMenuList(ctx context.Context) ([]*ent.Menu, error)
	CreateMenu(ctx context.Context, menu *ent.Menu) (*ent.Menu, error)
	UpdateMenu(ctx context.Context, id int64, menu *ent.Menu) (*ent.Menu, error)
	DeleteMenu(ctx context.Context, id int64) error

	GetDeptList(ctx context.Context) ([]*ent.Dept, error)
	AddDept(ctx context.Context, req *pb.DeptListItem) (*ent.Dept, error)
	UpdateDept(ctx context.Context, deptId int64, req *pb.DeptListItem) (*ent.Dept, error)
	DelDept(ctx context.Context, id int64) error

	IsUserExistsByUserName(ctx context.Context, req *pb.IsUserExistsRequest) (*ent.User, error)

	GetApiList(ctx context.Context, req *pb.GetApiPageParams) ([]*ent.ApiResources, int64, error)
	GetApi(ctx context.Context, id string) (*ent.ApiResources, error)
	AddApi(ctx context.Context, req *ent.ApiResources) (*ent.ApiResources, error)
	UpdateApi(ctx context.Context, req *ent.ApiResources) (*ent.ApiResources, error)
	DelApi(ctx context.Context, id string) error

	GetResourceList(ctx context.Context, req *pb.GetResourcePageParams) ([]*ent.Resource, int64, error)
	AddResource(ctx context.Context, req *ent.Resource) (*ent.Resource, error)
	GetResource(ctx context.Context, id string) (*ent.Resource, error)
	UpdateResource(ctx context.Context, req *ent.Resource) (*ent.Resource, error)
	DelResource(ctx context.Context, id string) error
}

// BaseUsecase is a Base usecase.
type BaseUsecase struct {
	repo BaseRepo
	auth *AuthUsecase
	log  *log.Helper
}

// NewBaseUsecase new a Base usecase.
func NewBaseUsecase(repo BaseRepo, logger log.Logger, auth *AuthUsecase) *BaseUsecase {
	return &BaseUsecase{repo: repo, auth: auth, log: log.NewHelper(logger)}
}

func (uc *BaseUsecase) ReLoadPolicy(ctx context.Context) error {
	return uc.auth.ReLoadPolicy()
}

// Login 登陆，存在返回Id
func (uc *BaseUsecase) Login(ctx context.Context, req *pb.LoginRequest) (string, error) {
	return uc.repo.Login(ctx, req)
}

// GenerateToken 生成Token
func (uc *BaseUsecase) GenerateToken(ctx context.Context, uid, key string) (*pb.LoginReply, error) {
	now := time.Now()
	// 生成accessToken
	claims := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, jwtv5.MapClaims{
		"user_id": uid,
		"sub":     uid,
		"aud":     "login",
		"exp":     now.Add(30 * time.Minute).Unix(), // 过期时间（30分钟后过期）
		"nbf":     now.Unix(),                       // 生效时间
		"iat":     now.Unix(),                       // 颁发时间
	})
	accessToken, err := claims.SignedString([]byte(key))
	if err != nil {
		return nil, errors.New("")
	}

	// 生成refreshToken，提前5分钟生效
	claims = jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, jwtv5.MapClaims{
		"user_id": uid,
		"sub":     uid,
		"aud":     "refresh",
		"exp":     now.Add(60 * time.Minute).Unix(), // 过期时间（60分钟后过期）
		"nbf":     now.Add(25 * time.Minute).Unix(), // 生效时间（accessToken过期前5分钟才能生效）
		"iat":     now.Unix(),                       // 颁发时间
		// 这里需要配合缓存，暂时没写
		"jti": uuid.New().String(), // 唯一标识符，主要用来作为一次性 token，从而回避重放（replay）攻击
	})
	refreshToken, err := claims.SignedString([]byte(key))
	if err != nil {
		return nil, errors.New("")
	}

	return &pb.LoginReply{
		UserId:       uid,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GetUserInfo 获取用户信息
func (uc *BaseUsecase) GetUserInfo(ctx context.Context, uuidString string) (*ent.User, error) {
	uid, err := uuid.Parse(uuidString)
	if err != nil {
		fmt.Println(err)
		uc.log.Debug("string to uuid err")
		return nil, err
	}
	return uc.repo.FindUserByID(ctx, &uid)
}

func (uc *BaseUsecase) CreateMenuTree(menuList []*ent.Menu) []*pb.SysMenuListItem {
	items := make([]*pb.SysMenuListItem, 0)
	for _, menu := range menuList {
		if menu.Pid == 0 {
			items = append(items, entMenuToMenu(menu))
			continue
		}

		//key := uc.BuildMenuTree(&reqs.MenuList, menu)
		key := uc.BuildMenuTree1(&items, menu)
		if !key {
			// 如果菜单不属于任何父级，则将其加入顶层列表
			items = append(items, entMenuToMenu(menu))
		}
	}
	return items
}

// CreateRouteMenuTree 创建菜单树
func (uc *BaseUsecase) CreateRouteMenuTree(ctx context.Context) (*pb.GetSysMenuListReply, error) {
	res := &pb.GetSysMenuListReply{Items: []*pb.SysMenuListItem{}}
	menuList, err := uc.repo.GetMenuList(ctx)
	if err != nil {
		return nil, err
	}

	var menus []int32

	uid := tools.GetUserId(ctx)
	user, err := uc.GetUserInfo(ctx, uid)
	if err != nil {
		return nil, err
	}

	// 获取用户本身所在域的角色
	roles, err := uc.repo.GetRolesFromUser(ctx, user)
	if err != nil {
		return nil, err
	}
	// 取并集
	for _, role := range roles {
		// 单独判断root角色权限
		if role.Value == "root" {
			for _, menu := range menuList {
				menus = append(menus, int32(menu.ID))
			}
			break
		}
		menus = append(menus, role.Menus...)
	}

	items := make([]*ent.Menu, 0)

	for _, menu := range menuList {
		if !menu.Status {
			continue
		}
		for _, item := range menus {
			if int64(item) == menu.ID {
				items = append(items, menu)

				//for _, menu := range menuList {
				//	if menu.Pid == 0 {
				//		res.Items = append(res.Items, menuToRoute(menu))
				//		continue
				//	}
				//
				//	key := uc.BuildMenuTree(&res.Items, menu)
				//	//key := uc.BuildMenuTree1(&items, menu)
				//	if !key {
				//		// 如果菜单不属于任何父级，则将其加入顶层列表
				//		res.Items = append(res.Items, menuToRoute(menu))
				//	}
				//}

			}
		}
	}

	res.Items = uc.CreateMenuTree(items)

	return res, err
}

// BuildMenuTree1 构造菜单树
func (uc *BaseUsecase) BuildMenuTree1(menuList *[]*pb.SysMenuListItem, menu *ent.Menu) bool {
	for _, item := range *menuList {
		if item.Children != nil {
			key := uc.BuildMenuTree1(&item.Children, menu)
			if key {
				return key
			}
		}
		if item.Id == menu.Pid {
			if item.Children == nil {
				item.Children = []*pb.SysMenuListItem{}
			}
			item.Children = append(item.Children, entMenuToMenu(menu))
			return true
		}
	}
	return false
}

func entMenuToMenu(menu *ent.Menu) *pb.SysMenuListItem {
	RouteItem := pb.SysMenuListItem{Meta: &pb.Meta{}}
	copier.Copy(&RouteItem, menu)
	RouteItem.Id = menu.ID
	RouteItem.Pid = menu.Pid
	//RouteItem.Status = func(status bool) int64 {
	//	if status {
	//		tmp := int64(1)
	//		return tmp
	//	} else {
	//		tmp := int64(0)
	//		return tmp
	//	}
	//}(menu.Status)

	tmpStatus := int32(0)
	if menu.Status {
		tmpStatus = 1
	}
	RouteItem.Status = &tmpStatus

	maxNumOfOpenTab := int64(menu.MaxNumOfOpenTab)

	RouteItem.Meta.Title = menu.Title
	RouteItem.Meta.Icon = menu.Icon
	RouteItem.Meta.Order = int64(menu.Order)
	RouteItem.Meta.Link = &menu.Link
	RouteItem.Meta.IframeSrc = &menu.IframeSrc
	RouteItem.Meta.IgnoreAccess = menu.IgnoreAccess
	RouteItem.Meta.KeepAlive = &menu.Keepalive
	RouteItem.Meta.ActivePath = &menu.ActivePath
	RouteItem.Meta.MaxNumOfOpenTab = &maxNumOfOpenTab
	RouteItem.Meta.HideInMenu = &menu.HideInMenu
	RouteItem.Meta.HideInTab = &menu.HideInTab
	RouteItem.Meta.HideInBreadcrumb = &menu.HideInBreadcrumb
	RouteItem.Meta.HideChildrenInMenu = &menu.HideChildrenInMenu

	RouteItem.Meta.Authority = strings.Split(menu.Authority, ",")

	return &RouteItem
}

func menuToEntMenu(menu *pb.SysMenuListItem) *ent.Menu {

	ptrToStr := func(ptr *string) string {
		if ptr == nil {
			return ""
		} else {
			return *ptr
		}
	}
	ptrToBool := func(ptr *bool) bool {
		if ptr == nil {
			return false
		} else {
			return *ptr
		}
	}
	ptrToInt := func(ptr *int64) int64 {
		if ptr == nil {
			return 0
		} else {
			return *ptr
		}
	}

	RouteItem := ent.Menu{}
	copier.Copy(&RouteItem, menu)
	RouteItem.ID = menu.Id
	RouteItem.Pid = menu.Pid
	RouteItem.Status = func(status *int32) bool {
		if *status == 1 {
			return true
		} else {
			return false
		}
	}(menu.Status)

	RouteItem.Title = menu.Meta.Title
	RouteItem.Icon = menu.Meta.Icon
	RouteItem.Order = int32(menu.Meta.Order)
	RouteItem.Link = ptrToStr(menu.Meta.Link)
	RouteItem.IframeSrc = ptrToStr(menu.Meta.IframeSrc)
	RouteItem.IgnoreAccess = menu.Meta.IgnoreAccess
	RouteItem.Keepalive = ptrToBool(menu.Meta.KeepAlive)
	RouteItem.ActivePath = ptrToStr(menu.Meta.ActivePath)
	RouteItem.MaxNumOfOpenTab = int16(ptrToInt(menu.Meta.MaxNumOfOpenTab))
	RouteItem.HideInMenu = ptrToBool(menu.Meta.HideInMenu)
	RouteItem.HideInTab = ptrToBool(menu.Meta.HideInTab)
	RouteItem.HideInBreadcrumb = ptrToBool(menu.Meta.HideInBreadcrumb)
	RouteItem.HideChildrenInMenu = ptrToBool(menu.Meta.HideChildrenInMenu)

	RouteItem.Authority = strings.Join(menu.Meta.Authority, ",")

	return &RouteItem
}

// BuildMenuTree 构造菜单树
func (uc *BaseUsecase) BuildMenuTree(menuList *[]*pb.RouteItem, menu *ent.Menu) bool {
	for _, item := range *menuList {
		if item.Children != nil {
			key := uc.BuildMenuTree(&item.Children, menu)
			if key {
				return key
			}
		}
		if item.Id == menu.Pid {
			if item.Children == nil {
				item.Children = []*pb.RouteItem{}
			}
			item.Children = append(item.Children, menuToRoute(menu))
			return true
		}
	}
	return false
}

func menuToRoute(menu *ent.Menu) *pb.RouteItem {
	RouteItem := pb.RouteItem{Meta: &pb.RouteMeta{}}
	copier.Copy(&RouteItem, menu)
	RouteItem.Id = menu.ID
	RouteItem.Pid = menu.Pid

	RouteItem.Meta.Title = menu.Title
	RouteItem.Meta.Icon = menu.Icon
	RouteItem.Meta.Order = int64(menu.Order)
	RouteItem.Meta.Link = menu.Link
	RouteItem.Meta.IframeSrc = menu.IframeSrc
	RouteItem.Meta.IgnoreAccess = menu.IgnoreAccess
	RouteItem.Meta.KeepAlive = menu.Keepalive
	RouteItem.Meta.ActivePath = menu.ActivePath
	RouteItem.Meta.MaxNumOfOpenTab = int64(menu.MaxNumOfOpenTab)
	RouteItem.Meta.HideInMenu = menu.HideInMenu
	RouteItem.Meta.HideInTab = menu.HideInTab
	RouteItem.Meta.HideInBreadcrumb = menu.HideInBreadcrumb
	RouteItem.Meta.HideChildrenInMenu = menu.HideChildrenInMenu

	RouteItem.Meta.Authority = strings.Split(menu.Authority, ",")

	return &RouteItem
}

//////////////////////////////////////////////////////////////////

// CreateDeptTree 创建部门
func (uc *BaseUsecase) CreateDeptTree(ctx context.Context) (*pb.GetDeptListReply, error) {
	reqs := &pb.GetDeptListReply{Items: []*pb.DeptListItem{}}
	var deptList []*ent.Dept

	deptList, err := uc.repo.GetDeptList(ctx)
	if err != nil {
		return nil, err
	}
	reqs.Total = int64(len(deptList))

	var deptForest []*Node
	for _, dept := range deptList {
		if dept.ID == 0 {
			continue
		}
		uc.BuildDeptTree(&deptForest, dept, true)
	}
	reqs.Items = ToDeptTree(deptForest, "")

	return reqs, err
}

type Node struct {
	Id       int64
	Pid      int64
	Value    interface{}
	Children []*Node
}

func deptToNode(value *ent.Dept) (node *Node) {
	return &Node{
		Id:       value.ID,
		Pid:      value.Pid,
		Value:    value,
		Children: nil,
	}
}

// BuildDeptTree 构造部门树
func (uc *BaseUsecase) BuildDeptTree(forest *[]*Node, dept *ent.Dept, top bool) bool {
	//if !dept.Status {
	//	return
	//}
	if dept.Pid == 0 {
		*forest = append(*forest, deptToNode(dept))
		return true
	}

	for _, item := range *forest {
		if item.Id == dept.Pid {
			if item.Children == nil {
				item.Children = []*Node{}
			}
			item.Children = append(item.Children, deptToNode(dept))
			return true
		}

		if item.Children != nil {
			key := uc.BuildDeptTree(&item.Children, dept, false)
			if key {
				return true
			}
		}
	}

	if top {
		// 如果部门不属于任何父级，则将其加入顶层列表
		*forest = append(*forest, deptToNode(dept))
		return true
	}
	return false
}

func ToDeptTree(forest []*Node, strPid string) []*pb.DeptListItem {
	var items []*pb.DeptListItem
	for _, item := range forest {
		//id := func() string {
		//	if strPid == "" {
		//		return strconv.FormatInt(item.Id, 10)
		//	} else {
		//		return strPid + "-" + strconv.FormatInt(item.Id, 10)
		//	}
		//}()
		id := strconv.FormatInt(item.Id, 10)
		items = append(items, &pb.DeptListItem{
			Id:      id,
			Pid:     strPid,
			Name:    item.Value.(*ent.Dept).Name,
			OrderNo: item.Value.(*ent.Dept).Sort,
			Remark:  item.Value.(*ent.Dept).Desc,
			Status: func() int32 {
				if item.Value.(*ent.Dept).Status {
					return 1
				} else {
					return 0
				}
			}(),
			CreateTime: item.Value.(*ent.Dept).CreateTime.Format(time.DateTime),
			Children: func() []*pb.DeptListItem {
				if item.Children == nil {
					return nil
				} else {
					// var deptList []*pb.DeptListItem
					return ToDeptTree(item.Children, id)
					// return deptList
				}
			}(),
		})
	}

	return items
}

// AddDept 添加部门
func (uc *BaseUsecase) AddDept(ctx context.Context, req *pb.DeptListItem) error {
	dept, err := uc.repo.AddDept(ctx, req)
	if err != nil {
		return err
	}

	// #添加了新的域
	if dept.Dom == 1 {
	}

	return nil
}

// DelDept 删除部门
func (uc *BaseUsecase) DelDept(ctx context.Context, deptId string) error {
	// 转移用户所属部门，转移用户所属域（删除域）
	id, err := tools.DeptStrSplitToInt(deptId)
	if err != nil {
		return err
	}

	dept, err := uc.repo.GetDeptById(ctx, id)
	if err != nil {
		return err
	}

	var childrenList []*ent.Dept
LOOP:
	childrenList, err = uc.repo.GetDeptLeafsChildren(ctx, id)
	if err != nil {
		return err
	}
	if len(childrenList) > 0 {
		for _, children := range childrenList {
			err := uc.DelDeptLink(ctx, children)
			if err != nil {
				return err
			}
			uc.repo.DelDept(ctx, children.ID)
		}
		goto LOOP
	}

	uc.DelDeptLink(ctx, dept)
	return uc.repo.DelDept(ctx, id)
}

// DelDeptLink 删除部门关联
func (uc *BaseUsecase) DelDeptLink(ctx context.Context, dept *ent.Dept) error {

	//deptId := dept.ID
	//// 刪除部门下的角色
	//roles, err := uc.repo.GetRolesByDept(ctx, deptId)
	//if err != nil {
	//	return err
	//}
	//for _, role := range roles {
	//	err := uc.DelRole(ctx, strconv.FormatInt(role.ID, 10))
	//	if err != nil {
	//		return err
	//	}
	//}
	//
	//// 转移部门下的用户(如果在域的部门下，转移到域，如果在域，转移到默认域)
	//users, err := uc.repo.GetUsersByDept(ctx, deptId)
	//if err != nil {
	//	return err
	//}
	//for _, user := range users {
	//	// 如果在域的部门下，转移到域，如果在域，转移到默认域
	//	if dept.Dom == 1 {
	//		// 转移到其他域，没有的话就转移到默认域
	//		depts, err := user.QueryDept().All(ctx)
	//		if err != nil {
	//			return err
	//		}
	//		for _, dept := range depts {
	//			if dept.ID != user.Dom && dept.Dom != user.Dom {
	//				if dept.Dom > 1 {
	//					user.Dom = dept.Dom
	//				} else if dept.Dom == 1 {
	//					user.Dom = dept.ID
	//				} else if dept.Dom == 0 {
	//					user.Dom = 0
	//				}
	//				break
	//			}
	//		}
	//		if user.Dom == dept.ID {
	//			user.Dom = 0
	//		}
	//	}
	//	// 部门关系转移
	//	if dept.Dom > 1 {
	//		user.Update().AddDeptIDs(user.Dom).Save(ctx)
	//	} else if dept.Dom == 1 {
	//		// #删除部门权限
	//		uc.auth.DelDom(strconv.FormatInt(dept.ID, 10))
	//		if user.Dom != 0 {
	//			user.Update().SetDom(user.Dom).AddDeptIDs(user.Dom).Save(ctx)
	//		} else {
	//			user.Update().AddDeptIDs(user.Dom).Save(ctx)
	//		}
	//	}
	//}
	return nil
}

// UpdateDept 更新部门
func (uc *BaseUsecase) UpdateDept(ctx context.Context, req *pb.DeptListItem) error {
	// #修改上级部门，权限变化
	deptId, err := tools.DeptStrSplitToInt(req.Id)
	if err != nil {
		return err
	}
	_, err = uc.repo.UpdateDept(ctx, deptId, req)
	if err != nil {
		return err
	}
	return nil
}

///////////////////////////////////////////////////////////// 系统角色管理

// GetAllRoleList 获取角色列表
func (uc *BaseUsecase) GetAllRoleList(ctx context.Context, req *pb.RolePageParams) (*pb.GetRoleListByPageReply, error) {
	// 获取某个域的角色
	res := &pb.GetRoleListByPageReply{Items: []*pb.RoleListItem{}}

	// # 权限验证

	roleList, err := uc.repo.GetAllRoleList(ctx, req)
	if err != nil {
		return nil, err
	}
	res.Total = int64(len(roleList))
	for i, item := range roleList {
		id := strconv.FormatInt(item.ID, 10)
		res.Items = append(res.Items, &pb.RoleListItem{
			Id:    id,
			Name:  item.Name,
			Value: item.Value,
			Status: func() int32 {
				if item.Status {
					return 1
				} else {
					return 0
				}
			}(),
			OrderNo:     strconv.Itoa(i),
			CreateTime:  item.CreateTime.Format(time.DateTime),
			Remark:      item.Desc,
			Permissions: item.Menus,
			ApiPermissions: func(item *ent.Role) []string {
				tmp := make([]string, 0)
				if item.Edges.Resource != nil {
					for _, resource := range item.Edges.Resource {
						if resource.Type == "api" {
							tmp = append(tmp, resource.ID)
						}
					}
				}
				return tmp
			}(item),
		})
	}

	return res, nil
}

// AddRole 添加角色
func (uc *BaseUsecase) AddRole(ctx context.Context, req *pb.RoleListItem) error {
	role, err := uc.repo.AddRole(ctx, req)
	if err != nil {
		return err
	}

	allResource, err := role.QueryResource().All(ctx)
	if err != nil {
		return err
	}
	// # 获取角色绑定的菜单，将菜单绑定的权限添加到对应角色（这里可以分类，然后分批插入提高性能）
	rulesMap := map[string][][]string{}
	for _, resource := range allResource {
		rulesMap[resource.Type] = append(rulesMap[resource.Type], []string{
			"role:" + role.Value,
			resource.Type + ":" + resource.Value,
			resource.Method,
		})
		//uc.auth.AddPolicy(role.Value, resource.Type, resource.Value, resource.Method)
	}
	for key, value := range rulesMap {
		uc.auth.AddPolicies(key, value)
	}

	return nil
}

// DelRole 删除角色
func (uc *BaseUsecase) DelRole(ctx context.Context, roleId string) error {
	id, _ := strconv.ParseInt(roleId, 10, 32)

	role, err := uc.repo.GetRole(ctx, id)
	if err != nil {
		return err
	}

	err = uc.repo.DelRole(ctx, id)
	if err != nil {
		return err
	}

	// 更新casbin权限
	uc.auth.DelRole(role.Value)

	return nil
}

// UpdateRole 更新角色
func (uc *BaseUsecase) UpdateRole(ctx context.Context, req *pb.RoleListItem) error {
	roleId, err := strconv.ParseInt(req.Id, 10, 32)
	if err != nil {
		return err
	}

	oldRole, err := uc.repo.GetRole(ctx, roleId)
	if err != nil {
		return err
	}
	//oldAllResource, err := oldRole.QueryResource().All(ctx)
	//if err != nil {
	//	return err
	//}

	newRole, err := uc.repo.UpdateRole(ctx, roleId, req)
	if err != nil {
		return err
	}

	// 删除旧资源权限
	uc.auth.DelRoleData(oldRole.Value)

	// 更新casbin权限
	if oldRole.Value != newRole.Value {
		uc.auth.UpdateRole(oldRole.Value, newRole.Value)
	}

	// 不止是角色值变化，还有角色关联的资源的变化
	allResource, err := newRole.QueryResource().All(ctx)
	if err != nil {
		return err
	}
	// # 获取角色绑定的菜单，将菜单绑定的权限添加到对应角色（这里可以分类，然后分批插入提高性能）
	rulesMap := map[string][][]string{}
	for _, resource := range allResource {
		rulesMap[resource.Type] = append(rulesMap[resource.Type], []string{
			"role:" + newRole.Value,
			resource.Type + ":" + resource.Value,
			resource.Method,
		})
		//uc.auth.AddPolicy(role.Value, resource.Type, resource.Value, resource.Method)
	}
	for key, value := range rulesMap {
		uc.auth.AddPolicies(key, value)
	}

	return nil
}

///////////////////////////////////////////////////////////// 系统菜单管理

// GetSysMenuList 获取菜单（非路由树）(系统菜单管理)
func (uc *BaseUsecase) GetSysMenuList(ctx context.Context) (*pb.GetSysMenuListReply, error) {
	res := &pb.GetSysMenuListReply{
		Items: make([]*pb.SysMenuListItem, 0),
	}
	menuList, err := uc.repo.GetMenuList(ctx)
	if err != nil {
		return nil, err
	}
	res.Items = uc.CreateMenuTree(menuList)
	return res, nil
}

func ToMenuTree(forest []*pb.SysMenuListItem) []*pb.SysMenuListItem {
	var items []*pb.SysMenuListItem
	for _, item := range forest {
		items = append(items, &pb.SysMenuListItem{
			Id:        item.Id,
			Component: item.Component,
			//Status: func() int64 {
			//	if item.Meta.HideInMenu {
			//		return 1
			//	} else {
			//		return 0
			//	}
			//}(),
			Status:   item.Status,
			AuthCode: "",
			Name:     item.Name,
			Path:     item.Path,
			Pid:      item.Pid,
			Redirect: nil,
			Type:     item.Type,
			Meta: &pb.Meta{
				Order:              item.Meta.Order,
				Icon:               item.Meta.Icon,
				Title:              item.Meta.Title,
				ActiveIcon:         nil,
				ActivePath:         item.Meta.ActivePath,
				AffixTab:           nil,
				AffixTabOrder:      nil,
				Badge:              nil,
				BadgeType:          nil,
				BadgeVariants:      nil,
				HideChildrenInMenu: nil,
				HideInBreadcrumb:   nil,
				HideInMenu:         nil,
				HideInTab:          nil,
				IframeSrc:          nil,
				Link:               nil,
				KeepAlive:          nil,
				MaxNumOfOpenTab:    nil,
				NoBasicLayout:      nil,
				OpenInNewWindow:    nil,
			},
			Children: func() []*pb.SysMenuListItem {
				if item.Children == nil {
					return nil
				} else {
					return ToMenuTree(item.Children)
				}
			}(),
			CreateTime: "",
		})
	}

	return items
}

func (uc *BaseUsecase) IsMenuNameExists(ctx context.Context, req *pb.IsMenuNameExistsRequest) (bool, error) {
	menuList, err := uc.repo.GetMenuList(ctx)
	if err != nil {
		return false, err
	}
	for _, menu := range menuList {
		if req.Name == menu.Name && menu.ID != req.Id {
			return true, nil
		}
	}
	return false, nil
}
func (uc *BaseUsecase) IsMenuPathExists(ctx context.Context, req *pb.IsMenuPathExistsRequest) (bool, error) {
	menuList, err := uc.repo.GetMenuList(ctx)
	if err != nil {
		return false, err
	}
	for _, menu := range menuList {
		if req.Path == menu.Path && menu.ID != req.Id {
			return true, nil
		}
	}
	return false, nil
}
func (uc *BaseUsecase) CreateMenu(ctx context.Context, req *pb.SysMenuListItem) (*emptypb.Empty, error) {
	menu := menuToEntMenu(req)
	_, err := uc.repo.CreateMenu(ctx, menu)
	return nil, err
}
func (uc *BaseUsecase) UpdateMenu(ctx context.Context, req *pb.SysMenuListItem) (*emptypb.Empty, error) {
	_, err := uc.repo.UpdateMenu(ctx, req.Id, menuToEntMenu(req))
	return nil, err
}
func (uc *BaseUsecase) DeleteMenu(ctx context.Context, req *pb.DeleteMenuRequest) (*emptypb.Empty, error) {
	err := uc.repo.DeleteMenu(ctx, req.Id)
	return nil, err
}

///////////////////////////////////////////////////////////// 系统用户管理

// GetUserList 获取账户列表 (系统用户管理)
func (uc *BaseUsecase) GetUserList(ctx context.Context, req *pb.GetUserParams) (*pb.GetUserListReply, error) {
	deptId, _ := tools.DeptStrSplitToInt(req.DeptId)

	userList, count, err := uc.repo.GetUserList(ctx, deptId, req)
	if err != nil {
		return nil, err
	}
	return &pb.GetUserListReply{
		Total: count,
		Items: func() []*pb.UserListItem {
			var items []*pb.UserListItem
			for _, user := range userList {
				var extension pb.UserExtension
				// 解析JSON字符串并填充结构体
				email := ""
				err := json.Unmarshal([]byte(user.Extension), &extension)
				if err != nil {
					fmt.Println("解析错误：", err)
				}
				email = extension.Email

				role := 0
				if user.Edges.Roles != nil {
					if len(user.Edges.Roles) > 0 {
						role = int(user.Edges.Roles[0].ID)
					}
				}

				items = append(items, &pb.UserListItem{
					Id:         user.ID.String(),
					Username:   user.Username,
					Avatar:     user.Avatar,
					Email:      email,
					Nickname:   user.Nickname,
					Remark:     user.Desc,
					Status:     int32(user.Status),
					Role:       int64(role),
					CreateTime: user.CreateTime.Format(time.DateTime),
				})
			}
			return items
		}(),
	}, nil
}

// AddUser 新增用户
func (uc *BaseUsecase) AddUser(ctx context.Context, req *pb.UserListItem) (*pb.UserListItem, error) {
	// 往数据库添加用户
	user, err := uc.repo.AddUser(ctx, req)
	if err != nil {
		return nil, err
	}

	roleList, err := user.QueryRoles().All(ctx)
	if err != nil {
		return nil, err
	}
	// 更新 casbin 权限
	var roles []string
	for _, role := range roleList {
		roles = append(roles, role.Value)
	}
	uc.auth.AddUserRoles(user.ID.String(), roles)

	return &pb.UserListItem{
		Id:         user.ID.String(),
		Username:   user.Username,
		Email:      user.Extension,
		Nickname:   user.Nickname,
		Remark:     user.Desc,
		Status:     int32(user.Status),
		CreateTime: user.CreateTime.Format(time.DateTime),
	}, nil
}

// UpdateUser 更新用户
func (uc *BaseUsecase) UpdateUser(ctx context.Context, req *pb.UserListItem) error {
	uid, _ := uuid.Parse(req.Id)
	user, err := uc.repo.UpdateUser(ctx, &uid, req)
	if err != nil {
		return err
	}
	roleList, err := user.QueryRoles().All(ctx)
	if err != nil {
		return err
	}
	// 更新 casbin 权限
	uc.auth.DelUser(uid.String())
	var roles []string
	for _, role := range roleList {
		roles = append(roles, role.Value)
	}
	uc.auth.AddUserRoles(uid.String(), roles)
	return nil
}

// DelUser 删除用户
func (uc *BaseUsecase) DelUser(ctx context.Context, uuidString string) error {
	uid, err := uuid.Parse(uuidString)
	if err != nil {
		uc.log.Debug("string to uid err")
		return err
	}

	err = uc.repo.DeleteByID(ctx, &uid)
	if err != nil {
		return err
	}
	// 删除用户
	uc.auth.DelUser(uuidString)
	return nil
}

// IsUserExist 检查用户是否存在，存在返回Id
func (uc *BaseUsecase) IsUserExist(ctx context.Context, req *pb.IsUserExistsRequest) (bool, error) {
	user, err := uc.repo.IsUserExistsByUserName(ctx, req)
	if err != nil {
		return true, err
	}
	if user != nil {
		if req.Id != user.ID.String() {
			return true, nil
		}
	}
	return false, nil
}

/////////////////////////////////////////////////////////////

// ChangePassword 修改密码
func (uc *BaseUsecase) ChangePassword(ctx context.Context, uuidString, passwordOld, passwordNew string) error {
	uid, err := uuid.Parse(uuidString)
	if err != nil {
		uc.log.Debug("string to uuid err")
		return err
	}
	return uc.repo.ChangePassword(ctx, &uid, passwordOld, passwordNew)
}

//// GetDom 获取当前用户所在的域
//func (uc *BaseUsecase) GetDom(ctx context.Context) int64 {
//	user, err := uc.GetUserInfo(ctx, tools.GetUserId(ctx))
//	if err != nil {
//		return 0
//	}
//	return user.Dom
//}

////////////////////////////////////////////////////////////////////

func (uc *BaseUsecase) GetApiList(ctx context.Context, req *pb.GetApiPageParams) (*pb.GetApiListByPageReply, error) {
	res := &pb.GetApiListByPageReply{
		Items: []*pb.ApiListItem{},
	}
	list, count, err := uc.repo.GetApiList(ctx, req)
	if err != nil {
		return nil, err
	}
	for _, api := range list {
		tmp := &pb.ApiListItem{}
		copier.Copy(tmp, api)

		res.Items = append(res.Items, tmp)
	}
	res.Total = count
	return res, nil
}

func (uc *BaseUsecase) AddApi(ctx context.Context, req *pb.ApiListItem) (*pb.ApiListItem, error) {
	api := &ent.ApiResources{}
	copier.Copy(api, req)
	entApi, err := uc.repo.AddApi(ctx, api)
	if err != nil {
		return nil, err
	}

	// 更新casbin权限
	uc.auth.AddApiToGroup(entApi.Path, entApi.ResourcesGroup)
	return nil, nil
}

func (uc *BaseUsecase) UpdateApi(ctx context.Context, req *pb.ApiListItem) (*pb.ApiListItem, error) {
	oldApi, err := uc.repo.GetApi(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	api := &ent.ApiResources{}
	copier.Copy(api, req)
	newApi, err := uc.repo.UpdateApi(ctx, api)
	if err != nil {
		return nil, err
	}

	// 更新casbin权限
	uc.auth.UpdateApiToGroup([]string{oldApi.Path, oldApi.ResourcesGroup}, []string{newApi.Path, newApi.ResourcesGroup})

	return &pb.ApiListItem{}, nil
}

func (uc *BaseUsecase) DelApi(ctx context.Context, req *pb.DeleteApi) error {
	api, err := uc.repo.GetApi(ctx, req.Id)
	if err != nil {
		return err
	}
	err = uc.repo.DelApi(ctx, req.Id)
	if err != nil {
		return err
	}

	// 更新casbin权限
	uc.auth.DelApiToGroup(api.Path, api.ResourcesGroup)

	return nil
}

func (uc *BaseUsecase) GetResourceList(ctx context.Context, req *pb.GetResourcePageParams) (*pb.GetResourceListByPageReply, error) {
	res := &pb.GetResourceListByPageReply{
		Items: []*pb.ResourceListItem{},
	}
	list, count, err := uc.repo.GetResourceList(ctx, req)
	if err != nil {
		return nil, err
	}
	for _, api := range list {
		tmp := &pb.ResourceListItem{}
		copier.Copy(tmp, api)

		res.Items = append(res.Items, tmp)
	}
	res.Total = count
	return res, nil
}

func (uc *BaseUsecase) AddResource(ctx context.Context, req *pb.ResourceListItem) (*pb.ResourceListItem, error) {
	api := &ent.Resource{}
	copier.Copy(api, req)
	_, err := uc.repo.AddResource(ctx, api)
	if err != nil {
		return nil, err
	}

	// 添加资源，资源无所属关系，不需要修改casbin权限
	return nil, nil
}

func (uc *BaseUsecase) UpdateResource(ctx context.Context, req *pb.ResourceListItem) (*pb.ResourceListItem, error) {
	data := &ent.Resource{}
	copier.Copy(data, req)
	_, err := uc.repo.UpdateResource(ctx, data)
	if err != nil {
		return nil, err
	}

	//uc.auth.UpdateDataPolicy([]string{}, []string{})

	return &pb.ResourceListItem{}, nil
}

func (uc *BaseUsecase) DelResource(ctx context.Context, req *pb.DeleteResource) error {
	data, err := uc.repo.GetResource(ctx, req.Id)
	if err != nil {
		return err
	}

	err = uc.repo.DelResource(ctx, req.Id)
	if err != nil {
		return err
	}

	// 更新casbin权限
	uc.auth.DelDataPolicy(data.Type, data.Value, data.Method)
	return nil
}
