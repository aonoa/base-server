package biz

import (
	pb "base-server/api/base_api/v1"
	"base-server/internal/conf"
	"base-server/internal/data/ent"
	"base-server/internal/tools"
	"base-server/internal/types"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
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
	IsUserExistsByUserName(context.Context, *pb.LoginRequest) (bool, error)
	Login(context.Context, *pb.LoginRequest) (string, error)
	GetAccountList(ctx context.Context, deptId int64, req *pb.AccountParams) ([]*ent.User, error)
	FindUserByID(ctx context.Context, id *uuid.UUID) (*ent.User, error)
	DeleteByID(ctx context.Context, id *uuid.UUID) error
	AddUser(ctx context.Context, req *pb.AccountListItem) (*ent.User, error)
	GetMenuList(ctx context.Context) ([]*ent.Menu, error)
	GetDeptList(ctx context.Context) ([]*ent.Dept, error)
	AddDept(ctx context.Context, req *pb.DeptListItem) (*ent.Dept, error)
	UpdateDept(ctx context.Context, deptId int64, req *pb.DeptListItem) (*ent.Dept, error)
	DelDept(ctx context.Context, id int64) error
	GetDeptLeafsChildren(ctx context.Context, id int64) ([]*ent.Dept, error)
	GetDeptChildren(ctx context.Context, id int64) ([]*ent.Dept, error)
	GetDeptById(ctx context.Context, id int64) (*ent.Dept, error)
	GetRolesByDept(ctx context.Context, id int64) ([]*ent.Role, error)
	GetRolesFromUser(ctx context.Context, user1 *ent.User) ([]*ent.Role, error)
	GetUsersByDept(ctx context.Context, id int64) ([]*ent.User, error)
	GetAllRoleList(ctx context.Context, req *pb.RolePageParams) ([]*ent.Role, error)
	AddRole(ctx context.Context, req *pb.RoleListItem) (*ent.Role, error)
	UpdateRole(ctx context.Context, deptId int64, req *pb.RoleListItem) (*ent.Role, error)
	DelRole(ctx context.Context, id int64) error
	ChangePassword(ctx context.Context, uid *uuid.UUID, passwordOld, passwordNew string) error
}

// BaseUsecase is a Base usecase.
type BaseUsecase struct {
	repo BaseRepo
	auth *AuthUsecase
	log  *log.Helper
	conf *conf.Menus
}

// NewBaseUsecase new a Base usecase.
func NewBaseUsecase(repo BaseRepo, logger log.Logger, auth *AuthUsecase, conf *conf.Menus) *BaseUsecase {
	return &BaseUsecase{repo: repo, auth: auth, log: log.NewHelper(logger), conf: conf}
}

// IsUserExists 检查用户是否存在，存在返回Id
func (uc *BaseUsecase) IsUserExists(ctx context.Context, req *pb.LoginRequest) (bool, error) {
	return uc.repo.IsUserExistsByUserName(ctx, req)
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
	// uuidString := "d1f7b7c1-c0b6-4707-aa17-5055b09b3ae8"
	uuid, err := uuid.Parse(uuidString)
	if err != nil {
		fmt.Println(err)
		uc.log.Debug("string to uuid err")
		return nil, err
	}
	return uc.repo.FindUserByID(ctx, &uuid)
}

