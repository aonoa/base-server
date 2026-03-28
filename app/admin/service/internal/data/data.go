package data

import (
	"context"
	dbsql "database/sql"
	"regexp"
	"strconv"
	"strings"

	"ariga.io/entcache"
	authv1 "base-server/api/gen/go/auth/service/v1"
	userv1 "base-server/api/gen/go/user/service/v1"
	v1 "base-server/api/gen/go/admin/service/v1"
	"base-server/app/admin/service/internal/biz"
	"base-server/app/admin/service/internal/conf"
	"base-server/pkg/authx"
	"base-server/pkg/data/ent"
	"base-server/pkg/data/ent/dept"
	"base-server/pkg/data/ent/menu"
	"base-server/pkg/data/ent/migrate"
	"base-server/pkg/data/ent/syslogrecord"
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
var ProviderSet = wire.NewSet(NewData, NewAdminRepo)

// Data .
type Data struct {
	db         *ent.Client
	userConn   *grpc.ClientConn
	userClient userv1.UserServiceClient
	authConn   *grpc.ClientConn
	authClient authv1.AuthServiceClient
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
	d := &Data{
		db:         client,
		userConn:   userConn,
		userClient: userv1.NewUserServiceClient(userConn),
		authConn:   authConn,
		authClient: authv1.NewAuthServiceClient(authConn),
	}
	cleanup := func() {
		helper.Info("closing the data resources")
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
	return &adminRepo{data: data, log: log.NewHelper(logger)}
}

func (r *adminRepo) GetMenuList(ctx context.Context) ([]*ent.Menu, error) {
	return r.data.db.Menu.Query().Order(menu.ByPid(), menu.ByOrder()).All(ctx)
}

func (r *adminRepo) GetUserAuthInfo(ctx context.Context, userID string) (*userv1.GetUserAuthInfoReply, error) {
	ctx = authx.ForwardAuthorizationContext(ctx)
	return r.data.userClient.GetUserAuthInfo(ctx, &userv1.GetUserAuthInfoRequest{UserId: userID})
}

func (r *adminRepo) GetCurrentUserMenuAuthority(ctx context.Context, userID string) (*authv1.GetCurrentUserMenuAuthorityReply, error) {
	ctx = authx.ForwardAuthorizationContext(ctx)
	return r.data.authClient.GetCurrentUserMenuAuthority(ctx, &authv1.GetCurrentUserMenuAuthorityRequest{UserId: userID})
}

func (r *adminRepo) CreateMenu(ctx context.Context, item *ent.Menu) (*ent.Menu, error) {
	defer r.data.db.Menu.Query().All(entcache.Evict(ctx))
	return r.data.db.Menu.Create().CreateAll(item).Save(ctx)
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
		migrate.SysMenuTable,
		migrate.SysDeptTable,
		migrate.SysLogTable,
	}
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
