package data

import (
	"ariga.io/entcache"
	pb "base-server/api/base_api/v1"
	"base-server/internal/biz"
	"base-server/internal/data/ent"
	"base-server/internal/data/ent/dept"
	"base-server/internal/data/ent/menu"
	"base-server/internal/data/ent/role"
	"base-server/internal/data/ent/user"
	"context"
	"encoding/json"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"fmt"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"strconv"
	"strings"
)

type baseRepo struct {
	data *Data
	log  *log.Helper
}

// NewBaseRepo .
func NewBaseRepo(data *Data, logger log.Logger) biz.BaseRepo {
	return &baseRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// Login 根据登陆信息查找用户id
func (r *baseRepo) Login(ctx context.Context, req *pb.LoginRequest) (string, error) {
	id, err := r.data.db.User.
		Query().
		Unique(false).
		//Select(user.FieldID).
		Where(
			user.And(
				user.UsernameEQ(req.Username),
				user.Password(req.Password),
			),
		).FirstID(ctx)
	return id.String(), err
}

// IsUserExistsByUserName 根据用户名检查用户是否存在
func (r *baseRepo) IsUserExistsByUserName(ctx context.Context, req *pb.LoginRequest) (bool, error) {
	_, err := r.data.db.User.
		Query().
		Unique(false).
		//Select(user.FieldID).
		Where(user.UsernameEQ(req.Username)).
		FirstID(ctx)
	if ent.IsNotFound(err) || err != nil {
		return false, err
	}
	return true, err
}

// FindUserByID 根据用户id查找用户信息
func (r *baseRepo) FindUserByID(ctx context.Context, id *uuid.UUID) (*ent.User, error) {
	return r.data.db.User.Get(ctx, *id)
}

// DeleteByID 根据用户id删除用户信息
func (r *baseRepo) DeleteByID(ctx context.Context, id *uuid.UUID) error {
	defer r.data.db.User.Query().All(entcache.NewContext(ctx))
	return r.data.db.User.DeleteOneID(*id).Exec(entcache.Evict(ctx))
}

// GetAccountList 获取用户列表
func (r *baseRepo) GetAccountList(ctx context.Context, deptId int64, req *pb.AccountParams) ([]*ent.User, error) {
	query := r.data.db.User.Query()
	if req.Account != "" {
		query = query.Where(user.UsernameEQ(req.Account))
	}
	if req.Nickname != "" {
		query = query.Where(user.NicknameEQ(req.Nickname))
	}
	if deptId != 0 {
		query.Where(user.HasDeptWith(dept.IDEQ(deptId)))
	}
	return query.All(ctx)
	// return r.data.db.User.Query().All(ctx)
}

// AddUser 新增用户
func (r *baseRepo) AddUser(ctx context.Context, req *pb.AccountListItem) (*ent.User, error) {
	// 添加角色关系、添加部门关系
	userRole, _ := r.data.db.Role.Get(ctx, req.Role)

	var extension *pb.UserExtension
	extensionStr := ""
	extension = &pb.UserExtension{UserRole: []*pb.UserRole{
		{Role: userRole.Value, Menu: ""},
	}, Email: req.Email}
	extensionByte, _ := json.Marshal(extension)
	extensionStr = string(extensionByte)

	return r.data.db.User.Create().
		SetUsername(req.Account).
		SetAvatar("https://q1.qlogo.cn/g?b=qq&nk=190848757&s=640").
		SetPassword(req.Password).
		SetNickname(req.Nickname).
		SetStatus(int8(req.Status)).
		SetDesc(req.Remark).
		SetExtension(extensionStr).
		AddRoles(userRole).
		Save(ctx)
}

// 获取用户所有部门

// GetMenuList 获取菜单列表
func (r *baseRepo) GetMenuList(ctx context.Context) ([]*ent.Menu, error) {
	return r.data.db.Menu.Query().Order(menu.ByPid(), menu.ByOrder()).All(ctx)
}

// GetDeptList 获取部门列表
func (r *baseRepo) GetDeptList(ctx context.Context) ([]*ent.Dept, error) {
	return r.data.db.Dept.Query().Order(dept.BySort()).All(ctx)
}

// AddDept 添加部门
func (r *baseRepo) AddDept(ctx context.Context, req *pb.DeptListItem) (*ent.Dept, error) {
	// 找父节点
	nodes := strings.Split(req.ParentDept, "-")
	parent := nodes[len(nodes)-1]
	intPid, err := strconv.ParseInt(parent, 10, 32)
	if err != nil {
		return r.data.db.Dept.Create().
			SetName(req.DeptName).
			SetSort(func() int {
				intVar := int(req.OrderNo)
				return intVar
			}()).
			SetStatus(func() bool {
				if req.Status == "0" {
					return false
				} else {
					return true
				}
			}()).
			SetDesc(req.Remark).
			SetExtension("").
			SetDom(0).
			Save(ctx)
	} else {
		// 有父节点，判断dom是否符合调条件（dom中不能新建dom）
		// 获取父节点信息
		root, err := r.data.db.Dept.Get(ctx, intPid)
		if err != nil {
			return nil, err
		}

		// 0: 没有dom，1：当前为dom，n属于某dom
		if root.Dom > 0 && req.Dom == 1 {
			// 报错
			return nil, errors.New(10001, "已在域中", "已在域中")
		}
		// 添加域部门的子部门
		if root.Dom == 1 && req.Dom == 0 {
			req.Dom = root.ID
		}
		// 在域部门的子部门添加部门
		if root.Dom > 1 && req.Dom == 0 {
			req.Dom = root.Dom
		}

		return r.data.db.Dept.Create().
			//SetID(9).
			SetName(req.DeptName).
			SetSort(func() int {
				intVar := int(req.OrderNo)
				return intVar
			}()).
			SetStatus(func() bool {
				if req.Status == "0" {
					return false
				} else {
					return true
				}
			}()).
			SetDesc(req.Remark).
			SetExtension("").
			SetPid(intPid).
			SetDom(req.Dom).
			Save(ctx)
	}
}

// DelDept 删除部门
func (r *baseRepo) DelDept(ctx context.Context, id int64) error {
	return r.data.db.Dept.DeleteOneID(id).Exec(ctx)
}

// UpdateDept 更新部门
func (r *baseRepo) UpdateDept(ctx context.Context, deptId int64, req *pb.DeptListItem) (*ent.Dept, error) {
	// 找父节点
	nodes := strings.Split(req.ParentDept, "-")
	parent := nodes[len(nodes)-1]
	return r.data.db.Dept.UpdateOneID(deptId).
		SetName(req.DeptName).
		SetSort(func() int {
			intVar := int(req.OrderNo)
			//intVar, err := strconv.Atoi(req.OrderNo)
			//if err != nil {
			//	return 99
			//}
			return intVar
		}()).
		SetStatus(func() bool {
			if req.Status == "0" {
				return false
			} else {
				return true
			}
		}()).
		SetDesc(req.Remark).
		SetExtension("").
		SetPid(func() int64 {
			intVar, err := strconv.ParseInt(parent, 10, 32)
			if err != nil {
				return 99
			}
			return intVar
		}()).Save(ctx)
}

// GetDeptLeafsChildren 获取部门叶子节点
func (r *baseRepo) GetDeptLeafsChildren(ctx context.Context, id int64) ([]*ent.Dept, error) {
	root, err := r.data.db.Dept.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return root.QueryChildren().Where(dept.Not(dept.HasChildren())).All(ctx)
}

// GetDeptChildren 获取部门子节点
func (r *baseRepo) GetDeptChildren(ctx context.Context, id int64) ([]*ent.Dept, error) {
	root, err := r.data.db.Dept.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return root.QueryChildren().All(ctx)
}

// GetDeptById 根据部门ID获取部门
func (r *baseRepo) GetDeptById(ctx context.Context, id int64) (*ent.Dept, error) {
	return r.data.db.Dept.Get(ctx, id)
}

// GetRolesByDept 根据部门Id获取所有角色
func (r *baseRepo) GetRolesByDept(ctx context.Context, id int64) ([]*ent.Role, error) {
	root, err := r.data.db.Dept.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return root.QueryRoles().All(ctx)
}

// GetUsersByDept 根据部门Id获取所有用户
func (r *baseRepo) GetUsersByDept(ctx context.Context, id int64) ([]*ent.User, error) {
	root, err := r.data.db.Dept.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return root.QueryUsers().All(ctx)
}

// GetAllRoleList 获取角色列表
func (r *baseRepo) GetAllRoleList(ctx context.Context, req *pb.RolePageParams) ([]*ent.Role, error) {
	query := r.data.db.Role.Query()
	if req.RoleNme != "" {
		query = query.Where(role.NameEQ(req.RoleNme))
	}
	if req.Status == 1 {
		query = query.Where(role.StatusEQ(func() bool {
			if req.Status == 1 {
				return true
			} else {
				return false
			}
		}()))
	}
	return query.All(ctx)
}

// GetRolesFromUser 获取用户的角色列表
func (r *baseRepo) GetRolesFromUser(ctx context.Context, user1 *ent.User) ([]*ent.Role, error) {
	return user1.QueryRoles().All(ctx)
}

// AddRole 添加角色
func (r *baseRepo) AddRole(ctx context.Context, req *pb.RoleListItem) (*ent.Role, error) {
	defer r.data.db.Role.Query().All(entcache.Evict(ctx))
	// 先不考虑关系表
	return r.data.db.Role.Create().
		SetName(req.RoleName).
		SetValue(req.RoleValue).
		//SetSort(func() int {
		//	intVar := int(req.OrderNo)
		//	//intVar, err := strconv.Atoi(req.OrderNo)
		//	//if err != nil {
		//	//	return 99
		//	//}
		//	return intVar
		//}()).
		SetStatus(func() bool {
			if req.Status == 0 {
				return false
			} else {
				return true
			}
		}()).
		SetDesc(req.Remark).
		SetMenu(strings.Join(req.Menu, ",")).
		Save(entcache.Evict(ctx))
}

// DelRole 删除角色
func (r *baseRepo) DelRole(ctx context.Context, id int64) error {
	defer r.data.db.Role.Query().All(entcache.Evict(ctx))
	// 先不考虑关系表
	return r.data.db.Role.DeleteOneID(id).Exec(entcache.Evict(ctx))
}

// UpdateRole 更新角色
func (r *baseRepo) UpdateRole(ctx context.Context, roleId int64, req *pb.RoleListItem) (*ent.Role, error) {
	//defer r.data.db.Role.Query().All(entcache.Evict(ctx))
	// 先不考虑关系表
	return r.data.db.Role.UpdateOneID(roleId).
		SetName(req.RoleName).
		SetValue(req.RoleValue).
		SetStatus(func() bool {
			if req.Status == 0 {
				return false
			} else {
				return true
			}
		}()).
		SetDesc(req.Remark).
		SetMenu(strings.Join(req.Menu, ",")).
		Save(entcache.NewContext(entcache.Evict(ctx)))
}

func (r *baseRepo) ChangePassword(ctx context.Context, uid *uuid.UUID, passwordOld, passwordNew string) error {
	password, err := r.data.db.User.Query().Where(user.IDEQ(*uid)).Select(user.FieldPassword).String(ctx)
	if err != nil {
		return err
	}
	if password == passwordOld {
		r.data.db.User.UpdateOneID(*uid).SetPassword(passwordNew).Save(entcache.Evict(ctx))
		return nil
	}
	return errors.New(500, "password err", "password err")
}

func (r *baseRepo) Test(ctx context.Context) {
	b := sql.Dialect(dialect.Postgres)
	query, args := b.Select().
		From(b.Table("users")).
		Where(sql.In(
			"id",
			sql.Select("user_id").From(b.Table("cars")).Where(sql.EQ("car_model", "Tesla")),
		)).
		Query()
	fmt.Println(query, args)
	//r.data.db.Role.Query().Modify(func(s *sql.Selector) {
	//	s.From(sql.Table("table_01").Schema("dbname"))
	//})
	//tx, err := r.data.db.Tx(ctx)
	//if err != nil {
	//	return
	//}

}
