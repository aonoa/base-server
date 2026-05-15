package data

import (
	"context"
	dbsql "database/sql"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"ariga.io/entcache"
	v1 "base-server/api/gen/go/admin/service/v1"
	authv1 "base-server/api/gen/go/auth/service/v1"
	commonv1 "base-server/api/gen/go/common/service/v1"
	userv1 "base-server/api/gen/go/user/service/v1"
	"base-server/app/admin/service/internal/biz"
	"base-server/app/admin/service/internal/conf"
	"base-server/pkg/authx"
	"base-server/pkg/data/ent"
	"base-server/pkg/data/ent/apiresources"
	"base-server/pkg/data/ent/dept"
	"base-server/pkg/data/ent/menu"
	"base-server/pkg/data/ent/migrate"
	"base-server/pkg/data/ent/organization"
	"base-server/pkg/data/ent/organizationpermissionscope"
	"base-server/pkg/data/ent/projectionsourcestatus"
	"base-server/pkg/data/ent/resource"
	"base-server/pkg/data/ent/role"
	"base-server/pkg/data/ent/syslogrecord"
	"base-server/pkg/data/ent/userdeptmembership"
	"base-server/pkg/data/ent/userorganization"
	"base-server/pkg/data/ent/userrolebinding"
	"base-server/pkg/tools"

	"entgo.io/ent/dialect"
	sql "entgo.io/ent/dialect/sql"
	schema "entgo.io/ent/dialect/sql/schema"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewAdminRepo)

const defaultOrganizationID = "9f740c1b-0210-4e3a-858d-d128edea924d"

// Data .
type Data struct {
	db               *ent.Client
	sqlDB            *dbsql.DB
	dbDriver         string
	userConn         *grpc.ClientConn
	userClient       userv1.UserServiceClient
	authConn         *grpc.ClientConn
	authClient       authv1.AuthServiceClient
	projectionClient *authx.ProjectionClient
	commonConn       *grpc.ClientConn
	commonClient     commonv1.CommonServiceClient
}

// NewData .
func NewData(c *conf.Data, services *conf.Services, logger log.Logger) (*Data, func(), error) {
	helper := log.NewHelper(logger)
	db, err := dbsql.Open(c.Database.Driver, c.Database.Source)
	if err != nil {
		return nil, nil, err
	}
	drv := sql.OpenDB(toEntDialect(c.Database.Driver), db)
	sqlDrv := dialect.DebugWithContext(drv, func(ctx context.Context, args ...interface{}) {
		helper.WithContext(ctx).Info(args...)
	})
	client := ent.NewClient(ent.Driver(sqlDrv))
	if err := migrate.Create(context.Background(), migrate.NewSchema(sqlDrv), adminTables(), schema.WithForeignKeys(false)); err != nil {
		_ = client.Close()
		return nil, nil, err
	}
	if err := ensureDefaultOrganizationAfterMigration(context.Background(), db, c.Database.Driver); err != nil {
		_ = client.Close()
		return nil, nil, err
	}
	userConn, err := grpc.NewClient(services.User.GrpcEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = client.Close()
		return nil, nil, err
	}
	authConn, err := grpc.NewClient(services.Auth.GrpcEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = userConn.Close()
		_ = client.Close()
		return nil, nil, err
	}
	commonConn, err := grpc.NewClient(services.Common.GrpcEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = authConn.Close()
		_ = userConn.Close()
		_ = client.Close()
		return nil, nil, err
	}
	d := &Data{
		db:               client,
		sqlDB:            db,
		dbDriver:         c.Database.Driver,
		userConn:         userConn,
		userClient:       userv1.NewUserServiceClient(userConn),
		authConn:         authConn,
		authClient:       authv1.NewAuthServiceClient(authConn),
		projectionClient: authx.NewProjectionClient(authv1.NewAuthServiceClient(authConn), "admin"),
		commonConn:       commonConn,
		commonClient:     commonv1.NewCommonServiceClient(commonConn),
	}
	cleanup := func() {
		helper.Info("closing the data resources")
		if d.commonConn != nil {
			_ = d.commonConn.Close()
		}
		if d.authConn != nil {
			_ = d.authConn.Close()
		}
		if d.userConn != nil {
			_ = d.userConn.Close()
		}
		_ = d.db.Close()
	}
	return d, cleanup, nil
}

type adminRepo struct {
	data *Data
	log  *log.Helper
}

func NewAdminRepo(data *Data, logger log.Logger) biz.AdminRepo {
	repo := &adminRepo{data: data, log: log.NewHelper(logger)}
	if data != nil && data.projectionClient != nil {
		data.projectionClient.SetStatusReporter(newAdminProjectionStatusReporter(repo), "admin permission projection source")
	}
	return repo
}

type bootstrapAPIResource struct {
	ID                string
	Description       string
	Path              string
	Method            string
	Module            string
	ModuleDescription string
	ResourceGroup     string
}

type bootstrapResource struct {
	ID          string
	Name        string
	Type        string
	Value       string
	Method      string
	Description string
}

var bootstrapAPIResourceGroups = []bootstrapResource{
	{
		ID:     "a3e1fca5-e7ab-41f2-b7a9-2ab46f3640ea",
		Name:   "系统管理api组",
		Type:   "api",
		Value:  "api",
		Method: "(GET|POST|PUT|DELETE)",
	},
	{
		ID:     "7d6b49f5-3ef5-41e1-a4e5-0ccb96da5a95",
		Name:   "系统管理资源组",
		Type:   "api",
		Value:  "data",
		Method: "(GET|POST|PUT|DELETE)",
	},
	{
		ID:     "33c2ca63-4b51-43c6-b432-075047e083d7",
		Name:   "系统管理角色组",
		Type:   "api",
		Value:  "role",
		Method: "(GET|POST|PUT|DELETE)",
	},
	{
		ID:     "f1ea1c6e-b1d4-4845-b0f5-07e2ddeae705",
		Name:   "admin接口操作权限",
		Type:   "api",
		Value:  "admin",
		Method: "(GET|POST|PUT|DELETE)",
	},
	{
		ID:          "9da89181-e2d3-4b7a-9860-a12badd4b415",
		Name:        "站内信管理接口权限",
		Type:        "api",
		Value:       "site_message_manage",
		Method:      "(GET|POST|DELETE)",
		Description: "站内信管理页接口权限",
	},
	{
		ID:          "6e24d8e7-d0e4-4c41-a19d-64fc2f1cf9df",
		Name:        "组织管理接口权限",
		Type:        "api",
		Value:       "organization",
		Method:      "(GET|POST|PUT|DELETE)",
		Description: "组织管理页接口权限",
	},
	{
		ID:          "94d41386-8054-423a-9de2-af80f424b424",
		Name:        "组织权限范围接口权限",
		Type:        "api",
		Value:       "organization_permission_scope",
		Method:      "(GET|PUT)",
		Description: "组织可用权限范围配置接口权限",
	},
}

