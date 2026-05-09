//go:build wireinject
// +build wireinject

package main

import (
	"base-server/app/common/service/internal/biz"
	"base-server/app/common/service/internal/conf"
	"base-server/app/common/service/internal/data"
	"base-server/app/common/service/internal/server"
	"base-server/app/common/service/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.Llm, *conf.Auth, *conf.Services, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
