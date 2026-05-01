package main

import (
	"context"
	"flag"
	"os"

	"base-server/app/admin/service/internal/biz"
	"base-server/app/admin/service/internal/conf"
	"base-server/app/admin/service/internal/data"
	"base-server/pkg/logx"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/encoding/json"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/encoding/protojson"

	_ "go.uber.org/automaxprocs"
)

var (
	flagconf string
	id, _    = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
	json.MarshalOptions = protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
	}
}

func main() {
	flag.Parse()

	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	logger, loggerCleanup := logx.New(id, "admin-projection-sync", "", bc.Logger)
	defer loggerCleanup()
	log.SetLogger(logger)

	dataLayer, cleanup, err := data.NewData(bc.Data, bc.Services, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	repo := data.NewAdminRepo(dataLayer, logger)
	uc := biz.NewAdminUsecase(repo, logger)
	if err := uc.RegisterPermissionSnapshot(context.Background()); err != nil {
		panic(err)
	}
}
