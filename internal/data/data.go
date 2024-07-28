package data

import (
	"base-server/internal/conf"
	"base-server/internal/data/ent"
	"context"
	"database/sql"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"time"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewBaseRepo)

const (
	maxIdleConns    = 6
	maxOpenConns    = 100
	connMaxLifetime = time.Hour * 2
)

// Data .
type Data struct {
	// TODO wrapped database client
	db *ent.Client
}

// NewData .
func NewData(conf *conf.Data, logger log.Logger) (*Data, func(), error) {

	log := log.NewHelper(logger)
	db, err := sql.Open(
		conf.Database.Driver,
		conf.Database.Source,
	)

	db.SetMaxIdleConns(maxIdleConns)
	db.SetMaxOpenConns(maxOpenConns)
	db.SetConnMaxLifetime(0)

	drv := entsql.OpenDB(dialect.Postgres, db)

	sqlDrv := dialect.DebugWithContext(drv, func(ctx context.Context, i ...interface{}) {
		log.WithContext(ctx).Info(i...)
		// 注释部分是链路追踪需要的
		//tracer := otel.Tracer("ent.")
		//kind := trace.SpanKindServer
		//_, span := tracer.Start(ctx,
		//	"Query",
		//	trace.WithAttributes(
		//		attribute.String("sql", fmt.Sprint(i...)),
		//	),
		//	trace.WithSpanKind(kind),
		//)
		//span.End()
	})

	// 数据库缓存相关
	//drv1 := entcache.NewDriver(
	//	sqlDrv,
	//	//entcache.TTL(time.Second*10),
	//	entcache.Levels(entcache.NewLRU(128)),
	//)
	//client := ent.NewClient(ent.Driver(drv1))

	client := ent.NewClient(ent.Driver(sqlDrv))
	if err != nil {
		log.Errorf("failed opening connection to sqlite: %v", err)
		return nil, nil, err
	}
	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Errorf("failed creating schema resources: %v", err)
		return nil, nil, err
	}

	//migrate.Create(context.Background(), client.Schema, []*schema.Table{migrate.UsersTable})

	d := &Data{
		db: client,
	}
	cleanup := func() {
		log.Info("message", "closing the data resources")
		if err := d.db.Close(); err != nil {
			log.Error(err)
		}
	}
	return d, cleanup, nil
}
