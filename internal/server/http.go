package server

import (
	"base-server/api/base_api"
	basev1 "base-server/api/base_api/v1"
	"base-server/internal/conf"
	"base-server/internal/service"
	"github.com/casbin/casbin/v2"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/go-kratos/swagger-api/openapiv2"
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
			MiddlewareAuth(ac, e),
		),
		// 重定义返回结构
		//http.ResponseEncoder(ResponseEncoder),
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
	// http://<ip>:<port>/q/services
	// http://127.0.0.1:8000/q/swagger-ui
	openAPIhandler := openapiv2.NewHandler()
	srv.HandlePrefix("/q/", openAPIhandler)

	return srv
}
