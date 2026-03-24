package biz

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	v1 "base-server/api/gen/go/user/service/v1"
	"base-server/pkg/data/ent"
	"github.com/google/uuid"
	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewUserUsecase)

type UserRepo interface {
	FindUserByID(context.Context, *uuid.UUID) (*ent.User, error)
	GetUserList(context.Context, *v1.GetUserParams) ([]*ent.User, int64, error)
	AddUser(context.Context, *v1.UserListItem) (*ent.User, error)
	UpdateUser(context.Context, *uuid.UUID, *v1.UserListItem) (*ent.User, error)
	DeleteByID(context.Context, *uuid.UUID) error
	IsUserExistsByUserName(context.Context, *v1.IsUserExistsRequest) (*ent.User, error)
	ChangePassword(context.Context, *uuid.UUID, string, string) error
	ValidateUserAuth(context.Context, string, string) (*ent.User, error)
	ListUserAuthBindings(context.Context) ([]*ent.User, error)
}

type UserUsecase struct {
	repo UserRepo
}

func NewUserUsecase(repo UserRepo) *UserUsecase {
	return &UserUsecase{repo: repo}
}

func (uc *UserUsecase) GetUserInfo(ctx context.Context, userID string) (*v1.GetUserInfoReply, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	user, err := uc.repo.FindUserByID(ctx, &uid)
	if err != nil {
		return nil, err
	}
	res := &v1.GetUserInfoReply{
		UserId:   user.ID.String(),
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Remark:   user.Desc,
		Status:   int32(user.Status),
	}
	var extension struct {
		Email string `json:"email"`
	}
	_ = json.Unmarshal([]byte(user.Extension), &extension)
	res.Email = extension.Email
	if user.RoleID != nil {
		res.Roles = append(res.Roles, &v1.RoleInfo{Id: *user.RoleID})
	}
	return res, nil
}

func (uc *UserUsecase) GetUserList(ctx context.Context, req *v1.GetUserParams) (*v1.GetUserListReply, error) {
	userList, count, err := uc.repo.GetUserList(ctx, req)
	if err != nil {
		return nil, err
	}
	res := &v1.GetUserListReply{Total: count}
	for _, user := range userList {
		var extension struct {
			Email string `json:"email"`
		}
		_ = json.Unmarshal([]byte(user.Extension), &extension)
		roleID := int64(0)
		if user.RoleID != nil {
			roleID = *user.RoleID
		}
		res.Items = append(res.Items, &v1.UserListItem{
			Id:         user.ID.String(),
			Username:   user.Username,
			Email:      extension.Email,
			Nickname:   user.Nickname,
			Role:       roleID,
			CreateTime: user.CreateTime.Format(time.DateTime),
			Remark:     user.Desc,
			Status:     int32(user.Status),
			Avatar:     user.Avatar,
		})
	}
	return res, nil
}

func (uc *UserUsecase) AddUser(ctx context.Context, req *v1.UserListItem) (*v1.UserListItem, error) {
	user, err := uc.repo.AddUser(ctx, req)
	if err != nil {
		return nil, err
	}
	return &v1.UserListItem{
		Id:         user.ID.String(),
		Username:   user.Username,
		Nickname:   user.Nickname,
		Remark:     user.Desc,
		Status:     int32(user.Status),
		Avatar:     user.Avatar,
		CreateTime: user.CreateTime.Format(time.DateTime),
		Role:       req.Role,
	}, nil
}

func (uc *UserUsecase) UpdateUser(ctx context.Context, req *v1.UserListItem) error {
	uid, err := uuid.Parse(req.Id)
	if err != nil {
		return err
	}
	_, err = uc.repo.UpdateUser(ctx, &uid, req)
	return err
}

func (uc *UserUsecase) DelUser(ctx context.Context, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	return uc.repo.DeleteByID(ctx, &uid)
}

func (uc *UserUsecase) IsUserExist(ctx context.Context, req *v1.IsUserExistsRequest) (bool, error) {
	user, err := uc.repo.IsUserExistsByUserName(ctx, req)
	if err != nil {
		return true, err
	}
	if user != nil && req.Id != user.ID.String() {
		return true, nil
	}
	return false, nil
}

func (uc *UserUsecase) ChangePassword(ctx context.Context, userID, passwordOld, passwordNew string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	return uc.repo.ChangePassword(ctx, &uid, passwordOld, passwordNew)
}

func (uc *UserUsecase) ValidateUserAuth(ctx context.Context, username, password string) (*v1.ValidateUserAuthReply, error) {
	user, err := uc.repo.ValidateUserAuth(ctx, username, password)
	if err != nil {
		return nil, err
	}
	return &v1.ValidateUserAuthReply{UserId: user.ID.String()}, nil
}

func (uc *UserUsecase) GetUserAuthInfo(ctx context.Context, userID string) (*v1.GetUserAuthInfoReply, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	user, err := uc.repo.FindUserByID(ctx, &uid)
	if err != nil {
		return nil, err
	}
	res := &v1.GetUserAuthInfoReply{
		UserId:   user.ID.String(),
		Username: user.Username,
	}
	if user.RoleID != nil {
		res.RoleId = *user.RoleID
	}
	for _, item := range parseAccessCodes(user.Extension) {
		res.AccessCodes = append(res.AccessCodes, item)
	}
	return res, nil
}

func (uc *UserUsecase) ListUserAuthBindings(ctx context.Context) (*v1.ListUserAuthBindingsReply, error) {
	users, err := uc.repo.ListUserAuthBindings(ctx)
	if err != nil {
		return nil, err
	}
	res := &v1.ListUserAuthBindingsReply{}
	for _, user := range users {
		roleID := int64(0)
		if user.RoleID != nil {
			roleID = *user.RoleID
		}
		res.Items = append(res.Items, &v1.UserAuthBinding{
			UserId: user.ID.String(),
			RoleId: roleID,
		})
	}
	return res, nil
}

func parseAccessCodes(extension string) []string {
	var payload struct {
		AccessCodes []string `json:"accessCodes"`
	}
	if err := json.Unmarshal([]byte(extension), &payload); err == nil && len(payload.AccessCodes) > 0 {
		return payload.AccessCodes
	}
	items := make([]string, 0)
	for _, item := range strings.Split(extension, ",") {
		item = strings.TrimSpace(item)
		if item != "" && !strings.Contains(item, "{") && !strings.Contains(item, "}") && !strings.Contains(item, ":") {
			items = append(items, item)
		}
	}
	return items
}
