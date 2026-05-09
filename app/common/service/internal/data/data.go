package data

import (
	"context"
	dbsql "database/sql"
	"strings"

	adminv1 "base-server/api/gen/go/admin/service/v1"
	userv1 "base-server/api/gen/go/user/service/v1"
	"base-server/app/common/service/internal/biz"
	"base-server/app/common/service/internal/conf"
	"base-server/pkg/data/ent"
	"base-server/pkg/data/ent/migrate"
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
var ProviderSet = wire.NewSet(NewData, NewCommonRepo)

// Data .
type Data struct {
	db          *ent.Client
	sqlDB       *dbsql.DB
	dbDriver    string
	userConn    *grpc.ClientConn
	userClient  userv1.UserServiceClient
	adminConn   *grpc.ClientConn
	adminClient adminv1.AdminServiceClient
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
	if err := migrate.Create(context.Background(), migrate.NewSchema(sqlDrv), commonTables(), schema.WithForeignKeys(false)); err != nil {
		_ = client.Close()
		return nil, nil, err
	}

	userConn, err := grpc.NewClient(services.User.GrpcEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = client.Close()
		return nil, nil, err
	}
	adminConn, err := grpc.NewClient(services.Admin.GrpcEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = userConn.Close()
		_ = client.Close()
		return nil, nil, err
	}

	d := &Data{
		db:          client,
		sqlDB:       db,
		dbDriver:    c.Database.Driver,
		userConn:    userConn,
		userClient:  userv1.NewUserServiceClient(userConn),
		adminConn:   adminConn,
		adminClient: adminv1.NewAdminServiceClient(adminConn),
	}
	cleanup := func() {
		helper.Info("closing the data resources")
		if d.adminConn != nil {
			_ = d.adminConn.Close()
		}
		if d.userConn != nil {
			_ = d.userConn.Close()
		}
		if d.db != nil {
			_ = d.db.Close()
		}
	}
	return d, cleanup, nil
}

type commonRepo struct {
	data *Data
	log  *log.Helper
}

func commonTables() []*schema.Table {
	return []*schema.Table{
		migrate.SysSiteMessageTable,
		migrate.SysSiteMessageReceiptTable,
	}
}

func NewCommonRepo(data *Data, logger log.Logger) biz.CommonRepo {
	return &commonRepo{data: data, log: log.NewHelper(logger)}
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