var bootstrapAdminAPIResources = []bootstrapAPIResource{
	{
		ID:                "api-admin-role-list",
		Description:       "获取角色列表",
		Path:              "/admin-api/v1/roles",
		Method:            "GET",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "role",
	},
	{
		ID:                "api-admin-role-create",
		Description:       "新增角色",
		Path:              "/admin-api/v1/roles",
		Method:            "POST",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "role",
	},
	{
		ID:                "api-admin-role-update",
		Description:       "更新角色",
		Path:              "/admin-api/v1/roles/{id}",
		Method:            "PUT",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "role",
	},
	{
		ID:                "api-admin-role-delete",
		Description:       "删除角色",
		Path:              "/admin-api/v1/roles/{id}",
		Method:            "DELETE",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "role",
	},
	{
		ID:                "api-admin-api-list",
		Description:       "获取 API 列表",
		Path:              "/admin-api/v1/apis",
		Method:            "GET",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "api",
	},
	{
		ID:                "api-admin-api-create",
		Description:       "新增 API",
		Path:              "/admin-api/v1/apis",
		Method:            "POST",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "api",
	},
	{
		ID:                "api-admin-api-update",
		Description:       "更新 API",
		Path:              "/admin-api/v1/apis/{id}",
		Method:            "PUT",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "api",
	},
	{
		ID:                "api-admin-api-delete",
		Description:       "删除 API",
		Path:              "/admin-api/v1/apis/{id}",
		Method:            "DELETE",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "api",
	},
	{
		ID:                "api-admin-resource-list",
		Description:       "获取资源列表",
		Path:              "/admin-api/v1/resources",
		Method:            "GET",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "data",
	},
	{
		ID:                "api-admin-resource-create",
		Description:       "新增资源",
		Path:              "/admin-api/v1/resources",
		Method:            "POST",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "data",
	},
	{
		ID:                "api-admin-resource-update",
		Description:       "更新资源",
		Path:              "/admin-api/v1/resources/{id}",
		Method:            "PUT",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "data",
	},
	{
		ID:                "api-admin-resource-delete",
		Description:       "删除资源",
		Path:              "/admin-api/v1/resources/{id}",
		Method:            "DELETE",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "data",
	},
	{
		ID:                "api-admin-platform-service-list",
		Description:       "获取服务注册列表",
		Path:              "/admin-api/v1/platform/services",
		Method:            "GET",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "admin",
	},
	{
		ID:                "api-admin-platform-projection-source-list",
		Description:       "获取投影源状态列表",
		Path:              "/admin-api/v1/platform/projection-sources",
		Method:            "GET",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "admin",
	},
	{
		ID:                "api-admin-organization-list",
		Description:       "获取组织列表",
		Path:              "/admin-api/v1/organizations",
		Method:            "GET",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "organization",
	},
	{
		ID:                "api-admin-organization-create",
		Description:       "新增组织",
		Path:              "/admin-api/v1/organizations",
		Method:            "POST",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "organization",
	},
	{
		ID:                "api-admin-organization-update",
		Description:       "更新组织",
		Path:              "/admin-api/v1/organizations/{id}",
		Method:            "PUT",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "organization",
	},
	{
		ID:                "api-admin-organization-delete",
		Description:       "删除组织",
		Path:              "/admin-api/v1/organizations/{id}",
		Method:            "DELETE",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "organization",
	},
	{
		ID:                "api-admin-organization-member-list",
		Description:       "获取组织成员",
		Path:              "/admin-api/v1/organizations/{organization_id}/members",
		Method:            "GET",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "organization",
	},
	{
		ID:                "api-admin-organization-member-save",
		Description:       "保存组织成员",
		Path:              "/admin-api/v1/organizations/{organization_id}/members",
		Method:            "PUT",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "organization",
	},
	{
		ID:                "api-admin-organization-permission-scope-get",
		Description:       "获取组织权限范围",
		Path:              "/admin-api/v1/organizations/{organization_id}/permission-scope",
		Method:            "GET",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "organization_permission_scope",
	},
	{
		ID:                "api-admin-organization-permission-scope-save",
		Description:       "保存组织权限范围",
		Path:              "/admin-api/v1/organizations/{organization_id}/permission-scope",
		Method:            "PUT",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "organization_permission_scope",
	},
	{
		ID:                "api-admin-permission-catalog-current",
		Description:       "获取当前组织权限目录",
		Path:              "/admin-api/v1/permission-catalog/current",
		Method:            "GET",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "role",
	},
	{
		ID:                "api-admin-organization-permission-catalog",
		Description:       "获取指定组织权限目录",
		Path:              "/admin-api/v1/organizations/{organization_id}/permission-catalog",
		Method:            "GET",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "organization_permission_scope",
	},
	{
		ID:                "api-admin-user-dept-binding-get",
		Description:       "获取用户部门绑定",
		Path:              "/admin-api/v1/user-dept-bindings/{user_id}",
		Method:            "GET",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "organization",
	},
	{
		ID:                "api-admin-user-dept-binding-upsert",
		Description:       "保存用户部门绑定",
		Path:              "/admin-api/v1/user-dept-bindings/{user_id}",
		Method:            "PUT",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "organization",
	},
	{
		ID:                "api-admin-user-dept-binding-delete",
		Description:       "删除用户部门绑定",
		Path:              "/admin-api/v1/user-dept-bindings/{user_id}",
		Method:            "DELETE",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "organization",
	},
	{
		ID:                "api-admin-my-organization-list",
		Description:       "获取我的组织",
		Path:              "/admin-api/v1/my/organizations",
		Method:            "GET",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "default",
	},
	{
		ID:                "api-admin-my-current-organization-switch",
		Description:       "切换当前组织",
		Path:              "/admin-api/v1/my/current-organization",
		Method:            "PUT",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "default",
	},
	{
		ID:                "api-admin-walk-route",
		Description:       "获取系统所有api接口",
		Path:              "/admin-api/v1/walk-routes",
		Method:            "GET",
		Module:            "admin",
		ModuleDescription: "系统管理",
		ResourceGroup:     "api",
	},
	{
		ID:                "api-common-site-message-my-list",
		Description:       "获取我的站内信列表",
		Path:              "/common-api/v1/site-messages/my",
		Method:            "GET",
		Module:            "common",
		ModuleDescription: "公共服务",
		ResourceGroup:     "default",
	},
	{
		ID:                "api-common-site-message-my-unread-count",
		Description:       "获取我的站内信未读数量",
		Path:              "/common-api/v1/site-messages/my/unread-count",
		Method:            "GET",
		Module:            "common",
		ModuleDescription: "公共服务",
		ResourceGroup:     "default",
	},
	{
		ID:                "api-common-site-message-my-read",
		Description:       "标记站内信已读",
		Path:              "/common-api/v1/site-messages/my/{message_id}/read",
		Method:            "POST",
		Module:            "common",
		ModuleDescription: "公共服务",
		ResourceGroup:     "default",
	},
	{
		ID:                "api-common-site-message-my-unread",
		Description:       "标记站内信未读",
		Path:              "/common-api/v1/site-messages/my/{message_id}/unread",
		Method:            "POST",
		Module:            "common",
		ModuleDescription: "公共服务",
		ResourceGroup:     "default",
	},
	{
		ID:                "api-common-site-message-my-read-all",
		Description:       "全部站内信标记已读",
		Path:              "/common-api/v1/site-messages/my/read-all",
		Method:            "POST",
		Module:            "common",
		ModuleDescription: "公共服务",
		ResourceGroup:     "default",
	},
	{
		ID:                "api-common-site-message-manage-list",
		Description:       "获取站内信发布记录",
		Path:              "/common-api/v1/site-messages/manage",
		Method:            "GET",
		Module:            "common",
		ModuleDescription: "公共服务",
		ResourceGroup:     "site_message_manage",
	},
	{
		ID:                "api-common-site-message-manage-create",
		Description:       "保存或发布站内信",
		Path:              "/common-api/v1/site-messages/manage",
		Method:            "POST",
		Module:            "common",
		ModuleDescription: "公共服务",
		ResourceGroup:     "site_message_manage",
	},
	{
		ID:                "api-common-site-message-manage-recall",
		Description:       "撤回站内信",
		Path:              "/common-api/v1/site-messages/manage/{id}/recall",
		Method:            "POST",
		Module:            "common",
		ModuleDescription: "公共服务",
		ResourceGroup:     "site_message_manage",
	},
	{
		ID:                "api-common-site-message-manage-delete",
		Description:       "删除未发布站内信",
		Path:              "/common-api/v1/site-messages/manage/{id}",
		Method:            "DELETE",
		Module:            "common",
		ModuleDescription: "公共服务",
		ResourceGroup:     "site_message_manage",
	},
}

func (r *adminRepo) ListAllRoles(ctx context.Context) ([]*ent.Role, error) {
	return r.data.db.Role.Query().
		Where(role.StatusEQ(true)).
		WithResource().
		All(ctx)
}

func (r *adminRepo) ListRoles(ctx context.Context, req *v1.RolePageParams) ([]*ent.Role, error) {
	query := r.data.db.Role.Query()
	if req.Name != "" {
		query = query.Where(role.NameEQ(req.Name))
	}
	if req.Status == 1 {
		query = query.Where(role.StatusEQ(true))
	}
	if organizationID := strings.TrimSpace(req.OrganizationId); organizationID != "" {
		query = query.Where(role.OrganizationIDEQ(organizationID))
	}
	query.WithResource(func(query *ent.ResourceQuery) {
		query.Select(resource.FieldID, resource.FieldType, resource.FieldValue, resource.FieldMethod)
	})
	return query.All(ctx)
}

func (r *adminRepo) GetRole(ctx context.Context, id int64) (*ent.Role, error) {
	return r.data.db.Role.Query().Where(role.IDEQ(id)).First(ctx)
}

func (r *adminRepo) ResolveRoleValues(ctx context.Context, roleIDs []int64) (map[int64]string, error) {
	roles, err := r.data.db.Role.Query().
		Where(role.IDIn(roleIDs...)).
		Select(role.FieldID, role.FieldValue).
		All(ctx)
	if err != nil {
		return nil, err
	}
	values := make(map[int64]string, len(roles))
	for _, item := range roles {
		values[item.ID] = item.Value
	}
	return values, nil
}

func (r *adminRepo) getDefaultUserRole(ctx context.Context) (*ent.Role, error) {
	return r.data.db.Role.Query().
		Where(
			role.OrganizationIDEQ(defaultOrganizationID),
			role.ValueEQ("default"),
			role.StatusEQ(true),
		).
		First(ctx)
}

