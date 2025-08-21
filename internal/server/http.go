package server

import (
	basev1 "base-server/api/gen/go/base_api/v1"
	"base-server/api/protos/base_api"
	"base-server/cmd/base-server/assets"
	"base-server/internal/conf"
	"base-server/internal/server/middleware"
	"base-server/internal/service"
	"github.com/casbin/casbin/v2"
	swaggerUI "github.com/tx7do/kratos-swagger-ui"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, ac *conf.Auth, e *casbin.Enforcer,
	base *service.BaseService,
	logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Filter(handlers.CORS(
			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedMethods(strings.Split("PUT,POST,GET,DELETE,PATCH,OPTIONS", ",")),
		)),
		http.Middleware(
			recovery.Recovery(),
			//MiddlewareDemo(),
			middleware.MiddlewareOperateLog(),
			middleware.MiddlewareAuth(ac, e, logger),
		),
		// 重定义返回结构
		http.ResponseEncoder(DefaultResponseEncoder),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}

	srv := http.NewServer(opts...)
	basev1.RegisterBaseHTTPServer(srv, base)
	base_api.RegisterFileServiceHTTPServer(srv, base)

	base.RestServer = srv

	// http://127.0.0.1:8000/docs
	swaggerUI.RegisterSwaggerUIServerWithOption(
		srv,
		swaggerUI.WithTitle("Kratos Admin"),
		swaggerUI.WithMemoryData(assets.OpenApiData, "yaml"),
		swaggerUI.WithBasePath("/docs/"),
		//swaggerUI.WithShowTopBar(true),
	)

	return srv
}
