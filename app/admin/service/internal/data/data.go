package data

import (
	"context"
	dbsql "database/sql"
	"regexp"
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
	"base-server/pkg/data/ent/businessdomain"
	"base-server/pkg/data/ent/dept"
	"base-server/pkg/data/ent/menu"
	"base-server/pkg/data/ent/migrate"
	"base-server/pkg/data/ent/projectionsourcestatus"
	"base-server/pkg/data/ent/resource"
	"base-server/pkg/data/ent/role"
	"base-server/pkg/data/ent/serviceregistry"
	"base-server/pkg/data/ent/syslogrecord"
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
	if err := cleanupDuplicateAPIResourcesBeforeMigration(context.Background(), db, c.Database.Driver); err != nil {
		_ = client.Close()
		return nil, nil, err
	}
	if err := cleanupUserRoleBindingBeforeMigration(context.Background(), db, c.Database.Driver); err != nil {
		_ = client.Close()
		return nil, nil, err
	}
	if err := migrate.Create(context.Background(), migrate.NewSchema(sqlDrv), adminTables(), schema.WithForeignKeys(false)); err != nil {
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
		data.projectionClient.SetStatusReporter(newAdminProjectionStatusReporter(repo), "platform", "admin permission projection source")
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
	Service           string
	Domain            string
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
}

var bootstrapAdminAPIResources = []bootstrapAPIResource{
	{
		ID:                "api-admin-api-list",
		Description:       "获取 API 列表",
		Path:              "/admin-api/v1/apis",
		Method:            "GET",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "api",
		Service:           "admin",
		Domain:            "platform",
	},
	{
		ID:                "api-admin-api-create",
		Description:       "新增 API",
		Path:              "/admin-api/v1/apis",
		Method:            "POST",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "api",
		Service:           "admin",
		Domain:            "platform",
	},
	{
		ID:                "api-admin-api-update",
		Description:       "更新 API",
		Path:              "/admin-api/v1/apis/{id}",
		Method:            "PUT",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "api",
		Service:           "admin",
		Domain:            "platform",
	},
	{
		ID:                "api-admin-api-delete",
		Description:       "删除 API",
		Path:              "/admin-api/v1/apis/{id}",
		Method:            "DELETE",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "api",
		Service:           "admin",
		Domain:            "platform",
	},
	{
		ID:                "api-admin-resource-list",
		Description:       "获取资源列表",
		Path:              "/admin-api/v1/resources",
		Method:            "GET",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "data",
		Service:           "admin",
		Domain:            "platform",
	},
	{
		ID:                "api-admin-resource-create",
		Description:       "新增资源",
		Path:              "/admin-api/v1/resources",
		Method:            "POST",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "data",
		Service:           "admin",
		Domain:            "platform",
	},
	{
		ID:                "api-admin-resource-update",
		Description:       "更新资源",
		Path:              "/admin-api/v1/resources/{id}",
		Method:            "PUT",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "data",
		Service:           "admin",
		Domain:            "platform",
	},
	{
		ID:                "api-admin-resource-delete",
		Description:       "删除资源",
		Path:              "/admin-api/v1/resources/{id}",
		Method:            "DELETE",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "data",
		Service:           "admin",
		Domain:            "platform",
	},
	{
		ID:                "api-admin-platform-domain-list",
		Description:       "获取业务域列表",
		Path:              "/admin-api/v1/platform/domains",
		Method:            "GET",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "admin",
		Service:           "admin",
		Domain:            "platform",
	},
	{
		ID:                "api-admin-platform-domain-create",
		Description:       "新增业务域",
		Path:              "/admin-api/v1/platform/domains",
		Method:            "POST",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "admin",
		Service:           "admin",
		Domain:            "platform",
	},
	{
		ID:                "api-admin-platform-domain-update",
		Description:       "更新业务域",
		Path:              "/admin-api/v1/platform/domains/{id}",
		Method:            "PUT",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "admin",
		Service:           "admin",
		Domain:            "platform",
	},
	{
		ID:                "api-admin-platform-domain-delete",
		Description:       "删除业务域",
		Path:              "/admin-api/v1/platform/domains/{id}",
		Method:            "DELETE",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "admin",
		Service:           "admin",
		Domain:            "platform",
	},
	{
		ID:                "api-admin-platform-service-list",
		Description:       "获取服务注册列表",
		Path:              "/admin-api/v1/platform/services",
		Method:            "GET",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "admin",
		Service:           "admin",
		Domain:            "platform",
	},
	{
		ID:                "api-admin-platform-service-create",
		Description:       "新增服务注册",
		Path:              "/admin-api/v1/platform/services",
		Method:            "POST",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "admin",
		Service:           "admin",
		Domain:            "platform",
	},
	{
		ID:                "api-admin-platform-service-update",
		Description:       "更新服务注册",
		Path:              "/admin-api/v1/platform/services/{id}",
		Method:            "PUT",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "admin",
		Service:           "admin",
		Domain:            "platform",
	},
	{
		ID:                "api-admin-platform-service-delete",
		Description:       "删除服务注册",
		Path:              "/admin-api/v1/platform/services/{id}",
		Method:            "DELETE",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "admin",
		Service:           "admin",
		Domain:            "platform",
	},
	{
		ID:                "api-admin-platform-projection-source-list",
		Description:       "获取投影源状态列表",
		Path:              "/admin-api/v1/platform/projection-sources",
		Method:            "GET",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "admin",
		Service:           "admin",
		Domain:            "platform",
	},
	{
		ID:                "api-admin-platform-projection-source-create",
		Description:       "新增投影源状态",
		Path:              "/admin-api/v1/platform/projection-sources",
		Method:            "POST",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "admin",
		Service:           "admin",
		Domain:            "platform",
	},
	{
		ID:                "api-admin-platform-projection-source-update",
		Description:       "更新投影源状态",
		Path:              "/admin-api/v1/platform/projection-sources/{id}",
		Method:            "PUT",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "admin",
		Service:           "admin",
		Domain:            "platform",
	},
	{
		ID:                "api-admin-platform-projection-source-report",
		Description:       "上报投影源状态",
		Path:              "/admin-api/v1/platform/projection-sources/report/{source_service}",
		Method:            "PUT",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "admin",
		Service:           "admin",
		Domain:            "platform",
	},
	{
		ID:                "api-admin-platform-projection-source-delete",
		Description:       "删除投影源状态",
		Path:              "/admin-api/v1/platform/projection-sources/{id}",
		Method:            "DELETE",
		Module:            "admin",
		ModuleDescription: "管理服务",
		ResourceGroup:     "admin",
		Service:           "admin",
		Domain:            "platform",
	},
	{
		ID:                "api-admin-walk-route",
		Description:       "获取系统所有api接口",
		Path:              "/admin-api/v1/walk-routes",
		Method:            "GET",
		Module:            "admin",
		ModuleDescription: "系统管理",
		ResourceGroup:     "api",
		Service:           "admin",
		Domain:            "platform",
	},
	{
		ID:                "api-common-site-message-my-list",
		Description:       "获取我的站内信列表",
		Path:              "/common-api/v1/site-messages/my",
		Method:            "GET",
		Module:            "common",
		ModuleDescription: "公共服务",
		ResourceGroup:     "default",
		Service:           "common",
		Domain:            "platform",
	},
	{
		ID:                "api-common-site-message-my-unread-count",
		Description:       "获取我的站内信未读数量",
		Path:              "/common-api/v1/site-messages/my/unread-count",
		Method:            "GET",
		Module:            "common",
		ModuleDescription: "公共服务",
		ResourceGroup:     "default",
		Service:           "common",
		Domain:            "platform",
	},
	{
		ID:                "api-common-site-message-my-read",
		Description:       "标记站内信已读",
		Path:              "/common-api/v1/site-messages/my/{message_id}/read",
		Method:            "POST",
		Module:            "common",
		ModuleDescription: "公共服务",
		ResourceGroup:     "default",
		Service:           "common",
		Domain:            "platform",
	},
	{
		ID:                "api-common-site-message-my-unread",
		Description:       "标记站内信未读",
		Path:              "/common-api/v1/site-messages/my/{message_id}/unread",
		Method:            "POST",
		Module:            "common",
		ModuleDescription: "公共服务",
		ResourceGroup:     "default",
		Service:           "common",
		Domain:            "platform",
	},
	{
		ID:                "api-common-site-message-my-read-all",
		Description:       "全部站内信标记已读",
		Path:              "/common-api/v1/site-messages/my/read-all",
		Method:            "POST",
		Module:            "common",
		ModuleDescription: "公共服务",
		ResourceGroup:     "default",
		Service:           "common",
		Domain:            "platform",
	},
	{
		ID:                "api-common-site-message-manage-list",
		Description:       "获取站内信发布记录",
		Path:              "/common-api/v1/site-messages/manage",
		Method:            "GET",
		Module:            "common",
		ModuleDescription: "公共服务",
		ResourceGroup:     "site_message_manage",
		Service:           "common",
		Domain:            "platform",
	},
	{
		ID:                "api-common-site-message-manage-create",
		Description:       "保存或发布站内信",
		Path:              "/common-api/v1/site-messages/manage",
		Method:            "POST",
		Module:            "common",
		ModuleDescription: "公共服务",
		ResourceGroup:     "site_message_manage",
		Service:           "common",
		Domain:            "platform",
	},
	{
		ID:                "api-common-site-message-manage-recall",
		Description:       "撤回站内信",
		Path:              "/common-api/v1/site-messages/manage/{id}/recall",
		Method:            "POST",
		Module:            "common",
		ModuleDescription: "公共服务",
		ResourceGroup:     "site_message_manage",
		Service:           "common",
		Domain:            "platform",
	},
	{
		ID:                "api-common-site-message-manage-delete",
		Description:       "删除未发布站内信",
		Path:              "/common-api/v1/site-messages/manage/{id}",
		Method:            "DELETE",
		Module:            "common",
		ModuleDescription: "公共服务",
		ResourceGroup:     "site_message_manage",
		Service:           "common",
		Domain:            "platform",
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

func (r *adminRepo) GetUserRoleBindings(ctx context.Context, userID string) ([]*ent.UserRoleBinding, error) {
	return r.data.db.UserRoleBinding.Query().
		Where(userrolebinding.UserIDEQ(userID)).
		Order(userrolebinding.ByRoleID()).
		All(ctx)
}

func (r *adminRepo) ListUserRoleBindings(ctx context.Context) ([]*ent.UserRoleBinding, error) {
	return r.data.db.UserRoleBinding.Query().
		Order(userrolebinding.ByUserID(), userrolebinding.ByRoleID()).
		All(ctx)
}

func (r *adminRepo) UpsertUserRoleBinding(ctx context.Context, userID string, roleIDs []int64) ([]*ent.UserRoleBinding, error) {
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

	query := tx.UserRoleBinding.Delete().Where(userrolebinding.UserIDEQ(userID))
	if len(roleIDs) > 0 {
		query = query.Where(userrolebinding.RoleIDNotIn(roleIDs...))
	}
	if _, err := query.Exec(ctx); err != nil {
		return nil, err
	}
	existing, err := tx.UserRoleBinding.Query().
		Where(userrolebinding.UserIDEQ(userID)).
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
			SetRoleID(roleID).
			Save(ctx); err != nil {
			return nil, err
		}
	}
	items, err := tx.UserRoleBinding.Query().
		Where(userrolebinding.UserIDEQ(userID)).
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

func (r *adminRepo) DeleteUserRoleBinding(ctx context.Context, userID string) error {
	_, err := r.data.db.UserRoleBinding.Delete().Where(userrolebinding.UserIDEQ(userID)).Exec(ctx)
	return err
}

func (r *adminRepo) AddRole(ctx context.Context, req *v1.RoleListItem) (*ent.Role, error) {
	return r.data.db.Role.Create().
		SetName(req.Name).
		SetValue(req.Value).
		SetStatus(req.Status != 0).
		SetDesc(req.Remark).
		SetMenus(req.Permissions).
		AddResourceIDs(req.ApiPermissions...).
		Save(ctx)
}

func (r *adminRepo) UpdateRole(ctx context.Context, roleID int64, req *v1.RoleListItem) (*ent.Role, error) {
	return r.data.db.Role.UpdateOneID(roleID).
		SetName(req.Name).
		SetValue(req.Value).
		SetStatus(req.Status != 0).
		SetDesc(req.Remark).
		SetMenus(req.Permissions).
		ClearResource().
		AddResourceIDs(req.ApiPermissions...).
		Save(ctx)
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
		if params.ServiceCode != "" {
			s.Where(sql.EQ(apiresources.FieldServiceCode, params.ServiceCode))
		}
		if params.DomainCode != "" {
			s.Where(sql.EQ(apiresources.FieldDomainCode, params.DomainCode))
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
	if toEntDialect(r.data.dbDriver) == dialect.Postgres {
		if err := cleanupDuplicateAPIResources(ctx, r.data.sqlDB, true); err != nil {
			return err
		}
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
	return r.ensureBootstrapAPIRoles(ctx)
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
  id, create_time, update_time, description, path, method, module, module_description, resources_group, service_code, domain_code
) VALUES (
  $1, NOW(), NOW(), $2, $3, $4, $5, $6, $7, $8, $9
)
ON CONFLICT (path, method) DO UPDATE SET
  update_time = NOW(),
  description = EXCLUDED.description,
  module = EXCLUDED.module,
  module_description = EXCLUDED.module_description,
  resources_group = EXCLUDED.resources_group,
  service_code = EXCLUDED.service_code,
  domain_code = EXCLUDED.domain_code`
	for _, item := range bootstrapAdminAPIResources {
		if _, err := r.data.sqlDB.ExecContext(ctx, query, item.ID, item.Description, item.Path, item.Method, item.Module, item.ModuleDescription, item.ResourceGroup, item.Service, item.Domain); err != nil {
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
		ID:        item.ID,
		Name:      item.Name,
		Value:     item.Value,
		Status:    item.Status,
		Remark:    item.Desc,
		MenuIDs:   append([]int32(nil), item.Menus...),
		Resources: resources,
		Domain:    "platform",
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
		Service:           item.ServiceCode,
		Domain:            item.DomainCode,
	}
}

func userRoleBindingToPolicyBinding(item *ent.UserRoleBinding, roleValue string) authx.UserRoleBindingProjection {
	if item == nil {
		return authx.UserRoleBindingProjection{}
	}
	return authx.UserRoleBindingProjection{
		ID:         strconv.FormatInt(item.ID, 10),
		UserID:     item.UserID,
		RoleID:     item.RoleID,
		RoleValue:  roleValue,
		CreateTime: item.CreateTime.Format(time.RFC3339),
		UpdateTime: item.UpdateTime.Format(time.RFC3339),
		Domain:     "platform",
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

func (r *adminRepo) GetDeptList(ctx context.Context) ([]*ent.Dept, error) {
	return r.data.db.Dept.Query().Order(dept.ByPid(func(options *sql.OrderTermOptions) {
		options.NullsFirst = true
	})).All(ctx)
}

func (r *adminRepo) AddDept(ctx context.Context, req *v1.DeptListItem) (*ent.Dept, error) {
	cmd := r.data.db.Dept.Create().
		SetName(req.Name).
		SetSort(req.OrderNo).
		SetStatus(req.Status != 0).
		SetDesc(req.Remark).
		SetExtension("").
		SetDom(req.Dom)
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
	pid, err := strconv.ParseInt(req.Pid, 10, 32)
	if err == nil && pid > 0 {
		cmd = cmd.SetPid(pid)
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

func (r *adminRepo) ListBusinessDomains(ctx context.Context) ([]*ent.BusinessDomain, error) {
	return r.data.db.BusinessDomain.Query().All(ctx)
}

func (r *adminRepo) GetBusinessDomain(ctx context.Context, id string) (*ent.BusinessDomain, error) {
	return r.data.db.BusinessDomain.Query().Where(businessdomain.IDEQ(id)).First(ctx)
}

func (r *adminRepo) AddBusinessDomain(ctx context.Context, req *v1.BusinessDomainItem) (*ent.BusinessDomain, error) {
	return r.data.db.BusinessDomain.Create().
		SetCode(req.Code).
		SetName(req.Name).
		SetOwnerService(req.OwnerService).
		SetOrgModelType(req.OrgModelType).
		SetAuthScopeType(req.AuthScopeType).
		SetStatus(req.Status != 0).
		SetDescription(req.Description).
		SetMetaJSON(req.MetaJson).
		Save(ctx)
}

func (r *adminRepo) UpdateBusinessDomain(ctx context.Context, id string, req *v1.BusinessDomainItem) (*ent.BusinessDomain, error) {
	return r.data.db.BusinessDomain.UpdateOneID(id).
		SetCode(req.Code).
		SetName(req.Name).
		SetOwnerService(req.OwnerService).
		SetOrgModelType(req.OrgModelType).
		SetAuthScopeType(req.AuthScopeType).
		SetStatus(req.Status != 0).
		SetDescription(req.Description).
		SetMetaJSON(req.MetaJson).
		Save(ctx)
}

func (r *adminRepo) DeleteBusinessDomain(ctx context.Context, id string) error {
	return r.data.db.BusinessDomain.DeleteOneID(id).Exec(ctx)
}

func (r *adminRepo) ListServiceRegistries(ctx context.Context) ([]*ent.ServiceRegistry, error) {
	return r.data.db.ServiceRegistry.Query().All(ctx)
}

func (r *adminRepo) GetServiceRegistry(ctx context.Context, id string) (*ent.ServiceRegistry, error) {
	return r.data.db.ServiceRegistry.Query().Where(serviceregistry.IDEQ(id)).First(ctx)
}

func (r *adminRepo) AddServiceRegistry(ctx context.Context, req *v1.ServiceRegistryItem) (*ent.ServiceRegistry, error) {
	return r.data.db.ServiceRegistry.Create().
		SetServiceCode(req.ServiceCode).
		SetServiceName(req.ServiceName).
		SetDomainCode(req.DomainCode).
		SetHTTPPrefix(req.HttpPrefix).
		SetGrpcService(req.GrpcService).
		SetStatus(req.Status != 0).
		SetProjectionEnabled(req.ProjectionEnabled).
		SetDescription(req.Description).
		Save(ctx)
}

func (r *adminRepo) UpdateServiceRegistry(ctx context.Context, id string, req *v1.ServiceRegistryItem) (*ent.ServiceRegistry, error) {
	return r.data.db.ServiceRegistry.UpdateOneID(id).
		SetServiceCode(req.ServiceCode).
		SetServiceName(req.ServiceName).
		SetDomainCode(req.DomainCode).
		SetHTTPPrefix(req.HttpPrefix).
		SetGrpcService(req.GrpcService).
		SetStatus(req.Status != 0).
		SetProjectionEnabled(req.ProjectionEnabled).
		SetDescription(req.Description).
		Save(ctx)
}

func (r *adminRepo) DeleteServiceRegistry(ctx context.Context, id string) error {
	return r.data.db.ServiceRegistry.DeleteOneID(id).Exec(ctx)
}

func (r *adminRepo) ListProjectionSourceStatuses(ctx context.Context) ([]*ent.ProjectionSourceStatus, error) {
	return r.data.db.ProjectionSourceStatus.Query().All(ctx)
}

func (r *adminRepo) GetProjectionSourceStatus(ctx context.Context, id string) (*ent.ProjectionSourceStatus, error) {
	return r.data.db.ProjectionSourceStatus.Query().Where(projectionsourcestatus.IDEQ(id)).First(ctx)
}

func (r *adminRepo) GetProjectionSourceStatusBySourceService(ctx context.Context, sourceService string) (*ent.ProjectionSourceStatus, error) {
	return r.data.db.ProjectionSourceStatus.Query().
		Where(projectionsourcestatus.SourceServiceEQ(sourceService)).
		First(ctx)
}

func (r *adminRepo) AddProjectionSourceStatus(ctx context.Context, req *v1.ProjectionSourceStatusItem) (*ent.ProjectionSourceStatus, error) {
	return r.data.db.ProjectionSourceStatus.Create().
		SetSourceService(req.SourceService).
		SetDomainCode(req.DomainCode).
		SetSyncMode(req.SyncMode).
		SetState(req.State).
		SetLastSnapshotRevision(req.LastSnapshotRevision).
		SetLastSyncTime(req.LastSyncTime).
		SetLastError(req.LastError).
		SetDescription(req.Description).
		Save(ctx)
}

func (r *adminRepo) UpdateProjectionSourceStatus(ctx context.Context, id string, req *v1.ProjectionSourceStatusItem) (*ent.ProjectionSourceStatus, error) {
	return r.data.db.ProjectionSourceStatus.UpdateOneID(id).
		SetSourceService(req.SourceService).
		SetDomainCode(req.DomainCode).
		SetSyncMode(req.SyncMode).
		SetState(req.State).
		SetLastSnapshotRevision(req.LastSnapshotRevision).
		SetLastSyncTime(req.LastSyncTime).
		SetLastError(req.LastError).
		SetDescription(req.Description).
		Save(ctx)
}

func (r *adminRepo) DeleteProjectionSourceStatus(ctx context.Context, id string) error {
	return r.data.db.ProjectionSourceStatus.DeleteOneID(id).Exec(ctx)
}

func (r *adminRepo) ReportProjectionSourceStatus(ctx context.Context, req *v1.ProjectionSourceStatusItem) (*ent.ProjectionSourceStatus, error) {
	item, err := r.GetProjectionSourceStatusBySourceService(ctx, req.SourceService)
	if err != nil {
		if !ent.IsNotFound(err) {
			return nil, err
		}
		return r.AddProjectionSourceStatus(ctx, req)
	}
	return r.UpdateProjectionSourceStatus(ctx, item.ID, req)
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
		migrate.SysMenuTable,
		migrate.SysDeptTable,
		migrate.SysLogTable,
		migrate.SysBusinessDomainTable,
		migrate.SysServiceRegistryTable,
		migrate.SysProjectionSourceStatusTable,
	}
}

func cleanupDuplicateAPIResourcesBeforeMigration(ctx context.Context, db *dbsql.DB, driver string) error {
	if toEntDialect(driver) != dialect.Postgres {
		return nil
	}
	exists, err := tableExists(ctx, db, "sys_api_resources")
	if err != nil || !exists {
		return err
	}
	roleTableExists, err := tableExists(ctx, db, "api_resources_roles")
	if err != nil {
		return err
	}
	return cleanupDuplicateAPIResources(ctx, db, roleTableExists)
}

func cleanupUserRoleBindingBeforeMigration(ctx context.Context, db *dbsql.DB, driver string) error {
	if toEntDialect(driver) != dialect.Postgres {
		return nil
	}
	exists, err := tableExists(ctx, db, "sys_user_role_binding")
	if err != nil || !exists {
		return err
	}
	if _, err := db.ExecContext(ctx, `
WITH ranked AS (
  SELECT
    id,
    ROW_NUMBER() OVER (
      PARTITION BY user_id, role_id
      ORDER BY create_time, id
    ) AS rn
  FROM sys_user_role_binding
)
DELETE FROM sys_user_role_binding
WHERE id IN (SELECT id FROM ranked WHERE rn > 1)`); err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
DO $$
DECLARE
  item record;
BEGIN
  FOR item IN
    SELECT c.conname
    FROM pg_constraint c
    JOIN LATERAL unnest(c.conkey) AS k(attnum) ON true
    JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = k.attnum
    WHERE c.conrelid = 'public.sys_user_role_binding'::regclass
      AND c.contype = 'u'
    GROUP BY c.conname
    HAVING bool_or(a.attname = 'user_id') AND NOT bool_or(a.attname = 'role_id')
  LOOP
    EXECUTE format('ALTER TABLE public.sys_user_role_binding DROP CONSTRAINT %I', item.conname);
  END LOOP;

  FOR item IN
    SELECT i.relname
    FROM pg_class t
    JOIN pg_index ix ON t.oid = ix.indrelid
    JOIN pg_class i ON i.oid = ix.indexrelid
    JOIN LATERAL unnest(ix.indkey) AS k(attnum) ON true
    JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = k.attnum
    WHERE t.oid = 'public.sys_user_role_binding'::regclass
      AND ix.indisunique
      AND NOT ix.indisprimary
    GROUP BY i.relname
    HAVING bool_or(a.attname = 'user_id') AND NOT bool_or(a.attname = 'role_id')
  LOOP
    EXECUTE format('DROP INDEX IF EXISTS public.%I', item.relname);
  END LOOP;
END $$`)
	return err
}

func tableExists(ctx context.Context, db *dbsql.DB, table string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, "SELECT to_regclass($1) IS NOT NULL", "public."+table).Scan(&exists)
	return exists, err
}

func cleanupDuplicateAPIResources(ctx context.Context, db *dbsql.DB, moveRoleLinks bool) error {
	if db == nil {
		return nil
	}
	if moveRoleLinks {
		if _, err := db.ExecContext(ctx, `
WITH ranked AS (
  SELECT
    id,
    FIRST_VALUE(id) OVER (
      PARTITION BY path, method
      ORDER BY
        CASE WHEN id LIKE 'api-%' THEN 0 ELSE 1 END,
        create_time,
        id
    ) AS keep_id,
    ROW_NUMBER() OVER (
      PARTITION BY path, method
      ORDER BY
        CASE WHEN id LIKE 'api-%' THEN 0 ELSE 1 END,
        create_time,
        id
    ) AS rn
  FROM sys_api_resources
),
duplicate_links AS (
  SELECT ranked.keep_id, api_resources_roles.role_id
  FROM ranked
  JOIN api_resources_roles ON api_resources_roles.api_resources_id = ranked.id
  WHERE ranked.rn > 1
)
INSERT INTO api_resources_roles (api_resources_id, role_id)
SELECT keep_id, role_id
FROM duplicate_links
ON CONFLICT DO NOTHING`); err != nil {
			return err
		}
		if _, err := db.ExecContext(ctx, `
WITH ranked AS (
  SELECT
    id,
    ROW_NUMBER() OVER (
      PARTITION BY path, method
      ORDER BY
        CASE WHEN id LIKE 'api-%' THEN 0 ELSE 1 END,
        create_time,
        id
    ) AS rn
  FROM sys_api_resources
)
DELETE FROM api_resources_roles
WHERE api_resources_id IN (SELECT id FROM ranked WHERE rn > 1)`); err != nil {
			return err
		}
	}
	_, err := db.ExecContext(ctx, `
WITH ranked AS (
  SELECT
    id,
    ROW_NUMBER() OVER (
      PARTITION BY path, method
      ORDER BY
        CASE WHEN id LIKE 'api-%' THEN 0 ELSE 1 END,
        create_time,
        id
    ) AS rn
  FROM sys_api_resources
)
DELETE FROM sys_api_resources
WHERE id IN (SELECT id FROM ranked WHERE rn > 1)`)
	return err
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
