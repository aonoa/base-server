//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build wireinject
// +build wireinject

package main

import (
	"base-server/app/gateway/service/internal/conf"
	"base-server/app/gateway/service/internal/server"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Gateway, *conf.Services, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, newApp))
}