func (r *adminRepo) GetUserRoleBindings(ctx context.Context, userID string, organizationID string) ([]*ent.UserRoleBinding, error) {
	items, err := r.data.db.UserRoleBinding.Query().
		Where(
			userrolebinding.UserIDEQ(userID),
			userrolebinding.OrganizationIDEQ(strings.TrimSpace(organizationID)),
		).
		Order(userrolebinding.ByRoleID()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return r.withDefaultRoleFallbackForUser(ctx, userID, organizationID, items)
}

func (r *adminRepo) ListUserOrganizationRoles(ctx context.Context, userID string, organizationID string) ([]*ent.Role, error) {
	bindings, err := r.GetUserRoleBindings(ctx, strings.TrimSpace(userID), strings.TrimSpace(organizationID))
	if err != nil {
		return nil, err
	}
	roleIDs := make([]int64, 0, len(bindings))
	seen := make(map[int64]struct{}, len(bindings))
	for _, item := range bindings {
		if item == nil || item.RoleID <= 0 {
			continue
		}
		if _, ok := seen[item.RoleID]; ok {
			continue
		}
		seen[item.RoleID] = struct{}{}
		roleIDs = append(roleIDs, item.RoleID)
	}
	if len(roleIDs) == 0 {
		return []*ent.Role{}, nil
	}
	return r.data.db.Role.Query().
		Where(role.IDIn(roleIDs...)).
		WithResource(func(query *ent.ResourceQuery) {
			query.Select(resource.FieldID, resource.FieldType, resource.FieldValue, resource.FieldMethod)
		}).
		All(ctx)
}

func (r *adminRepo) ListUserRoleBindings(ctx context.Context) ([]*ent.UserRoleBinding, error) {
	items, err := r.listExplicitUserRoleBindings(ctx)
	if err != nil {
		return nil, err
	}
	return r.withDefaultRoleFallbackForAllUsers(ctx, items)
}

func (r *adminRepo) listExplicitUserRoleBindings(ctx context.Context) ([]*ent.UserRoleBinding, error) {
	return r.data.db.UserRoleBinding.Query().
		Order(userrolebinding.ByOrganizationID(), userrolebinding.ByUserID(), userrolebinding.ByRoleID()).
		All(ctx)
}

func (r *adminRepo) UpsertUserRoleBinding(ctx context.Context, userID string, organizationID string, roleIDs []int64) ([]*ent.UserRoleBinding, error) {
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	query := tx.UserRoleBinding.Delete().Where(
		userrolebinding.UserIDEQ(userID),
		userrolebinding.OrganizationIDEQ(strings.TrimSpace(organizationID)),
	)
	if len(roleIDs) > 0 {
		query = query.Where(userrolebinding.RoleIDNotIn(roleIDs...))
	}
	if _, err := query.Exec(ctx); err != nil {
		return nil, err
	}
	existing, err := tx.UserRoleBinding.Query().
		Where(
			userrolebinding.UserIDEQ(userID),
			userrolebinding.OrganizationIDEQ(strings.TrimSpace(organizationID)),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	existingRoleIDs := make(map[int64]struct{}, len(existing))
	for _, item := range existing {
		existingRoleIDs[item.RoleID] = struct{}{}
	}
	for _, roleID := range roleIDs {
		if _, ok := existingRoleIDs[roleID]; ok {
			continue
		}
		if _, err := tx.UserRoleBinding.Create().
			SetUserID(userID).
			SetOrganizationID(strings.TrimSpace(organizationID)).
			SetRoleID(roleID).
			Save(ctx); err != nil {
			return nil, err
		}
	}
	items, err := tx.UserRoleBinding.Query().
		Where(
			userrolebinding.UserIDEQ(userID),
			userrolebinding.OrganizationIDEQ(strings.TrimSpace(organizationID)),
		).
		Order(userrolebinding.ByRoleID()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true
	return items, nil
}

func (r *adminRepo) DeleteUserRoleBinding(ctx context.Context, userID string, organizationID string) error {
	_, err := r.data.db.UserRoleBinding.Delete().
		Where(
			userrolebinding.UserIDEQ(userID),
			userrolebinding.OrganizationIDEQ(strings.TrimSpace(organizationID)),
		).
		Exec(ctx)
	return err
}

func (r *adminRepo) GetUserDeptBinding(ctx context.Context, userID string, organizationID string) (*ent.UserDeptMembership, error) {
	return r.data.db.UserDeptMembership.Query().
		Where(
			userdeptmembership.UserIDEQ(strings.TrimSpace(userID)),
			userdeptmembership.OrganizationIDEQ(strings.TrimSpace(organizationID)),
		).
		Only(ctx)
}

func (r *adminRepo) UpsertUserDeptBinding(ctx context.Context, userID string, organizationID string, deptID int64) (*ent.UserDeptMembership, error) {
	userID = strings.TrimSpace(userID)
	organizationID = strings.TrimSpace(organizationID)
	existing, err := r.data.db.UserDeptMembership.Query().
		Where(
			userdeptmembership.UserIDEQ(userID),
			userdeptmembership.OrganizationIDEQ(organizationID),
		).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}
	if existing != nil {
		return r.data.db.UserDeptMembership.UpdateOne(existing).
			SetDeptID(deptID).
			Save(ctx)
	}
	return r.data.db.UserDeptMembership.Create().
		SetUserID(userID).
		SetOrganizationID(organizationID).
		SetDeptID(deptID).
		Save(ctx)
}

func (r *adminRepo) DeleteUserDeptBinding(ctx context.Context, userID string, organizationID string) error {
	_, err := r.data.db.UserDeptMembership.Delete().
		Where(
			userdeptmembership.UserIDEQ(strings.TrimSpace(userID)),
			userdeptmembership.OrganizationIDEQ(strings.TrimSpace(organizationID)),
		).
		Exec(ctx)
	return err
}

func (r *adminRepo) RemoveOrganizationRoleBindings(ctx context.Context, userID string, organizationID string) error {
	_, err := r.data.db.UserRoleBinding.Delete().
		Where(
			userrolebinding.UserIDEQ(strings.TrimSpace(userID)),
			userrolebinding.OrganizationIDEQ(strings.TrimSpace(organizationID)),
		).
		Exec(ctx)
	return err
}

func (r *adminRepo) RemoveOrganizationDeptBinding(ctx context.Context, userID string, organizationID string) error {
	_, err := r.data.db.UserDeptMembership.Delete().
		Where(
			userdeptmembership.UserIDEQ(strings.TrimSpace(userID)),
			userdeptmembership.OrganizationIDEQ(strings.TrimSpace(organizationID)),
		).
		Exec(ctx)
	return err
}

func (r *adminRepo) ListOrganizationPermissionScopes(ctx context.Context, organizationID string) ([]*ent.OrganizationPermissionScope, error) {
	return r.data.db.OrganizationPermissionScope.Query().
		Where(organizationpermissionscope.OrganizationIDEQ(strings.TrimSpace(organizationID))).
		Order(
			organizationpermissionscope.ByPermissionType(),
			organizationpermissionscope.ByPermissionRef(),
		).
		All(ctx)
}

func (r *adminRepo) SaveOrganizationPermissionScopes(ctx context.Context, organizationID string, menuIDs []int32, resourceIDs []string, createdBy string) ([]*ent.OrganizationPermissionScope, error) {
	organizationID = strings.TrimSpace(organizationID)
	createdBy = strings.TrimSpace(createdBy)
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.OrganizationPermissionScope.Delete().
		Where(organizationpermissionscope.OrganizationIDEQ(organizationID)).
		Exec(ctx); err != nil {
		return nil, err
	}
	builders := make([]*ent.OrganizationPermissionScopeCreate, 0, len(menuIDs)+len(resourceIDs))
	for _, menuID := range menuIDs {
		if menuID <= 0 {
			continue
		}
		builders = append(builders, tx.OrganizationPermissionScope.Create().
			SetOrganizationID(organizationID).
			SetPermissionType("menu").
			SetPermissionRef(strconv.FormatInt(int64(menuID), 10)).
			SetCreatedBy(createdBy))
	}
	for _, resourceID := range resourceIDs {
		resourceID = strings.TrimSpace(resourceID)
		if resourceID == "" {
			continue
		}
		builders = append(builders, tx.OrganizationPermissionScope.Create().
			SetOrganizationID(organizationID).
			SetPermissionType("resource").
			SetPermissionRef(resourceID).
			SetCreatedBy(createdBy))
	}
	if len(builders) > 0 {
		if _, err = tx.OrganizationPermissionScope.CreateBulk(builders...).Save(ctx); err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return r.ListOrganizationPermissionScopes(ctx, organizationID)
}

func (r *adminRepo) withDefaultRoleFallbackForUser(ctx context.Context, userID string, organizationID string, items []*ent.UserRoleBinding) ([]*ent.UserRoleBinding, error) {
	if len(items) > 0 || strings.TrimSpace(userID) == "" {
		return items, nil
	}
	defaultOrganizationID, err := r.GetDefaultOrganizationID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(organizationID) != defaultOrganizationID {
		return items, nil
	}
	defaultRole, err := r.getDefaultUserRole(ctx)
	if ent.IsNotFound(err) {
		return items, nil
	}
	if err != nil {
		return nil, err
	}
	return []*ent.UserRoleBinding{syntheticUserRoleBinding(userID, defaultOrganizationID, defaultRole.ID)}, nil
}

func (r *adminRepo) withDefaultRoleFallbackForAllUsers(ctx context.Context, items []*ent.UserRoleBinding) ([]*ent.UserRoleBinding, error) {
	defaultOrganizationID, err := r.GetDefaultOrganizationID(ctx)
	if err != nil {
		return nil, err
	}
	defaultRole, err := r.getDefaultUserRole(ctx)
	if ent.IsNotFound(err) {
		return items, nil
	}
	if err != nil {
		return nil, err
	}
	userIDs, err := r.listAllUserIDs(ctx)
	if err != nil {
		return nil, err
	}
	if len(userIDs) == 0 {
		return items, nil
	}
	next := make([]*ent.UserRoleBinding, 0, len(items)+len(userIDs))
	next = append(next, items...)
	boundUserIDs := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item == nil || strings.TrimSpace(item.UserID) == "" {
			continue
		}
		boundUserIDs[item.UserID+"|"+strings.TrimSpace(item.OrganizationID)] = struct{}{}
	}
	for _, userID := range userIDs {
		key := userID + "|" + defaultOrganizationID
		if _, ok := boundUserIDs[key]; ok {
			continue
		}
		next = append(next, syntheticUserRoleBinding(userID, defaultOrganizationID, defaultRole.ID))
	}
	sort.SliceStable(next, func(i, j int) bool {
		left := next[i]
		right := next[j]
		if left == nil {
			return false
		}
		if right == nil {
			return true
		}
		if left.OrganizationID != right.OrganizationID {
			return left.OrganizationID < right.OrganizationID
		}
		if left.UserID != right.UserID {
			return left.UserID < right.UserID
		}
		return left.RoleID < right.RoleID
	})
	return next, nil
}

func syntheticUserRoleBinding(userID string, organizationID string, roleID int64) *ent.UserRoleBinding {
	return &ent.UserRoleBinding{
		OrganizationID: strings.TrimSpace(organizationID),
		UserID:         strings.TrimSpace(userID),
		RoleID:         roleID,
	}
}

func (r *adminRepo) ListOrganizations(ctx context.Context, req *v1.GetOrganizationListParams) ([]*ent.Organization, int64, error) {
	query := r.data.db.Organization.Query()
	if req.Name != "" {
		query = query.Where(organization.NameContains(req.Name))
	}
	if req.Code != "" {
		query = query.Where(organization.CodeContains(req.Code))
	}
	if req.Status == 1 {
		query = query.Where(organization.StatusEQ(true))
	} else if req.Status == 2 {
		query = query.Where(organization.StatusEQ(false))
	}
	count, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	query = query.Order(organization.BySort(), organization.ByID())
	if req.PageSize > 0 {
		query = query.Limit(int(req.PageSize))
		if req.CurrentPage > 0 {
			query = query.Offset(int(tools.GetPageOffset(req.CurrentPage, req.PageSize)))
		}
	}
	items, err := query.All(ctx)
	return items, int64(count), err
}

func (r *adminRepo) GetOrganization(ctx context.Context, id string) (*ent.Organization, error) {
	return r.data.db.Organization.Get(ctx, id)
}

func (r *adminRepo) AddOrganization(ctx context.Context, req *v1.OrganizationItem) (*ent.Organization, error) {
	return r.data.db.Organization.Create().
		SetName(req.Name).
		SetCode(req.Code).
		SetSort(req.OrderNo).
		SetStatus(req.Status != 0).
		SetDesc(req.Remark).
		SetExtension("").
		Save(ctx)
}

func (r *adminRepo) UpdateOrganization(ctx context.Context, id string, req *v1.OrganizationItem) (*ent.Organization, error) {
	return r.data.db.Organization.UpdateOneID(id).
		SetName(req.Name).
		SetCode(req.Code).
		SetSort(req.OrderNo).
		SetStatus(req.Status != 0).
		SetDesc(req.Remark).
		SetExtension("").
		Save(ctx)
}

func (r *adminRepo) DelOrganization(ctx context.Context, id string) error {
	return r.data.db.Organization.DeleteOneID(id).Exec(ctx)
}

func (r *adminRepo) GetDefaultOrganizationID(ctx context.Context) (string, error) {
	item, err := r.data.db.Organization.Query().
		Where(organization.CodeEQ("default")).
		First(ctx)
	if err != nil {
		return "", err
	}
	return item.ID, nil
}

func (r *adminRepo) CountOrganizationDepts(ctx context.Context, organizationID string) (int64, error) {
	count, err := r.data.db.Dept.Query().
		Where(dept.OrganizationIDEQ(organizationID)).
		Count(ctx)
	return int64(count), err
}

func (r *adminRepo) CountOrganizationMembers(ctx context.Context, organizationID string) (int64, error) {
	count, err := r.data.db.UserOrganization.Query().
		Where(userorganization.OrganizationIDEQ(organizationID)).
		Count(ctx)
	return int64(count), err
}

func (r *adminRepo) ListOrganizationMembers(ctx context.Context, organizationID string) ([]*v1.OrganizationMemberItem, error) {
	members, err := r.data.db.UserOrganization.Query().
		Where(userorganization.OrganizationIDEQ(organizationID)).
		Order(userorganization.ByUserID()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return r.organizationMembersToReply(ctx, members)
}

func (r *adminRepo) SaveOrganizationMembers(ctx context.Context, organizationID string, userIDs []string) ([]*v1.OrganizationMemberItem, error) {
	normalizedUserIDs := normalizeUserIDs(userIDs)
	defaultOrganizationID, err := r.GetDefaultOrganizationID(ctx)
	if err != nil {
		return nil, err
	}
	if organizationID == defaultOrganizationID {
		normalizedUserIDs, err = r.listAllUserIDs(ctx)
		if err != nil {
			return nil, err
		}
	}
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	query := tx.UserOrganization.Delete().Where(userorganization.OrganizationIDEQ(organizationID))
	if len(normalizedUserIDs) > 0 {
		query = query.Where(userorganization.UserIDNotIn(normalizedUserIDs...))
	}
	if _, err := query.Exec(ctx); err != nil {
		return nil, err
	}
	existing, err := tx.UserOrganization.Query().
		Where(userorganization.OrganizationIDEQ(organizationID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	existingUserIDs := make(map[string]struct{}, len(existing))
	for _, item := range existing {
		existingUserIDs[item.UserID] = struct{}{}
	}
	for _, userID := range normalizedUserIDs {
		if _, ok := existingUserIDs[userID]; ok {
			if _, err := tx.UserOrganization.Update().
				Where(
					userorganization.OrganizationIDEQ(organizationID),
					userorganization.UserIDEQ(userID),
				).
				SetStatus(true).
				Save(ctx); err != nil {
				return nil, err
			}
			continue
		}
		if _, err := tx.UserOrganization.Create().
			SetOrganizationID(organizationID).
			SetUserID(userID).
			SetIsPrimary(false).
			SetStatus(true).
			Save(ctx); err != nil {
			return nil, err
		}
	}
	items, err := tx.UserOrganization.Query().
		Where(userorganization.OrganizationIDEQ(organizationID)).
		Order(userorganization.ByUserID()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true
	if organizationID != defaultOrganizationID {
		normalizedUserIDSet := make(map[string]struct{}, len(normalizedUserIDs))
		for _, userID := range normalizedUserIDs {
			normalizedUserIDSet[userID] = struct{}{}
		}
		existingMemberUserIDs := make(map[string]struct{}, len(existing))
		for _, item := range existing {
			if item == nil {
				continue
			}
			existingMemberUserIDs[item.UserID] = struct{}{}
		}
		for userID := range existingMemberUserIDs {
			if _, ok := normalizedUserIDSet[userID]; ok {
				continue
			}
			currentOrganizationID, currentErr := r.GetCurrentOrganizationID(ctx, userID)
			if currentErr != nil {
				return nil, currentErr
			}
			if currentOrganizationID != organizationID {
				continue
			}
			if switchErr := r.SwitchCurrentOrganization(ctx, userID, defaultOrganizationID); switchErr != nil {
				return nil, switchErr
			}
		}
	}
	return r.organizationMembersToReply(ctx, items)
}

func (r *adminRepo) EnsureAllUsersInDefaultOrganization(ctx context.Context) error {
	defaultOrganizationID, err := r.GetDefaultOrganizationID(ctx)
	if err != nil {
		return err
	}
	userIDs, err := r.listAllUserIDs(ctx)
	if err != nil {
		return err
	}
	existing, err := r.data.db.UserOrganization.Query().
		Where(userorganization.OrganizationIDEQ(defaultOrganizationID)).
		All(ctx)
	if err != nil {
		return err
	}
	existingByUserID := make(map[string]*ent.UserOrganization, len(existing))
	for _, item := range existing {
		existingByUserID[item.UserID] = item
	}
	for _, userID := range userIDs {
		if item, ok := existingByUserID[userID]; ok {
			if _, err := r.data.db.UserOrganization.UpdateOneID(item.ID).
				SetStatus(true).
				Save(ctx); err != nil {
				return err
			}
			continue
		}
		hasPrimary, err := r.userHasCurrentOrganization(ctx, userID)
		if err != nil {
			return err
		}
		if _, err := r.data.db.UserOrganization.Create().
			SetUserID(userID).
			SetOrganizationID(defaultOrganizationID).
			SetIsPrimary(!hasPrimary).
			SetStatus(true).
			Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *adminRepo) EnsureUserInDefaultOrganization(ctx context.Context, userID string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return fmt.Errorf("user id is required")
	}
	defaultOrganizationID, err := r.GetDefaultOrganizationID(ctx)
	if err != nil {
		return err
	}
	hasPrimary, err := r.userHasCurrentOrganization(ctx, userID)
	if err != nil {
		return err
	}
	item, err := r.data.db.UserOrganization.Query().
		Where(
			userorganization.UserIDEQ(userID),
			userorganization.OrganizationIDEQ(defaultOrganizationID),
		).
		First(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return err
		}
		_, err = r.data.db.UserOrganization.Create().
			SetUserID(userID).
			SetOrganizationID(defaultOrganizationID).
			SetIsPrimary(!hasPrimary).
			SetStatus(true).
			Save(ctx)
		return err
	}
	cmd := r.data.db.UserOrganization.UpdateOneID(item.ID).SetStatus(true)
	if !hasPrimary {
		cmd = cmd.SetIsPrimary(true)
	}
	_, err = cmd.Save(ctx)
	return err
}

func (r *adminRepo) ListUserOrganizations(ctx context.Context, userID string) ([]*ent.UserOrganization, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}
	return r.data.db.UserOrganization.Query().
		Where(
			userorganization.UserIDEQ(userID),
			userorganization.StatusEQ(true),
		).
		Order(
			userorganization.ByIsPrimary(sql.OrderDesc()),
			userorganization.ByCreateTime(),
			userorganization.ByOrganizationID(),
		).
		All(ctx)
}

func (r *adminRepo) GetCurrentOrganizationID(ctx context.Context, userID string) (string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", fmt.Errorf("user id is required")
	}
	if err := r.EnsureUserInDefaultOrganization(ctx, userID); err != nil {
		return "", err
	}
	item, err := r.data.db.UserOrganization.Query().
		Where(
			userorganization.UserIDEQ(userID),
			userorganization.StatusEQ(true),
			userorganization.IsPrimaryEQ(true),
		).
		Order(userorganization.ByID()).
		First(ctx)
	if err == nil {
		organizationItem, orgErr := r.GetOrganization(ctx, item.OrganizationID)
		if orgErr == nil && organizationItem.Status {
			return item.OrganizationID, nil
		}
		if orgErr != nil && !ent.IsNotFound(orgErr) {
			return "", orgErr
		}
		defaultOrganizationID, err := r.GetDefaultOrganizationID(ctx)
		if err != nil {
			return "", err
		}
		if err := r.SwitchCurrentOrganization(ctx, userID, defaultOrganizationID); err != nil {
			return "", err
		}
		return defaultOrganizationID, nil
	}
	if !ent.IsNotFound(err) {
		return "", err
	}
	defaultOrganizationID, err := r.GetDefaultOrganizationID(ctx)
	if err != nil {
		return "", err
	}
	if err := r.SwitchCurrentOrganization(ctx, userID, defaultOrganizationID); err != nil {
		return "", err
	}
	return defaultOrganizationID, nil
}

func (r *adminRepo) UserBelongsToOrganization(ctx context.Context, userID string, organizationID string) (bool, error) {
	userID = strings.TrimSpace(userID)
	organizationID = strings.TrimSpace(organizationID)
	if userID == "" || organizationID == "" {
		return false, nil
	}
	return r.data.db.UserOrganization.Query().
		Where(
			userorganization.UserIDEQ(userID),
			userorganization.OrganizationIDEQ(organizationID),
			userorganization.StatusEQ(true),
		).
		Exist(ctx)
}

func (r *adminRepo) SwitchCurrentOrganization(ctx context.Context, userID string, organizationID string) error {
	userID = strings.TrimSpace(userID)
	organizationID = strings.TrimSpace(organizationID)
	if userID == "" {
		return fmt.Errorf("user id is required")
	}
	if organizationID == "" {
		return fmt.Errorf("organization id is required")
	}
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if _, err := tx.UserOrganization.Update().
		Where(userorganization.UserIDEQ(userID)).
		SetIsPrimary(false).
		Save(ctx); err != nil {
		return err
	}
	affected, err := tx.UserOrganization.Update().
		Where(
			userorganization.UserIDEQ(userID),
			userorganization.OrganizationIDEQ(organizationID),
			userorganization.StatusEQ(true),
		).
		SetIsPrimary(true).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("user is not a member of organization")
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func (r *adminRepo) userHasCurrentOrganization(ctx context.Context, userID string) (bool, error) {
	return r.data.db.UserOrganization.Query().
		Where(
			userorganization.UserIDEQ(userID),
			userorganization.StatusEQ(true),
			userorganization.IsPrimaryEQ(true),
		).
		Exist(ctx)
}

func (r *adminRepo) listAllUserIDs(ctx context.Context) ([]string, error) {
	const pageSize int64 = 1000
	var (
		currentPage int64 = 1
		userIDs           = make([]string, 0)
	)
	for {
		res, err := r.getUserList(ctx, &userv1.GetUserParams{
			CurrentPage: currentPage,
			PageSize:    pageSize,
		})
		if err != nil {
			return nil, err
		}
		for _, item := range res.Items {
			if item != nil && strings.TrimSpace(item.Id) != "" {
				userIDs = append(userIDs, item.Id)
			}
		}
		if int64(len(userIDs)) >= res.Total || len(res.Items) == 0 {
			return normalizeUserIDs(userIDs), nil
		}
		currentPage++
	}
}

func (r *adminRepo) organizationMembersToReply(ctx context.Context, members []*ent.UserOrganization) ([]*v1.OrganizationMemberItem, error) {
	if len(members) == 0 {
		return []*v1.OrganizationMemberItem{}, nil
	}
	res, err := r.getUserList(ctx, &userv1.GetUserParams{CurrentPage: 1, PageSize: 10000})
	if err != nil {
		return nil, err
	}
	userMap := make(map[string]*userv1.UserListItem, len(res.Items))
	for _, item := range res.Items {
		if item != nil {
			userMap[item.Id] = item
		}
	}
	items := make([]*v1.OrganizationMemberItem, 0, len(members))
	for _, member := range members {
		userItem := userMap[member.UserID]
		memberStatus := int32(0)
		if member.Status {
			memberStatus = 1
		}
		reply := &v1.OrganizationMemberItem{
			UserId:       member.UserID,
			MemberStatus: memberStatus,
			Primary:      member.IsPrimary,
			CreateTime:   member.CreateTime.Format(time.DateTime),
		}
		if userItem != nil {
			reply.Username = userItem.Username
			reply.Nickname = userItem.Nickname
			reply.Email = userItem.Email
			reply.Avatar = userItem.Avatar
			reply.UserStatus = userItem.Status
		}
		items = append(items, reply)
	}
	return items, nil
}

func (r *adminRepo) getUserList(ctx context.Context, req *userv1.GetUserParams) (*userv1.GetUserListReply, error) {
	ctx = authx.ForwardAuthorizationContext(ctx)
	return r.data.userClient.GetUserList(ctx, req)
}

func (r *adminRepo) AddRole(ctx context.Context, req *v1.RoleListItem) (*ent.Role, error) {
	cmd := r.data.db.Role.Create().
		SetName(req.Name).
		SetValue(req.Value).
		SetOrganizationID(defaultOrganizationID).
		SetStatus(req.Status != 0).
		SetDesc(req.Remark).
		SetMenus(req.Permissions).
		SetDataScope(req.DataScope).
		SetDataScopeDeptIds(req.DataScopeDeptIds).
		AddResourceIDs(req.ApiPermissions...)
	if organizationID := strings.TrimSpace(req.OrganizationId); organizationID != "" {
		cmd = cmd.SetOrganizationID(organizationID)
	}
	return cmd.Save(ctx)
}

func (r *adminRepo) UpdateRole(ctx context.Context, roleID int64, req *v1.RoleListItem) (*ent.Role, error) {
	cmd := r.data.db.Role.UpdateOneID(roleID).
		SetName(req.Name).
		SetValue(req.Value).
		SetStatus(req.Status != 0).
		SetDesc(req.Remark).
		SetMenus(req.Permissions).
		SetDataScope(req.DataScope).
		SetDataScopeDeptIds(req.DataScopeDeptIds).
		ClearResource().
		AddResourceIDs(req.ApiPermissions...)
	if organizationID := strings.TrimSpace(req.OrganizationId); organizationID != "" {
		cmd = cmd.SetOrganizationID(organizationID)
	}
	return cmd.Save(ctx)
}

func (r *adminRepo) DelRole(ctx context.Context, id int64) error {
	return r.data.db.Role.DeleteOneID(id).Exec(ctx)
}

func (r *adminRepo) GetMenuList(ctx context.Context) ([]*ent.Menu, error) {
	return r.data.db.Menu.Query().Order(menu.ByPid(), menu.ByOrder()).All(ctx)
}

func (r *adminRepo) GetUserAuthInfo(ctx context.Context, userID string) (*userv1.GetUserAuthInfoReply, error) {
	ctx = authx.ForwardAuthorizationContext(ctx)
	return r.data.userClient.GetUserAuthInfo(ctx, &userv1.GetUserAuthInfoRequest{UserId: userID})
}

func getAPIListQuery(params *v1.GetApiPageParams, isPage bool) func(s *sql.Selector) {
	return func(s *sql.Selector) {
		if params.Path != "" {
			s.Where(sql.EQ(apiresources.FieldPath, params.Path))
		}
		if params.ResourcesGroup != "" {
			s.Where(sql.EQ(apiresources.FieldResourcesGroup, params.ResourcesGroup))
		}
		if params.Method != "" {
			s.Where(sql.EQ(apiresources.FieldMethod, params.Method))
		}
		if params.Description != "" {
			s.Where(sql.Like(apiresources.FieldDescription, "%"+params.Description+"%"))
		}
		if isPage {
			if params.PageSize != 0 {
				s.Limit(int(params.PageSize))
			}
			if params.CurrentPage != 0 {
				s.Offset(int(tools.GetPageOffset(params.CurrentPage, params.PageSize)))
			}
		}
	}
}

func (r *adminRepo) GetApiList(ctx context.Context, req *v1.GetApiPageParams) ([]*ent.ApiResources, int64, error) {
	list, err := r.data.db.ApiResources.Query().Modify(getAPIListQuery(req, true)).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	count, err := r.data.db.ApiResources.Query().Modify(getAPIListQuery(req, false)).Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	return list, int64(count), nil
}

func (r *adminRepo) GetApi(ctx context.Context, id string) (*ent.ApiResources, error) {
	return r.data.db.ApiResources.Query().Where(apiresources.IDEQ(id)).First(ctx)
}

func (r *adminRepo) AddApi(ctx context.Context, req *ent.ApiResources) (*ent.ApiResources, error) {
	return r.data.db.ApiResources.Create().CreateAll(req).Save(ctx)
}

func (r *adminRepo) UpdateApi(ctx context.Context, req *ent.ApiResources) (*ent.ApiResources, error) {
	return r.data.db.ApiResources.UpdateOneID(req.ID).UpdateAll(req).Save(ctx)
}

func (r *adminRepo) DelApi(ctx context.Context, id string) error {
	return r.data.db.ApiResources.DeleteOneID(id).Exec(ctx)
}

func getResourceListQuery(params *v1.GetResourcePageParams, isPage bool) func(s *sql.Selector) {
	return func(s *sql.Selector) {
		if params.Name != "" {
			s.Where(sql.Like(resource.FieldName, "%"+params.Name+"%"))
		}
		if params.Type != "" {
			s.Where(sql.EQ(resource.FieldType, params.Type))
		}
		if params.Value != "" {
			s.Where(sql.EQ(resource.FieldValue, params.Value))
		}
		if params.Method != "" {
			s.Where(sql.Like(resource.FieldMethod, "%"+params.Method+"%"))
		}
		if params.Description != "" {
			s.Where(sql.Like(resource.FieldDescription, "%"+params.Description+"%"))
		}
		if isPage {
			if params.PageSize != 0 {
				s.Limit(int(params.PageSize))
			}
			if params.CurrentPage != 0 {
				s.Offset(int(tools.GetPageOffset(params.CurrentPage, params.PageSize)))
			}
		}
	}
}

func (r *adminRepo) GetResourceList(ctx context.Context, req *v1.GetResourcePageParams) ([]*ent.Resource, int64, error) {
	list, err := r.data.db.Resource.Query().Modify(getResourceListQuery(req, true)).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	count, err := r.data.db.Resource.Query().Modify(getResourceListQuery(req, false)).Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	return list, int64(count), nil
}

func (r *adminRepo) AddResource(ctx context.Context, req *ent.Resource) (*ent.Resource, error) {
	return r.data.db.Resource.Create().CreateAll(req).Save(ctx)
}

func (r *adminRepo) GetResource(ctx context.Context, id string) (*ent.Resource, error) {
	return r.data.db.Resource.Query().Where(resource.IDEQ(id)).First(ctx)
}

func (r *adminRepo) UpdateResource(ctx context.Context, req *ent.Resource) (*ent.Resource, error) {
	return r.data.db.Resource.UpdateOneID(req.ID).UpdateAll(req).Save(ctx)
}

func (r *adminRepo) DelResource(ctx context.Context, id string) error {
	return r.data.db.Resource.DeleteOneID(id).Exec(ctx)
}

func (r *adminRepo) EnsurePermissionBootstrap(ctx context.Context) error {
	if r == nil || r.data == nil || r.data.sqlDB == nil {
		return nil
	}
	if err := r.ensureBootstrapResourceGroups(ctx); err != nil {
		return err
	}
	if err := r.ensureBootstrapAPIResources(ctx); err != nil {
		return err
	}
	if err := r.ensureBootstrapResourceRoles(ctx); err != nil {
		return err
	}
	if err := r.ensureBootstrapAPIRoles(ctx); err != nil {
		return err
	}
	return r.ensureDefaultOrganizationPermissionScope(ctx)
}

func (r *adminRepo) ensureBootstrapResourceGroups(ctx context.Context) error {
	const query = `
INSERT INTO sys_resources (
  id, create_time, update_time, name, type, value, method, description
) VALUES (
  $1, NOW(), NOW(), $2, $3, $4, $5, $6
)
ON CONFLICT (id) DO UPDATE SET
  update_time = NOW(),
  name = EXCLUDED.name,
  type = EXCLUDED.type,
  value = EXCLUDED.value,
  method = EXCLUDED.method,
  description = EXCLUDED.description`
	for _, item := range bootstrapAPIResourceGroups {
		if _, err := r.data.sqlDB.ExecContext(ctx, query, item.ID, item.Name, item.Type, item.Value, item.Method, item.Description); err != nil {
			return err
		}
	}
	return nil
}

func (r *adminRepo) ensureBootstrapAPIResources(ctx context.Context) error {
	const query = `
	INSERT INTO sys_api_resources (
	  id, create_time, update_time, description, path, method, module, module_description, resources_group
	) VALUES (
	  $1, NOW(), NOW(), $2, $3, $4, $5, $6, $7
	)
	ON CONFLICT (path, method) DO UPDATE SET
	  update_time = NOW(),
	  description = EXCLUDED.description,
	  module = EXCLUDED.module,
	  module_description = EXCLUDED.module_description,
	  resources_group = EXCLUDED.resources_group`
	for _, item := range bootstrapAdminAPIResources {
		if _, err := r.data.sqlDB.ExecContext(ctx, query, item.ID, item.Description, item.Path, item.Method, item.Module, item.ModuleDescription, item.ResourceGroup); err != nil {
			return err
		}
	}
	return nil
}

func (r *adminRepo) ensureBootstrapResourceRoles(ctx context.Context) error {
	const query = `
INSERT INTO resource_roles (resource_id, role_id)
VALUES
  ($1, 1),
  ($1, 2)
ON CONFLICT DO NOTHING`
	for _, item := range bootstrapAPIResourceGroups {
		if _, err := r.data.sqlDB.ExecContext(ctx, query, item.ID); err != nil {
			return err
		}
	}
	return nil
}

func (r *adminRepo) ensureBootstrapAPIRoles(ctx context.Context) error {
	const query = `
INSERT INTO api_resources_roles (api_resources_id, role_id)
SELECT id, role_id
FROM sys_api_resources
CROSS JOIN (VALUES (1), (2)) AS roles(role_id)
WHERE path = $1 AND method = $2
ON CONFLICT DO NOTHING`
	for _, item := range bootstrapAdminAPIResources {
		if _, err := r.data.sqlDB.ExecContext(ctx, query, item.Path, item.Method); err != nil {
			return err
		}
	}
	return nil
}

func (r *adminRepo) ensureDefaultOrganizationPermissionScope(ctx context.Context) error {
	if err := r.ensureDefaultOrganizationMenuScope(ctx); err != nil {
		return err
	}
	return r.ensureDefaultOrganizationResourceScope(ctx)
}

func (r *adminRepo) ensureDefaultOrganizationMenuScope(ctx context.Context) error {
	const query = `
INSERT INTO sys_organization_permission_scope (
  id, create_time, update_time, organization_id, permission_type, permission_ref, created_by
)
SELECT 'scope-' || md5($1 || ':menu:' || id::text), NOW(), NOW(), $1, 'menu', id::text, ''
FROM sys_menu
ON CONFLICT (organization_id, permission_type, permission_ref) DO NOTHING`
	_, err := r.data.sqlDB.ExecContext(ctx, query, defaultOrganizationID)
	return err
}

func (r *adminRepo) ensureDefaultOrganizationResourceScope(ctx context.Context) error {
	const query = `
INSERT INTO sys_organization_permission_scope (
  id, create_time, update_time, organization_id, permission_type, permission_ref, created_by
)
SELECT 'scope-' || md5($1 || ':resource:' || id), NOW(), NOW(), $1, 'resource', id, ''
FROM sys_resources
ON CONFLICT (organization_id, permission_type, permission_ref) DO NOTHING`
	_, err := r.data.sqlDB.ExecContext(ctx, query, defaultOrganizationID)
	return err
}

func (r *adminRepo) RegisterPermissionSnapshot(ctx context.Context) error {
	return r.data.projectionClient.RegisterSourceSnapshot(ctx, newAdminProjectionSource(r))
}

func (r *adminRepo) ApplyAuthRoleDelta(ctx context.Context, before, after *ent.Role) error {
	beforeRole, err := r.roleToPolicyRole(ctx, before)
	if err != nil {
		return err
	}
	afterRole, err := r.roleToPolicyRole(ctx, after)
	if err != nil {
		return err
	}
	return r.data.projectionClient.ApplyRoleDelta(ctx, "admin", beforeRole, afterRole)
}

func (r *adminRepo) ApplyAuthApiDelta(ctx context.Context, before, after *ent.ApiResources) error {
	return r.data.projectionClient.ApplyAPIDelta(ctx, "admin", apiToPolicyAPI(before), apiToPolicyAPI(after))
}

func (r *adminRepo) ApplyAuthUserRoleBindingDelta(ctx context.Context, before, after *ent.UserRoleBinding) error {
	beforeBinding, err := r.bindingToPolicyBinding(ctx, before)
	if err != nil {
		return err
	}
	afterBinding, err := r.bindingToPolicyBinding(ctx, after)
	if err != nil {
		return err
	}
	return r.data.projectionClient.ApplyUserRoleBindingDelta(ctx, "admin", beforeBinding, afterBinding)
}

func (r *adminRepo) resolveBindingRoleValues(ctx context.Context, bindings []*ent.UserRoleBinding) (map[int64]string, error) {
	roleIDs := make([]int64, 0, len(bindings))
	seen := make(map[int64]struct{}, len(bindings))
	for _, binding := range bindings {
		if binding == nil || binding.RoleID == 0 {
			continue
		}
		if _, ok := seen[binding.RoleID]; ok {
			continue
		}
		seen[binding.RoleID] = struct{}{}
		roleIDs = append(roleIDs, binding.RoleID)
	}
	if len(roleIDs) == 0 {
		return map[int64]string{}, nil
	}
	return r.ResolveRoleValues(ctx, roleIDs)
}

func (r *adminRepo) roleToPolicyRole(ctx context.Context, item *ent.Role) (*authx.ProjectionRole, error) {
	if item == nil {
		return nil, nil
	}
	if item.Edges.Resource == nil {
		roleItem, err := r.data.db.Role.Query().Where(role.IDEQ(item.ID)).WithResource().Only(ctx)
		if err != nil {
			return nil, err
		}
		item = roleItem
	}
	roleProjection := roleToPolicyRole(item)
	return &roleProjection, nil
}

func (r *adminRepo) bindingToPolicyBinding(ctx context.Context, item *ent.UserRoleBinding) (*authx.UserRoleBindingProjection, error) {
	if item == nil {
		return nil, nil
	}
	roleValue := ""
	if item.RoleID > 0 {
		values, err := r.ResolveRoleValues(ctx, []int64{item.RoleID})
		if err != nil {
			return nil, err
		}
		roleValue = values[item.RoleID]
	}
	binding := userRoleBindingToPolicyBinding(item, roleValue)
	return &binding, nil
}

func roleToPolicyRole(item *ent.Role) authx.ProjectionRole {
	if item == nil {
		return authx.ProjectionRole{}
	}
	resources := make([]authx.ProjectionRoleResource, 0, len(item.Edges.Resource))
	for _, resourceItem := range item.Edges.Resource {
		if resourceItem == nil {
			continue
		}
		resources = append(resources, authx.ProjectionRoleResource{
			ID:     resourceItem.ID,
			Type:   resourceItem.Type,
			Value:  resourceItem.Value,
			Method: resourceItem.Method,
		})
	}
	return authx.ProjectionRole{
		ID:               item.ID,
		Name:             item.Name,
		Value:            item.Value,
		Status:           item.Status,
		Remark:           item.Desc,
		MenuIDs:          append([]int32(nil), item.Menus...),
		Resources:        resources,
		OrganizationID:   item.OrganizationID,
		DataScope:        item.DataScope,
		DataScopeDeptIDs: append([]int64(nil), item.DataScopeDeptIds...),
	}
}

func apiToPolicyAPI(item *ent.ApiResources) *authx.ProjectionAPI {
	if item == nil {
		return nil
	}
	return &authx.ProjectionAPI{
		ID:                item.ID,
		Path:              item.Path,
		Method:            item.Method,
		Description:       item.Description,
		Module:            item.Module,
		ModuleDescription: item.ModuleDescription,
		ResourceGroup:     item.ResourcesGroup,
	}
}

func userRoleBindingToPolicyBinding(item *ent.UserRoleBinding, roleValue string) authx.UserRoleBindingProjection {
	if item == nil {
		return authx.UserRoleBindingProjection{}
	}
	return authx.UserRoleBindingProjection{
		ID:             strconv.FormatInt(item.ID, 10),
		UserID:         item.UserID,
		RoleID:         item.RoleID,
		RoleValue:      roleValue,
		CreateTime:     item.CreateTime.Format(time.RFC3339),
		OrganizationID: item.OrganizationID,
		UpdateTime:     item.UpdateTime.Format(time.RFC3339),
	}
}

func (r *adminRepo) CreateMenu(ctx context.Context, item *ent.Menu) (*ent.Menu, error) {
	defer r.data.db.Menu.Query().All(entcache.Evict(ctx))
	return r.data.db.Menu.Create().CreateAll(item).Save(ctx)
}

func (r *adminRepo) ListAuthWalkRoutes(ctx context.Context) ([]tools.WalkRouteItem, error) {
	ctx = authx.ForwardAuthorizationContext(ctx)
	res, err := r.data.authClient.GetWalkRoute(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	items := make([]tools.WalkRouteItem, 0, len(res.Items))
	for _, item := range res.Items {
		items = append(items, tools.WalkRouteItem{URL: item.Url, Method: item.Method})
	}
	return items, nil
}

func (r *adminRepo) ListUserWalkRoutes(ctx context.Context) ([]tools.WalkRouteItem, error) {
	ctx = authx.ForwardAuthorizationContext(ctx)
	res, err := r.data.userClient.GetWalkRoute(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	items := make([]tools.WalkRouteItem, 0, len(res.Items))
	for _, item := range res.Items {
		items = append(items, tools.WalkRouteItem{URL: item.Url, Method: item.Method})
	}
	return items, nil
}

func (r *adminRepo) ListCommonWalkRoutes(ctx context.Context) ([]tools.WalkRouteItem, error) {
	ctx = authx.ForwardAuthorizationContext(ctx)
	res, err := r.data.commonClient.GetWalkRoute(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	items := make([]tools.WalkRouteItem, 0, len(res.Items))
	for _, item := range res.Items {
		items = append(items, tools.WalkRouteItem{URL: item.Url, Method: item.Method})
	}
	return items, nil
}

func (r *adminRepo) UpdateMenu(ctx context.Context, id int64, item *ent.Menu) (*ent.Menu, error) {
	defer r.data.db.Menu.Query().All(entcache.Evict(ctx))
	return r.data.db.Menu.UpdateOneID(id).UpdateAll(item).Save(ctx)
}

func (r *adminRepo) DeleteMenu(ctx context.Context, id int64) error {
	defer r.data.db.Menu.Query().All(entcache.NewContext(ctx))
	return r.data.db.Menu.DeleteOneID(id).Exec(entcache.Evict(ctx))
}

func (r *adminRepo) GetDeptList(ctx context.Context, organizationID string) ([]*ent.Dept, error) {
	return r.data.db.Dept.Query().
		Where(dept.OrganizationIDEQ(organizationID)).
		Order(dept.ByPid(func(options *sql.OrderTermOptions) {
			options.NullsFirst = true
		}), dept.BySort(), dept.ByID()).
		All(ctx)
}

func (r *adminRepo) AddDept(ctx context.Context, req *v1.DeptListItem) (*ent.Dept, error) {
	cmd := r.data.db.Dept.Create().
		SetName(req.Name).
		SetSort(req.OrderNo).
		SetStatus(req.Status != 0).
		SetDesc(req.Remark).
		SetExtension("")
	if organizationID := strings.TrimSpace(req.OrganizationId); organizationID != "" {
		cmd = cmd.SetOrganizationID(organizationID)
	}
	pid, err := strconv.ParseInt(req.Pid, 10, 32)
	if err == nil && pid > 0 {
		cmd = cmd.SetPid(pid)
	}
	return cmd.Save(ctx)
}

func (r *adminRepo) UpdateDept(ctx context.Context, deptID int64, req *v1.DeptListItem) (*ent.Dept, error) {
	cmd := r.data.db.Dept.UpdateOneID(deptID).
		SetName(req.Name).
		SetSort(req.OrderNo).
		SetStatus(req.Status != 0).
		SetDesc(req.Remark).
		SetExtension("")
	if organizationID := strings.TrimSpace(req.OrganizationId); organizationID != "" {
		cmd = cmd.SetOrganizationID(organizationID)
	}
	pid, err := strconv.ParseInt(req.Pid, 10, 32)
	if err == nil && pid > 0 {
		cmd = cmd.SetPid(pid)
	} else {
		cmd = cmd.ClearPid()
	}
	return cmd.Save(ctx)
}

func (r *adminRepo) DelDept(ctx context.Context, id int64) error {
	return r.data.db.Dept.DeleteOneID(id).Exec(ctx)
}

func (r *adminRepo) GetDeptLeafsChildren(ctx context.Context, id int64) ([]*ent.Dept, error) {
	root, err := r.data.db.Dept.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return root.QueryChildren().Where(dept.Not(dept.HasChildren())).All(ctx)
}

func (r *adminRepo) GetDeptById(ctx context.Context, id int64) (*ent.Dept, error) {
	return r.data.db.Dept.Get(ctx, id)
}

func (r *adminRepo) RemoveDeptBindings(ctx context.Context, deptID int64) error {
	_, err := r.data.db.UserDeptMembership.Delete().
		Where(userdeptmembership.DeptIDEQ(deptID)).
		Exec(ctx)
	return err
}

func (r *adminRepo) ListServiceRegistries(ctx context.Context) ([]*ent.ServiceRegistry, error) {
	return r.data.db.ServiceRegistry.Query().All(ctx)
}

func (r *adminRepo) ListProjectionSourceStatuses(ctx context.Context) ([]*ent.ProjectionSourceStatus, error) {
	return r.data.db.ProjectionSourceStatus.Query().All(ctx)
}

func (r *adminRepo) GetProjectionSourceStatusBySourceService(ctx context.Context, sourceService string) (*ent.ProjectionSourceStatus, error) {
	return r.data.db.ProjectionSourceStatus.Query().
		Where(projectionsourcestatus.SourceServiceEQ(sourceService)).
		First(ctx)
}

func (r *adminRepo) addProjectionSourceStatus(ctx context.Context, status authx.ProjectionStatus) (*ent.ProjectionSourceStatus, error) {
	return r.data.db.ProjectionSourceStatus.Create().
		SetSourceService(status.SourceService).
		SetSyncMode(status.SyncMode).
		SetState(status.State).
		SetLastSnapshotRevision(status.LastSnapshotRevision).
		SetLastSyncTime(status.LastSyncTime).
		SetLastError(status.LastError).
		SetDescription(status.Description).
		Save(ctx)
}

func (r *adminRepo) updateProjectionSourceStatus(ctx context.Context, id string, status authx.ProjectionStatus) (*ent.ProjectionSourceStatus, error) {
	return r.data.db.ProjectionSourceStatus.UpdateOneID(id).
		SetSourceService(status.SourceService).
		SetSyncMode(status.SyncMode).
		SetState(status.State).
		SetLastSnapshotRevision(status.LastSnapshotRevision).
		SetLastSyncTime(status.LastSyncTime).
		SetLastError(status.LastError).
		SetDescription(status.Description).
		Save(ctx)
}

func (r *adminRepo) SaveProjectionSourceStatus(ctx context.Context, status authx.ProjectionStatus) error {
	item, err := r.GetProjectionSourceStatusBySourceService(ctx, status.SourceService)
	if err != nil {
		if !ent.IsNotFound(err) {
			return err
		}
		_, err = r.addProjectionSourceStatus(ctx, status)
		return err
	}
	_, err = r.updateProjectionSourceStatus(ctx, item.ID, status)
	return err
}

func (r *adminRepo) CreateSysLog(ctx context.Context, item *ent.SysLogRecord) error {
	_, err := r.data.db.SysLogRecord.Create().
		SetUserID(item.UserID).
		SetUserName(item.UserName).
		SetIsLogin(item.IsLogin).
		SetSessionID(item.SessionID).
		SetMethod(item.Method).
		SetPath(item.Path).
		SetRequestTime(item.RequestTime).
		SetIPAddress(item.IPAddress).
		SetIPLocation(item.IPLocation).
		SetLatency(item.Latency).
		SetOs(item.Os).
		SetBrowser(item.Browser).
		SetUserAgent(item.UserAgent).
		SetHeader(item.Header).
		SetGetParams(item.GetParams).
		SetPostData(item.PostData).
		SetResCode(item.ResCode).
		SetReason(item.Reason).
		SetResStatus(item.ResStatus).
		SetStack(item.Stack).
		Save(ctx)
	return err
}

func adminTables() []*schema.Table {
	return []*schema.Table{
		migrate.SysAPIResourcesTable,
		migrate.SysResourcesTable,
		migrate.SysRoleTable,
		migrate.APIResourcesRolesTable,
		migrate.ResourceRolesTable,
		migrate.SysUserRoleBindingTable,
		migrate.SysUserDeptMembershipTable,
		migrate.SysOrganizationTable,
		migrate.SysOrganizationPermissionScopeTable,
		migrate.SysUserOrganizationTable,
		migrate.SysMenuTable,
		migrate.SysDeptTable,
		migrate.SysLogTable,
		migrate.SysServiceRegistryTable,
		migrate.SysProjectionSourceStatusTable,
	}
}

func ensureDefaultOrganizationAfterMigration(ctx context.Context, db *dbsql.DB, driver string) error {
	if db == nil || toEntDialect(driver) != dialect.Postgres {
		return nil
	}
	organizationExists, err := tableExists(ctx, db, "sys_organization")
	if err != nil || !organizationExists {
		return err
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO public.sys_organization (
  id, create_time, update_time, name, code, sort, status, "desc", extension
) VALUES (
  $1, NOW(), NOW(), '默认组织', 'default', 0, true, '系统默认组织', ''
)
ON CONFLICT (id) DO UPDATE SET
  update_time = NOW(),
  name = EXCLUDED.name,
  code = EXCLUDED.code,
  sort = EXCLUDED.sort,
  status = EXCLUDED.status,
  "desc" = EXCLUDED."desc",
  extension = EXCLUDED.extension`, defaultOrganizationID); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
UPDATE public.sys_role
SET data_scope = CASE
  WHEN value IN ('root', 'admin') THEN 'all'
  WHEN data_scope IS NULL OR data_scope = '' THEN 'self'
  ELSE data_scope
END
WHERE data_scope IS NULL OR data_scope = '' OR value IN ('root', 'admin')`); err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
UPDATE public.sys_role
SET data_scope_dept_ids = '[]'
WHERE data_scope_dept_ids IS NULL`)
	return err
}

func tableExists(ctx context.Context, db *dbsql.DB, table string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, "SELECT to_regclass($1) IS NOT NULL", "public."+table).Scan(&exists)
	return exists, err
}

func getSysLogListQuery(params *v1.GetSysLogListParams, isPage bool) func(s *sql.Selector) {
	return func(s *sql.Selector) {
		if params.IsLogin {
			s.Where(sql.EQ(syslogrecord.FieldIsLogin, true))
		}
		if params.UserName != "" {
			s.Where(sql.EQ(syslogrecord.FieldUserName, params.UserName))
		}
		if params.IpAddress != "" {
			s.Where(sql.Like(syslogrecord.FieldIPAddress, params.IpAddress+"%"))
		}
		if params.SessionId != "" {
			s.Where(sql.EQ(syslogrecord.FieldSessionID, params.SessionId))
		}
		if params.Path != "" {
			if ok, path := replaceBracesIfExists(params.Path); ok {
				s.Where(sql.Like(syslogrecord.FieldPath, path))
			} else {
				s.Where(sql.EQ(syslogrecord.FieldPath, path))
			}
		}
		if params.RequestTimeStart != "" {
			s.Where(sql.GTE(syslogrecord.FieldRequestTime, params.RequestTimeStart))
		}
		if params.RequestTimeEnd != "" {
			s.Where(sql.LTE(syslogrecord.FieldRequestTime, params.RequestTimeEnd))
		}
		if params.Method != "" {
			s.Where(sql.EQ(syslogrecord.FieldMethod, params.Method))
		}
		if params.Latency != 0 {
			if params.Latency > 1000 {
				s.Where(sql.GTE(syslogrecord.FieldLatency, params.Latency))
			} else {
				s.Where(sql.LTE(syslogrecord.FieldLatency, params.Latency))
			}
		}
		if isPage {
			s.OrderBy(sql.Desc(syslogrecord.FieldCreateTime))
			if params.PageSize != 0 {
				s.Limit(int(params.PageSize))
			}
			if params.CurrentPage != 0 {
				s.Offset(int(tools.GetPageOffset(params.CurrentPage, params.PageSize)))
			}
		}
	}
}

func (r *adminRepo) GetSysLogList(ctx context.Context, req *v1.GetSysLogListParams) ([]*ent.SysLogRecord, int64, error) {
	res, err := r.data.db.SysLogRecord.Query().Modify(getSysLogListQuery(req, true)).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	count, err := r.data.db.SysLogRecord.Query().Modify(getSysLogListQuery(req, false)).Count(ctx)
	return res, int64(count), err
}

func (r *adminRepo) GetSysLogInfo(ctx context.Context, id string) (*ent.SysLogRecord, error) {
	return r.data.db.SysLogRecord.Query().Where(syslogrecord.IDEQ(id)).First(ctx)
}

func replaceBracesIfExists(str string) (bool, string) {
	hasBraces := strings.Contains(str, "{") || strings.Contains(str, "}")
	if !hasBraces {
		return false, str
	}
	re := regexp.MustCompile(`{[^}]*}`)
	return true, re.ReplaceAllString(str, "%")
}

func normalizeUserIDs(userIDs []string) []string {
	seen := make(map[string]struct{}, len(userIDs))
	items := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		userID = strings.TrimSpace(userID)
		if userID == "" {
			continue
		}
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		items = append(items, userID)
	}
	return items
}

func toEntDialect(driver string) string {
	switch strings.ToLower(driver) {
	case "mysql":
		return dialect.MySQL
	case "sqlite", "sqlite3":
		return dialect.SQLite
	default:
		return dialect.Postgres
	}
}