// CreateMenuTree 创建菜单
func (uc *BaseUsecase) CreateMenuTree(ctx context.Context) (*pb.GetMenuListReply, error) {
	reqs := &pb.GetMenuListReply{MenuList: []*pb.RouteItem{}}
	menuList, err := uc.repo.GetMenuList(ctx)
	if err != nil {
		return nil, err
	}

	var menus []string

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
		menus = append(menus, strings.Split(role.Menu, ",")...)
	}

	//if user.Dom != 0 {
	//	// 获取用户本身所在域的角色
	//	roles, err := uc.repo.GetRolesFromUser(ctx, user)
	//	if err != nil {
	//		return nil, err
	//	}
	//	// 取并集
	//
	//	for _, role := range roles {
	//		menus = append(menus, strings.Split(role.Menu, ",")...)
	//	}
	//	if len(roles) == 0 {
	//		if uc.auth.HasRoleForUser(uid, "role:default", "dom:"+strconv.FormatInt(user.Dom, 10)) {
	//			menus = append(menus, strings.Split(uc.conf.DefaultMenus, ",")...)
	//		}
	//	}
	//} else {
	//	// 确定用户是否有角色
	//	if uc.auth.HasRoleForUser(uid, "role:admin", "dom:default") {
	//		for _, menu := range menuList {
	//			menus = append(menus, strconv.FormatInt(menu.ID, 10))
	//		}
	//	} else if uc.auth.HasRoleForUser(uid, "role:default", "dom:default") {
	//		menus = append(menus, strings.Split(uc.conf.DefaultMenus, ",")...)
	//	}
	//}

	//for _, menu := range menuList {
	//	//if !menu.Status {
	//	//	continue
	//	//}
	//	menus = append(menus, strconv.FormatInt(menu.ID, 10))
	//}

	for _, menu := range menuList {
		if !menu.Status {
			continue
		}
		for _, item := range menus {
			menuID, _ := strconv.ParseInt(item, 10, 64)
			if menuID == menu.ID {
				if menu.Pid == 0 {
					reqs.MenuList = append(reqs.MenuList, menuToRoute(menu))
					continue
				}

				key := uc.BuildMenuTree(&reqs.MenuList, menu)
				if !key {
					// 如果菜单不属于任何父级，则将其加入顶层列表
					reqs.MenuList = append(reqs.MenuList, menuToRoute(menu))
				}
			}
		}
	}

	return reqs, err
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

	RouteItem.Meta.Title = menu.Title
	RouteItem.Meta.Icon = menu.Icon
	RouteItem.Meta.Order = int64(menu.Order)
	RouteItem.Meta.Link = menu.Link
	RouteItem.Meta.IframeSrc = menu.IframeSrc
	RouteItem.Meta.IgnoreAccess = menu.IgnoreAuth
	RouteItem.Meta.KeepAlive = menu.Keepalive
	RouteItem.Meta.ActivePath = menu.ActivePath
	RouteItem.Meta.MaxNumOfOpenTab = int64(menu.MaxNumOfOpenTab)
	RouteItem.Meta.HideInMenu = menu.HideInMenu
	RouteItem.Meta.HideInTab = menu.HideInTab
	RouteItem.Meta.HideInBreadcrumb = menu.HideInBreadcrumb
	RouteItem.Meta.HideChildrenInMenu = menu.HideChildrenInMenu

	RouteItem.Meta.Authority = strings.Split(menu.Permission, ",")

	return &RouteItem
}

//////////////////////////////////////////////////////////////////

