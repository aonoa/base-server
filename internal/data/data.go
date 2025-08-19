package data

import (
	"base-server/internal/conf"
	"base-server/internal/data/ent"
	"context"
	"database/sql"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/dgraph-io/ristretto"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	ristretto_store "github.com/eko/gocache/store/ristretto/v4"
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
	// 本地缓存库
	cacheManager *cache.Cache[string]
	// gorm数据库 演示
	//msdb *gorm.DB
}

// NewData .
func NewData(conf *conf.Data, logger log.Logger) (*Data, func(), error) {

	log := log.NewHelper(logger)
	db, err := sql.Open(
		conf.Database.Driver,
		conf.Database.Source,
	)

	if err != nil {
		panic(err)
	}

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
	//////////////////////////////////////////////////////////////////////////////////////////////////////////// 本地缓存
	ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,
		MaxCost:     100000000, // 100MB
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	ristrettoStore := ristretto_store.NewRistretto(ristrettoCache, store.WithSynchronousSet())
	cacheManager := cache.New[string](ristrettoStore)
	//// 存
	//err = cacheManager.Set(context.Background(), "my-key", "my-value", store.WithCost(2))
	//if err != nil {
	//	panic(err)
	//}
	//// 取
	//value, err := cacheManager.Get(context.Background(), "my-key")
	//if err != nil {
	//	panic(err)
	//}
	//println(value)

	/////////////////////////////////////////////////////////////////////////////////////////////////////////// gorm演示
	//msdb, err := gorm.Open(conf.OtherDatabase.Driver, conf.OtherDatabase.Source)
	//msdb, err := gorm.Open(sqlserver.Open(conf.OtherDatabase.Source), &gorm.Config{})
	//if err != nil {
	//	log.Errorf("failed opening connection to gorm: %v", err)
	//	return nil, nil, err
	//}
	///////////////////////////////////////////////////////////////////////////////////////////////////////////

	d := &Data{
		db:           client,
		cacheManager: cacheManager,
		//msdb:         msdb,       // gorm演示
	}
	cleanup := func() {
		log.Info("message", "closing the data resources")
		if err := d.db.Close(); err != nil {
			log.Error(err)
		}
	}
	return d, cleanup, nil
}
