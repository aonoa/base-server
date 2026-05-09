package server

import (
	pb "base-server/api/gen/go/common/service/v1"
	commonhttp "base-server/api/protos/common/service"
	"base-server/app/common/service/internal/conf"
	"base-server/app/common/service/internal/service"
	"base-server/pkg/authx"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, ac *conf.Auth, common *service.CommonService, logger log.Logger) *http.Server {
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
	pb.RegisterCommonServiceHTTPServer(srv, common)
	commonhttp.RegisterUploadServiceHTTPServer(srv, common)
	commonhttp.RegisterSSEServiceHTTPServer(srv, common)
	common.RestServer = srv
	return srv
}
