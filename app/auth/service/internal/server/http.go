package server

import (
	v1 "base-server/api/gen/go/auth/service/v1"
	"base-server/app/auth/service/internal/conf"
	"base-server/app/auth/service/internal/service"
	"base-server/pkg/authx"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, ac *conf.Auth, auth *service.AuthService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			authx.NewServerMiddleware(ac.ApiKey, ac.Whitelist),
		),
	}
	_ = logger
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
	v1.RegisterAuthServiceHTTPServer(srv, auth)
	auth.RestServer = srv
	return srv
}
