package data

import (
	"context"
	dbsql "database/sql"
	"encoding/json"
	"strings"

	"ariga.io/entcache"
	v1 "base-server/api/gen/go/user/service/v1"
	"base-server/app/user/service/internal/biz"
	"base-server/app/user/service/internal/conf"
	"base-server/pkg/data/ent"
	"base-server/pkg/data/ent/migrate"
	"base-server/pkg/data/ent/user"
	"base-server/pkg/tools"
	"entgo.io/ent/dialect"
	sql "entgo.io/ent/dialect/sql"
	schema "entgo.io/ent/dialect/sql/schema"
	kratoserrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/google/wire"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewUserRepo)

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
	if err := migrate.Create(context.Background(), migrate.NewSchema(sqlDrv), userTables(), schema.WithForeignKeys(false)); err != nil {
		return nil, nil, err
	}
	d := &Data{db: client}
	cleanup := func() {
		helper.Info("closing the data resources")
		_ = d.db.Close()
	}
	return d, cleanup, nil
}

type userRepo struct {
	data *Data
	log  *log.Helper
}

func userTables() []*schema.Table {
	return []*schema.Table{
		migrate.SysUserTable,
	}
}

func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &userRepo{data: data, log: log.NewHelper(logger)}
}

func (r *userRepo) FindUserByID(ctx context.Context, id *uuid.UUID) (*ent.User, error) {
	return r.data.db.User.Query().Where(user.IDEQ(*id)).First(ctx)
}

func getUserListQuery(params *v1.GetUserParams, isPage bool) func(s *sql.Selector) {
	return func(s *sql.Selector) {
		if params.Username != "" {
			s.Where(sql.EQ(user.FieldUsername, params.Username))
		}
		if params.Nickname != "" {
			s.Where(sql.EQ(user.FieldNickname, params.Nickname))
		}
		if params.Status == 1 {
			s.Where(sql.EQ(user.FieldStatus, params.Status))
		} else if params.Status == 2 {
			s.Where(sql.EQ(user.FieldStatus, 0))
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

func (r *userRepo) GetUserList(ctx context.Context, req *v1.GetUserParams) ([]*ent.User, int64, error) {
	query := r.data.db.User.Query()
	query.Modify(getUserListQuery(req, true))
	res, err := query.All(ctx)
	if err != nil {
		return nil, 0, err
	}
	queryCount := r.data.db.User.Query()
	count, err := queryCount.Modify(getUserListQuery(req, false)).Count(ctx)
	return res, int64(count), err
}

func (r *userRepo) AddUser(ctx context.Context, req *v1.UserListItem) (*ent.User, error) {
	payload, _ := json.Marshal(struct {
		Email string `json:"email"`
	}{Email: req.Email})
	cmd := r.data.db.User.Create().
		SetUsername(req.Username).
		SetAvatar(defaultAvatar(req.Avatar)).
		SetPassword(req.Password).
		SetNickname(req.Nickname).
		SetStatus(int8(req.Status)).
		SetDesc(req.Remark).
		SetExtension(string(payload))
	return cmd.Save(ctx)
}

func (r *userRepo) UpdateUser(ctx context.Context, id *uuid.UUID, req *v1.UserListItem) (*ent.User, error) {
	defer r.data.db.User.Query().All(entcache.Evict(ctx))
	cmd := r.data.db.User.UpdateOneID(*id).
		SetUsername(req.Username).
		SetNickname(req.Nickname).
		SetAvatar(defaultAvatar(req.Avatar)).
		SetStatus(int8(req.Status)).
		SetDesc(req.Remark)
	return cmd.Save(ctx)
}

func (r *userRepo) DeleteByID(ctx context.Context, id *uuid.UUID) error {
	defer r.data.db.User.Query().All(entcache.NewContext(ctx))
	return r.data.db.User.DeleteOneID(*id).Exec(entcache.Evict(ctx))
}

func (r *userRepo) IsUserExistsByUserName(ctx context.Context, req *v1.IsUserExistsRequest) (*ent.User, error) {
	data, err := r.data.db.User.Query().Unique(false).Where(user.UsernameEQ(req.Username)).First(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	return data, err
}

func (r *userRepo) ChangePassword(ctx context.Context, uid *uuid.UUID, passwordOld, passwordNew string) error {
	password, err := r.data.db.User.Query().Where(user.IDEQ(*uid)).Select(user.FieldPassword).String(ctx)
	if err != nil {
		return err
	}
	if password != passwordOld {
		return kratoserrors.New(500, "password_err", "password err")
	}
	_, err = r.data.db.User.UpdateOneID(*uid).SetPassword(passwordNew).Save(entcache.Evict(ctx))
	return err
}

func (r *userRepo) ValidateUserAuth(ctx context.Context, username, password string) (*ent.User, error) {
	return r.data.db.User.Query().
		Unique(false).
		Where(user.And(user.UsernameEQ(username), user.PasswordEQ(password))).
		First(ctx)
}

func defaultAvatar(avatar string) string {
	if avatar != "" {
		return avatar
	}
	return "https://cdn.jsdelivr.net/gh/BaiMo-zyc/baimo.images@master/img/user-mini.png"
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
