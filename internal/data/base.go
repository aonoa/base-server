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
func (r *baseRepo) IsUserExistsByUserName(ctx context.Context, req *pb.IsUserExistsRequest) (*ent.User, error) {
	data, err := r.data.db.User.
		Query().
		Unique(false).
		Where(user.UsernameEQ(req.Username)).
		First(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	return data, err
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

// GetUserList 获取用户列表
func (r *baseRepo) GetUserList(ctx context.Context, deptId int64, req *pb.GetUserParams) ([]*ent.User, error) {
	query := r.data.db.User.Query()
	if req.Username != "" {
		query = query.Where(user.UsernameEQ(req.Username))
	}
	if req.Nickname != "" {
		query = query.Where(user.NicknameEQ(req.Nickname))
	}
	if deptId != 0 {
		query.Where(user.HasDeptWith(dept.IDEQ(deptId)))
	}
	query.WithRoles(func(query *ent.RoleQuery) {
		query.Select(role.FieldID, role.FieldName, role.FieldValue)
	})
	return query.All(ctx)
	// return r.data.db.User.Query().All(ctx)
}

// AddUser 新增用户
func (r *baseRepo) AddUser(ctx context.Context, req *pb.UserListItem) (*ent.User, error) {
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
		SetUsername(req.Username).
		SetAvatar("https://cdn.jsdelivr.net/gh/BaiMo-zyc/baimo.images@master/img/user-mini.png").
		SetPassword(req.Password).
		SetNickname(req.Nickname).
		SetStatus(int8(req.Status)).
		SetDesc(req.Remark).
		SetExtension(extensionStr).
		AddRoles(userRole).
		Save(ctx)
}

func (r *baseRepo) UpdateUser(ctx context.Context, id *uuid.UUID, req *pb.UserListItem) (*ent.User, error) {
	defer r.data.db.User.Query().All(entcache.Evict(ctx))
	userRole, _ := r.data.db.Role.Get(ctx, req.Role)
	return r.data.db.User.UpdateOneID(*id).
		SetUsername(req.Username).
		SetNickname(req.Nickname).
		SetStatus(int8(req.Status)).
		SetDesc(req.Remark).
		ClearRoles(). // 设计的是用户可以有多个角色，这里是为了降低复杂性，让他只能同时拥有一个角色
		AddRoles(userRole).
		Save(ctx)
}

// GetMenuList 获取菜单列表
func (r *baseRepo) GetMenuList(ctx context.Context) ([]*ent.Menu, error) {
	return r.data.db.Menu.Query().Order(menu.ByPid(), menu.ByOrder()).All(ctx)
}

func (r *baseRepo) CreateMenu(ctx context.Context, menu *ent.Menu) (*ent.Menu, error) {
	defer r.data.db.Menu.Query().All(entcache.Evict(ctx))
	return r.data.db.Menu.Create().CreateAll(menu).Save(ctx)
}
func (r *baseRepo) UpdateMenu(ctx context.Context, id int64, menu *ent.Menu) (*ent.Menu, error) {
	defer r.data.db.Menu.Query().All(entcache.Evict(ctx))
	return r.data.db.Menu.UpdateOneID(id).UpdateAll(menu).Save(ctx)
}
func (r *baseRepo) DeleteMenu(ctx context.Context, id int64) error {
	defer r.data.db.Menu.Query().All(entcache.NewContext(ctx))
	return r.data.db.Menu.DeleteOneID(id).Exec(entcache.Evict(ctx))
}

// GetDeptList 获取部门列表
func (r *baseRepo) GetDeptList(ctx context.Context) ([]*ent.Dept, error) {
	return r.data.db.Dept.Query().Order(dept.ByPid(func(options *sql.OrderTermOptions) {
		options.NullsFirst = true
	})).All(ctx)
}

// AddDept 添加部门
func (r *baseRepo) AddDept(ctx context.Context, req *pb.DeptListItem) (*ent.Dept, error) {
	sqlCmd := r.data.db.Dept.Create().
		SetName(req.Name).
		SetSort(req.OrderNo).
		SetStatus(func() bool {
			if req.Status == 0 {
				return false
			} else {
				return true
			}
		}()).
		SetDesc(req.Remark).
		SetExtension("").
		SetDom(req.Dom)

	pid, err := strconv.ParseInt(req.Pid, 10, 32)
	if err != nil {
		pid = 0
	}

	if pid > 0 {
		sqlCmd = sqlCmd.SetPid(pid)
	}

	return sqlCmd.Save(ctx)
}

// DelDept 删除部门
func (r *baseRepo) DelDept(ctx context.Context, id int64) error {
	return r.data.db.Dept.DeleteOneID(id).Exec(ctx)
}

// UpdateDept 更新部门
func (r *baseRepo) UpdateDept(ctx context.Context, deptId int64, req *pb.DeptListItem) (*ent.Dept, error) {
	sqlCmd := r.data.db.Dept.UpdateOneID(deptId).
		SetName(req.Name).
		SetSort(req.OrderNo).
		SetStatus(func() bool {
			if req.Status == 0 {
				return false
			} else {
				return true
			}
		}()).
		SetDesc(req.Remark).
		SetExtension("")

	pid, err := strconv.ParseInt(req.Pid, 10, 32)
	if err != nil {
		pid = 0
	}

	if pid > 0 {
		sqlCmd = sqlCmd.SetPid(pid)
	}

	return sqlCmd.Save(ctx)
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
	if req.Name != "" {
		query = query.Where(role.NameEQ(req.Name))
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
		SetName(req.Name).
		SetValue(req.Value).
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
		SetMenus(req.Permissions).
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
		SetName(req.Name).
		SetValue(req.Value).
		SetStatus(func() bool {
			if req.Status == 0 {
				return false
			} else {
				return true
			}
		}()).
		SetDesc(req.Remark).
		SetMenus(req.Permissions).
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