// CreateDeptTree 创建部门
func (uc *BaseUsecase) CreateDeptTree(ctx context.Context) (*pb.GetDeptListReply, error) {
	reqs := &pb.GetDeptListReply{Items: []*pb.DeptListItem{}}
	var deptList []*ent.Dept

	//uid := tools.GetUserId(ctx)
	//user, err := uc.GetUserInfo(ctx, uid)
	//if err != nil {
	//	return nil, err
	//}
	//if user.Dom == 0 {
	//	deptList, err = uc.repo.GetDeptList(ctx)
	//	if err != nil {
	//		return nil, err
	//	}
	//	reqs.Total = int64(len(deptList))
	//} else {
	//	dept, _ := uc.repo.GetDeptById(ctx, user.Dom)
	//	deptList = append(deptList, dept)
	//	depts, err := uc.repo.GetDeptChildren(ctx, user.Dom)
	//	if err != nil {
	//		return nil, err
	//	}
	//	deptList = append(deptList, depts...)
	//	reqs.Total = int64(len(deptList))
	//}
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
		id := func() string {
			if strPid == "" {
				return strconv.FormatInt(item.Id, 10)
			} else {
				return strPid + "-" + strconv.FormatInt(item.Id, 10)
			}
		}()
		items = append(items, &pb.DeptListItem{
			Id:       id,
			DeptName: item.Value.(*ent.Dept).Name,
			// OrderNo:  strconv.Itoa(item.Value.(*ent.Dept).Sort),
			OrderNo: int64(item.Value.(*ent.Dept).Sort),
			Remark:  item.Value.(*ent.Dept).Desc,
			Status: func() string {
				if item.Value.(*ent.Dept).Status {
					return "1"
				} else {
					return "0"
				}
			}(),
			CreateTime: item.Value.(*ent.Dept).CreateTime.Format("2006-01-02 15:04:05"),
			ParentDept: strPid,
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

// GetAllRoleList 获取角色列表
func (uc *BaseUsecase) GetAllRoleList(ctx context.Context, req *pb.RolePageParams) (*pb.GetRoleListByPageReply, error) {
	// 获取某个域的角色
	reqs := &pb.GetRoleListByPageReply{Items: []*pb.RoleListItem{}}

	// # 权限验证

	roleList, err := uc.repo.GetAllRoleList(ctx, req)
	if err != nil {
		return nil, err
	}
	reqs.Total = int64(len(roleList))
	for i, item := range roleList {
		id := strconv.FormatInt(item.ID, 10)
		reqs.Items = append(reqs.Items, &pb.RoleListItem{
			Id:        id,
			RoleName:  item.Name,
			RoleValue: item.Value,
			Status: func() int64 {
				if item.Status {
					return 1
				} else {
					return 0
				}
			}(),
			OrderNo:    strconv.Itoa(i),
			CreateTime: item.CreateTime.Format("2006-01-02 15:04:05"),
			Remark:     item.Desc,
			Menu:       strings.Split(item.Menu, ","),
		})
	}

	return reqs, nil
}

// AddRole 添加角色
func (uc *BaseUsecase) AddRole(ctx context.Context, req *pb.RoleListItem) error {
	deptId, err := tools.DeptStrSplitToInt(req.Dept)

	//// 获取所属域
	//var domId int64
	//if err != nil {
	//	domId = uc.GetDom(ctx)
	//} else {
	//	dept, err := uc.repo.GetDeptById(ctx, deptId)
	//	if err != nil {
	//		return err
	//	}
	//	domId = dept.Dom
	//}

	// # 检查是否有这个域添加角色权限
	if ok, err := uc.auth.EnforcePolicy(tools.GetUserId(ctx), string(types.PolicyType_Role)+":"+strconv.FormatInt(deptId, 10), "add"); ok {
		fmt.Println("ok ")
	} else {
		fmt.Println("err ")
		return err
	}

	// 添加到域
	_, err = uc.repo.AddRole(ctx, req)
	if err != nil {
		return err
	}
	// # 获取角色绑定的菜单，将菜单绑定的权限添加到对应角色

	return nil
}

// DelRole 删除角色
func (uc *BaseUsecase) DelRole(ctx context.Context, roleId string) error {
	// #删除角色权限
	//uc.auth.DelUserRole(strconv.FormatInt(uc.GetDom(ctx), 10), roleId)
	id, _ := strconv.ParseInt(roleId, 10, 32)
	return uc.repo.DelRole(ctx, id)
}

// UpdateRole 更新角色
func (uc *BaseUsecase) UpdateRole(ctx context.Context, req *pb.RoleListItem) error {
	roleId, err := strconv.ParseInt(req.Id, 10, 32)
	if err != nil {
		return err
	}
	_, err = uc.repo.UpdateRole(ctx, roleId, req)
	if err != nil {
		return err
	}
	return nil
}

// GetSysMenuList 获取菜单（非路由树）
func (uc *BaseUsecase) GetSysMenuList(ctx context.Context) (*pb.GetSysMenuListReply, error) {
	resp := &pb.GetSysMenuListReply{}
	tree, err := uc.CreateMenuTree(ctx)
	if err != nil {
		return nil, err
	}
	resp.Items = ToMenuTree(tree.MenuList)
	return resp, nil
}

func ToMenuTree(forest []*pb.RouteItem) []*pb.SysMenuListItem {
	var items []*pb.SysMenuListItem
	for i, item := range forest {
		items = append(items, &pb.SysMenuListItem{
			Id:       strconv.FormatInt(item.Id, 10),
			OrderNo:  strconv.Itoa(i),
			Icon:     item.Meta.Icon,
			MenuName: item.Name,
			Status: func() int64 {
				if item.Meta.HideInMenu {
					return 1
				} else {
					return 0
				}
			}(),
			// CreateTime: item.CreateTime.Format("2006-01-02 15:04:05"),
			Component: item.Component,
			Children: func() []*pb.SysMenuListItem {
				if item.Children == nil {
					return nil
				} else {
					return ToMenuTree(item.Children)
				}
			}(),
		})
	}

	return items
}

func (uc *BaseUsecase) GetAccountList(ctx context.Context, req *pb.AccountParams) (*pb.GetAccountListReply, error) {
	deptId, _ := tools.DeptStrSplitToInt(req.DeptId)

	userList, err := uc.repo.GetAccountList(ctx, deptId, req)
	if err != nil {
		return nil, err
	}
	return &pb.GetAccountListReply{
		Total: int64(len(userList)),
		Items: func() []*pb.AccountListItem {
			var items []*pb.AccountListItem
			for _, user := range userList {
				var extension pb.UserExtension
				// 解析JSON字符串并填充结构体
				email := ""
				err := json.Unmarshal([]byte(user.Extension), &extension)
				if err != nil {
					fmt.Println("解析错误：", err)
				}
				email = extension.Email
				items = append(items, &pb.AccountListItem{
					Id:         user.ID.String(),
					Account:    user.Username,
					Avatar:     user.Avatar,
					Email:      email,
					Nickname:   user.Nickname,
					Remark:     user.Desc,
					Status:     int64(user.Status),
					CreateTime: user.CreateTime.Format("2006-01-02 15:04:05"),
				})
			}
			return items
		}(),
	}, nil
}

// AddUser 新增用户
func (uc *BaseUsecase) AddUser(ctx context.Context, req *pb.AccountListItem) (*pb.AccountListItem, error) {
	// 往数据库添加用户
	user, err := uc.repo.AddUser(ctx, req)
	if err != nil {
		return nil, err
	}

	role := "default"
	if req.Role == 0 {
		role = "default"
	} else if req.Role == 2 {
		role = "admin"
	}

	// 设置用户角色
	uc.auth.AddUserRoles(user.ID.String(), []string{role})

	return &pb.AccountListItem{
		Id:         user.ID.String(),
		Account:    user.Username,
		Email:      user.Extension,
		Nickname:   user.Nickname,
		Remark:     user.Desc,
		Status:     int64(user.Status),
		CreateTime: user.CreateTime.Format("2006-01-02 15:04:05"),
	}, nil
}

// DelUser 删除用户
func (uc *BaseUsecase) DelUser(ctx context.Context, uuidString string) error {
	uuid, err := uuid.Parse(uuidString)
	if err != nil {
		fmt.Println(err)
		uc.log.Debug("string to uuid err")
		return err
	}

	err = uc.repo.DeleteByID(ctx, &uuid)
	if err != nil {
		return err
	}
	// 删除用户
	uc.auth.DelUser(uuidString)
	//uc.auth.DelUserRoles(uuidString, []string{"default"})
	return nil
}

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
