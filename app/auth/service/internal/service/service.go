package service

import (
	"context"

	v1 "base-server/api/gen/go/auth/service/v1"
	"base-server/app/auth/service/internal/biz"
	"base-server/pkg/authx"
	"base-server/pkg/tools"

	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewAuthService)

type AuthService struct {
	v1.UnimplementedAuthServiceServer

	uc         *biz.AuthUsecase
	RestServer *kratoshttp.Server
}

func NewAuthService(uc *biz.AuthUsecase) *AuthService {
	return &AuthService{uc: uc}
}

func (s *AuthService) RegisterPermissionSnapshot(ctx context.Context, req *v1.RegisterPermissionSnapshotRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.RegisterPermissionSnapshot(ctx, req)
}

func (s *AuthService) ApplyRoleDelta(ctx context.Context, req *v1.ApplyRoleDeltaRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.ApplyRoleDelta(ctx, req)
}

func (s *AuthService) ApplyApiDelta(ctx context.Context, req *v1.ApplyApiDeltaRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.ApplyApiDelta(ctx, req)
}

func (s *AuthService) ApplyUserRoleBindingDelta(ctx context.Context, req *v1.ApplyUserRoleBindingDeltaRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.ApplyUserRoleBindingDelta(ctx, req)
}

func (s *AuthService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	return s.uc.Login(ctx, req)
}

func (s *AuthService) GetAccessCodes(ctx context.Context, req *emptypb.Empty) (*v1.GetAccessCodesReply, error) {
	return s.uc.GetAccessCodes(ctx, authx.UserID(ctx))
}

func (s *AuthService) CheckAuthorization(ctx context.Context, req *v1.CheckAuthorizationRequest) (*v1.CheckAuthorizationReply, error) {
	return s.uc.CheckAuthorization(ctx, req)
}

func (s *AuthService) GetWalkRoute(ctx context.Context, req *emptypb.Empty) (*v1.GetWalkRouteReply, error) {
	items, err := tools.WalkHTTPRoutes(s.RestServer)
	if err != nil {
		return nil, err
	}
	res := &v1.GetWalkRouteReply{Items: make([]*v1.WalkRouteItem, 0, len(items))}
	for _, item := range tools.SortAndUniqueWalkRoutes(items) {
		res.Items = append(res.Items, &v1.WalkRouteItem{Url: item.URL, Method: item.Method})
	}
	return res, nil
}

func (s *AuthService) Logout(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *emptypb.Empty) (*v1.LoginReply, error) {
	return s.uc.RefreshToken(ctx, authx.UserID(ctx), authx.Audience(ctx), authx.SessionID(ctx))
}
