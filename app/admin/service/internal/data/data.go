package data

import (
	"context"
	dbsql "database/sql"
	"regexp"
	"strconv"
	"strings"

	"ariga.io/entcache"
	v1 "base-server/api/gen/go/admin/service/v1"
	"base-server/app/admin/service/internal/biz"
	"base-server/app/admin/service/internal/conf"
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
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewAdminRepo)

// Data .
type Data struct {
	db *ent.Client
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
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
	d := &Data{db: client}
	cleanup := func() {
		helper.Info("closing the data resources")
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
