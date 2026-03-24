package service

import (
	"context"

	v1 "base-server/api/gen/go/user/service/v1"
	"base-server/app/user/service/internal/biz"

	"github.com/google/wire"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewUserService)

type UserService struct {
	v1.UnimplementedUserServiceServer

	uc *biz.UserUsecase
}

func NewUserService(uc *biz.UserUsecase) *UserService {
	return &UserService{uc: uc}
}

func (s *UserService) GetUserInfo(ctx context.Context, req *v1.GetUserInfoRequest) (*v1.GetUserInfoReply, error) {
	return s.uc.GetUserInfo(ctx, req.UserId)
}

func (s *UserService) GetUserList(ctx context.Context, req *v1.GetUserParams) (*v1.GetUserListReply, error) {
	return s.uc.GetUserList(ctx, req)
}

func (s *UserService) AddUser(ctx context.Context, req *v1.UserListItem) (*v1.UserListItem, error) {
	return s.uc.AddUser(ctx, req)
}

func (s *UserService) UpdateUser(ctx context.Context, req *v1.UserListItem) (*v1.UserListItem, error) {
	if err := s.uc.UpdateUser(ctx, req); err != nil {
		return nil, err
	}
	return &v1.UserListItem{}, nil
}

func (s *UserService) DelUser(ctx context.Context, req *v1.DeleteUser) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.DelUser(ctx, req.Id)
}

func (s *UserService) IsUserExist(ctx context.Context, req *v1.IsUserExistsRequest) (*v1.IsUserExistsReply, error) {
	ok, err := s.uc.IsUserExist(ctx, req)
	if err != nil {
		return nil, err
	}
	return &v1.IsUserExistsReply{Data: ok}, nil
}

func (s *UserService) ChangePassword(ctx context.Context, req *v1.ChangePasswordRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.ChangePassword(ctx, req.UserId, req.PasswordOld, req.PasswordNew)
}

func (s *UserService) ValidateUserAuth(ctx context.Context, req *v1.ValidateUserAuthRequest) (*v1.ValidateUserAuthReply, error) {
	return s.uc.ValidateUserAuth(ctx, req.Username, req.Password)
}

func (s *UserService) GetUserAuthInfo(ctx context.Context, req *v1.GetUserAuthInfoRequest) (*v1.GetUserAuthInfoReply, error) {
	return s.uc.GetUserAuthInfo(ctx, req.UserId)
}

func (s *UserService) ListUserAuthBindings(ctx context.Context, req *emptypb.Empty) (*v1.ListUserAuthBindingsReply, error) {
	return s.uc.ListUserAuthBindings(ctx)
}
