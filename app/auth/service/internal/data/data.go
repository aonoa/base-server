package data

import (
	"context"
	dbsql "database/sql"
	"strings"

	authv1 "base-server/api/gen/go/auth/service/v1"
	userv1 "base-server/api/gen/go/user/service/v1"
	"base-server/app/auth/service/internal/biz"
	"base-server/app/auth/service/internal/conf"
	"base-server/pkg/data/ent"
	"base-server/pkg/data/ent/apiresources"
	"base-server/pkg/data/ent/migrate"
	"base-server/pkg/data/ent/resource"
	"base-server/pkg/data/ent/role"
	"base-server/pkg/tools"

	"entgo.io/ent/dialect"
	sql "entgo.io/ent/dialect/sql"
	schema "entgo.io/ent/dialect/sql/schema"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewAuthRepo)

// Data .
type Data struct {
	db         *ent.Client
	userConn   *grpc.ClientConn
	userClient userv1.UserServiceClient
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
	if err := migrate.Create(context.Background(), migrate.NewSchema(sqlDrv), authTables(), schema.WithForeignKeys(false)); err != nil {
		return nil, nil, err
	}
	conn, err := grpc.NewClient(services.User.GrpcEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = client.Close()
		return nil, nil, err
	}
	d := &Data{db: client, userConn: conn, userClient: userv1.NewUserServiceClient(conn)}
	cleanup := func() {
		helper.Info("closing the data resources")
		if d.userConn != nil {
			_ = d.userConn.Close()
		}
		_ = d.db.Close()
	}
	return d, cleanup, nil
}

type authRepo struct {
	data *Data
	log  *log.Helper
}

func NewAuthRepo(data *Data, logger log.Logger) biz.AuthRepo {
	return &authRepo{data: data, log: log.NewHelper(logger)}
}

func authTables() []*schema.Table {
	return []*schema.Table{
		migrate.SysAPIResourcesTable,
		migrate.SysResourcesTable,
		migrate.SysRoleTable,
		migrate.APIResourcesRolesTable,
		migrate.ResourceRolesTable,
	}
}

func (r *authRepo) Login(ctx context.Context, req *authv1.LoginRequest) (string, error) {
	res, err := r.data.userClient.ValidateUserAuth(ctx, &userv1.ValidateUserAuthRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return "", err
	}
	return res.UserId, nil
}

func (r *authRepo) GetUserAuthInfo(ctx context.Context, userID string) (*userv1.GetUserAuthInfoReply, error) {
	return r.data.userClient.GetUserAuthInfo(ctx, &userv1.GetUserAuthInfoRequest{UserId: userID})
}

func (r *authRepo) ListUserAuthBindings(ctx context.Context) ([]*userv1.UserAuthBinding, error) {
	res, err := r.data.userClient.ListUserAuthBindings(ctx, nil)
	if err != nil {
		return nil, err
	}
	return res.Items, nil
}

func (r *authRepo) ListRoles(ctx context.Context) ([]*ent.Role, error) {
	return r.data.db.Role.Query().
		Where(role.StatusEQ(true)).
		WithResource().
		All(ctx)
}

func (r *authRepo) ListAPIResources(ctx context.Context) ([]*ent.ApiResources, error) {
	return r.data.db.ApiResources.Query().All(ctx)
}

func (r *authRepo) ResolveRoleValues(ctx context.Context, roleIDs []int64) (map[int64]string, error) {
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

func (r *authRepo) GetAllRoleList(ctx context.Context, req *authv1.RolePageParams) ([]*ent.Role, error) {
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

func (r *authRepo) GetRole(ctx context.Context, id int64) (*ent.Role, error) {
	return r.data.db.Role.Query().Where(role.IDEQ(id)).First(ctx)
}

func (r *authRepo) AddRole(ctx context.Context, req *authv1.RoleListItem) (*ent.Role, error) {
	return r.data.db.Role.Create().
		SetName(req.Name).
		SetValue(req.Value).
		SetStatus(req.Status != 0).
		SetDesc(req.Remark).
		SetMenus(req.Permissions).
		AddResourceIDs(req.ApiPermissions...).
		Save(ctx)
}

func (r *authRepo) UpdateRole(ctx context.Context, roleID int64, req *authv1.RoleListItem) (*ent.Role, error) {
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

func (r *authRepo) DelRole(ctx context.Context, id int64) error {
	return r.data.db.Role.DeleteOneID(id).Exec(ctx)
}

func getAPIListQuery(params *authv1.GetApiPageParams, isPage bool) func(s *sql.Selector) {
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

func (r *authRepo) GetApiList(ctx context.Context, req *authv1.GetApiPageParams) ([]*ent.ApiResources, int64, error) {
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

func (r *authRepo) GetApi(ctx context.Context, id string) (*ent.ApiResources, error) {
	return r.data.db.ApiResources.Query().Where(apiresources.IDEQ(id)).First(ctx)
}

func (r *authRepo) AddApi(ctx context.Context, req *ent.ApiResources) (*ent.ApiResources, error) {
	return r.data.db.ApiResources.Create().CreateAll(req).Save(ctx)
}

func (r *authRepo) UpdateApi(ctx context.Context, req *ent.ApiResources) (*ent.ApiResources, error) {
	return r.data.db.ApiResources.UpdateOneID(req.ID).UpdateAll(req).Save(ctx)
}

func (r *authRepo) DelApi(ctx context.Context, id string) error {
	return r.data.db.ApiResources.DeleteOneID(id).Exec(ctx)
}

func getResourceListQuery(params *authv1.GetResourcePageParams, isPage bool) func(s *sql.Selector) {
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

func (r *authRepo) GetResourceList(ctx context.Context, req *authv1.GetResourcePageParams) ([]*ent.Resource, int64, error) {
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

func (r *authRepo) GetResource(ctx context.Context, id string) (*ent.Resource, error) {
	return r.data.db.Resource.Query().Where(resource.IDEQ(id)).First(ctx)
}

func (r *authRepo) AddResource(ctx context.Context, req *ent.Resource) (*ent.Resource, error) {
	return r.data.db.Resource.Create().CreateAll(req).Save(ctx)
}

func (r *authRepo) UpdateResource(ctx context.Context, req *ent.Resource) (*ent.Resource, error) {
	return r.data.db.Resource.UpdateOneID(req.ID).UpdateAll(req).Save(ctx)
}

func (r *authRepo) DelResource(ctx context.Context, id string) error {
	return r.data.db.Resource.DeleteOneID(id).Exec(ctx)
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
