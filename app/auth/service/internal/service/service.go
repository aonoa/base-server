package service

import (
	"context"

	v1 "base-server/api/gen/go/auth/service/v1"
	"base-server/app/auth/service/internal/biz"
	"base-server/pkg/authx"

	"github.com/google/wire"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewAuthService)

type AuthService struct {
	v1.UnimplementedAuthServiceServer

	uc *biz.AuthUsecase
}

func NewAuthService(uc *biz.AuthUsecase) *AuthService {
	return &AuthService{uc: uc}
}

func (s *AuthService) ReLoadPolicy(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.ReLoadPolicy(ctx)
}

func (s *AuthService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	return s.uc.Login(ctx, req)
}

func (s *AuthService) GetAccessCodes(ctx context.Context, req *emptypb.Empty) (*v1.GetAccessCodesReply, error) {
	return s.uc.GetAccessCodes(ctx, authx.UserID(ctx))
}

func (s *AuthService) Logout(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *emptypb.Empty) (*v1.LoginReply, error) {
	return s.uc.RefreshToken(ctx, authx.UserID(ctx), authx.Audience(ctx), authx.SessionID(ctx))
}

func (s *AuthService) GetRoleList(ctx context.Context, req *v1.RolePageParams) (*v1.GetRoleListByPageReply, error) {
	return s.uc.GetRoleList(ctx, req)
}

func (s *AuthService) AddRole(ctx context.Context, req *v1.RoleListItem) (*v1.RoleListItem, error) {
	return s.uc.AddRole(ctx, req)
}

func (s *AuthService) UpdateRole(ctx context.Context, req *v1.RoleListItem) (*v1.RoleListItem, error) {
	return s.uc.UpdateRole(ctx, req)
}

func (s *AuthService) DelRole(ctx context.Context, req *v1.DeleteRole) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.DelRole(ctx, req.Id)
}

func (s *AuthService) GetApiList(ctx context.Context, req *v1.GetApiPageParams) (*v1.GetApiListByPageReply, error) {
	return s.uc.GetApiList(ctx, req)
}

func (s *AuthService) AddApi(ctx context.Context, req *v1.ApiListItem) (*v1.ApiListItem, error) {
	return s.uc.AddApi(ctx, req)
}

func (s *AuthService) UpdateApi(ctx context.Context, req *v1.ApiListItem) (*v1.ApiListItem, error) {
	return s.uc.UpdateApi(ctx, req)
}

func (s *AuthService) DelApi(ctx context.Context, req *v1.DeleteApi) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.DelApi(ctx, req.Id)
}

func (s *AuthService) GetResourceList(ctx context.Context, req *v1.GetResourcePageParams) (*v1.GetResourceListByPageReply, error) {
	return s.uc.GetResourceList(ctx, req)
}

func (s *AuthService) AddResource(ctx context.Context, req *v1.ResourceListItem) (*v1.ResourceListItem, error) {
	return s.uc.AddResource(ctx, req)
}

func (s *AuthService) UpdateResource(ctx context.Context, req *v1.ResourceListItem) (*v1.ResourceListItem, error) {
	return s.uc.UpdateResource(ctx, req)
}

func (s *AuthService) DelResource(ctx context.Context, req *v1.DeleteResource) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.DelResource(ctx, req.Id)
}
