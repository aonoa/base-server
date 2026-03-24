package server

import (
	"context"
	"errors"
	"math"
	"net/http"
	"time"

	"base-server/app/gateway/service/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type httpProxyServer struct {
	*http.Server
}

func newHTTPProxyServer(handler http.Handler, addr string, c *conf.Server) *httpProxyServer {
	return &httpProxyServer{
		Server: &http.Server{
			Addr:              addr,
			Handler:           h2c.NewHandler(handler, &http2.Server{IdleTimeout: httpIdleTimeout(c), MaxConcurrentStreams: math.MaxUint32}),
			ReadTimeout:       httpReadTimeout(c),
			ReadHeaderTimeout: httpReadHeaderTimeout(c),
			WriteTimeout:      httpWriteTimeout(c),
			IdleTimeout:       httpIdleTimeout(c),
		},
	}
}

func (s *httpProxyServer) Start(ctx context.Context) error {
	log.Infof("gateway listening on %s", s.Addr)
	err := s.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *httpProxyServer) Stop(ctx context.Context) error {
	log.Info("gateway stopping")
	return s.Shutdown(ctx)
}

func httpReadHeaderTimeout(c *conf.Server) time.Duration {
	if c != nil && c.Http != nil && c.Http.ReadHeaderTimeout != nil && c.Http.ReadHeaderTimeout.AsDuration() > 0 {
		return c.Http.ReadHeaderTimeout.AsDuration()
	}
	return 10 * time.Second
}

func httpReadTimeout(c *conf.Server) time.Duration {
	if c != nil && c.Http != nil && c.Http.ReadTimeout != nil && c.Http.ReadTimeout.AsDuration() > 0 {
		return c.Http.ReadTimeout.AsDuration()
	}
	return 0
}

func httpWriteTimeout(c *conf.Server) time.Duration {
	return 0
}

func httpIdleTimeout(c *conf.Server) time.Duration {
	if c != nil && c.Http != nil && c.Http.IdleTimeout != nil && c.Http.IdleTimeout.AsDuration() > 0 {
		return c.Http.IdleTimeout.AsDuration()
	}
	return 120 * time.Second
}
