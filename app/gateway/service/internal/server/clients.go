package server

import (
	authv1 "base-server/api/gen/go/auth/service/v1"
	adminv1 "base-server/api/gen/go/admin/service/v1"
	"base-server/app/gateway/service/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GatewayClients struct {
	Auth  authv1.AuthServiceClient
	Admin adminv1.AdminServiceClient
}

func NewGatewayClients(services *conf.Services, logger log.Logger) (*GatewayClients, func(), error) {
	_ = logger
	cleanupFns := make([]func(), 0, 2)
	cleanup := func() {
		for i := len(cleanupFns) - 1; i >= 0; i-- {
			cleanupFns[i]()
		}
	}
	clients := &GatewayClients{}
	if services != nil && services.Auth != nil && services.Auth.GrpcEndpoint != "" {
		conn, err := grpc.NewClient(services.Auth.GrpcEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			cleanup()
			return nil, nil, err
		}
		cleanupFns = append(cleanupFns, func() { _ = conn.Close() })
		clients.Auth = authv1.NewAuthServiceClient(conn)
	}
	if services != nil && services.Admin != nil && services.Admin.GrpcEndpoint != "" {
		conn, err := grpc.NewClient(services.Admin.GrpcEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			cleanup()
			return nil, nil, err
		}
		cleanupFns = append(cleanupFns, func() { _ = conn.Close() })
		clients.Admin = adminv1.NewAdminServiceClient(conn)
	}
	return clients, cleanup, nil
}
