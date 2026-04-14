//go:generate go run -mod=mod github.com/google/wire/cmd/wire
package main

import (
	"flag"
	"os"

	"base-server/app/gateway/service/internal/conf"
	_ "base-server/app/gateway/service/internal/middleware/casbin"
	_ "base-server/app/gateway/service/internal/middleware/casbin/v1"
	_ "base-server/app/gateway/service/internal/middleware/httplog"
	_ "base-server/app/gateway/service/internal/middleware/httplog/v1"
	_ "base-server/app/gateway/service/internal/middleware/jwt"
	_ "base-server/app/gateway/service/internal/middleware/jwt/v1"
	_ "base-server/app/gateway/service/internal/middleware/ratelimit"
	_ "base-server/app/gateway/service/internal/middleware/ratelimit/v1"
	_ "base-server/app/gateway/service/internal/middleware/whitelist"
	_ "base-server/app/gateway/service/internal/middleware/whitelist/v1"
	"base-server/pkg/logx"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"

	_ "go.uber.org/automaxprocs"
)

var (
	defaultServiceName = "gateway"

	Name     string
	Version  string
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gateway transport.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(gateway),
	)
}

func main() {
	flag.Parse()
	if Name == "" {
		Name = defaultServiceName
	}
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

	logger, loggerCleanup := logx.New(id, Name, Version, bc.Logger)
	defer loggerCleanup()
	log.SetLogger(logger)

	app, cleanup, err := wireApp(bc.Server, bc.Gateway, bc.Services, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
