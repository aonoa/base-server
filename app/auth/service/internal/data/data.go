package data

import (
	"context"

	authv1 "base-server/api/gen/go/auth/service/v1"
	userv1 "base-server/api/gen/go/user/service/v1"
	"base-server/app/auth/service/internal/biz"
	"base-server/app/auth/service/internal/conf"
	"base-server/pkg/authx"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewAuthRepo)

// Data .
type Data struct {
	userConn   *grpc.ClientConn
	userClient userv1.UserServiceClient
}

// NewData .
func NewData(c *conf.Data, services *conf.Services, logger log.Logger) (*Data, func(), error) {
	helper := log.NewHelper(logger)
	userConn, err := grpc.NewClient(services.User.GrpcEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	d := &Data{
		userConn:   userConn,
		userClient: userv1.NewUserServiceClient(userConn),
	}
	cleanup := func() {
		helper.Info("closing the data resources")
		if d.userConn != nil {
			_ = d.userConn.Close()
		}
	}
	return d, cleanup, nil
}

type authRepo struct {
	data *Data
	log  *log.Helper
}

func NewAuthRepo(data *Data, logger log.Logger) biz.AuthRepo {
	return &authRepo{data: data, log: log.NewHelper(logger)}
}

func (r *authRepo) Login(ctx context.Context, req *authv1.LoginRequest) (string, error) {
	ctx = authx.ForwardAuthorizationContext(ctx)
	res, err := r.data.userClient.ValidateUserAuth(ctx, &userv1.ValidateUserAuthRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return "", err
	}
	return res.UserId, nil
}

func (r *authRepo) GetUserAuthInfo(ctx context.Context, userID string) (*userv1.GetUserAuthInfoReply, error) {
	ctx = authx.ForwardAuthorizationContext(ctx)
	return r.data.userClient.GetUserAuthInfo(ctx, &userv1.GetUserAuthInfoRequest{UserId: userID})
}
