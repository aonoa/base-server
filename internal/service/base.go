package service

import (
	pb "base-server/api/base_api/v1"
	"base-server/internal/biz"
	"base-server/internal/conf"
	"base-server/internal/tools"
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/emptypb"
	"strings"
)

type BaseService struct {
	pb.UnimplementedBaseServer
	uc  *biz.BaseUsecase
	key string
}

var (
	ErrLoginFailed = errors.New("login failed")
)

func NewBaseService(uc *biz.BaseUsecase, conf *conf.Auth) *BaseService {
	return &BaseService{
		uc:  uc,
		key: conf.ApiKey,
	}
}

func (s *BaseService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	//req.Password = tools.UserPasswdEncrypt(req.Password, "")
	// 检查是否有这个人
	uid, err := s.uc.Login(ctx, req)
	if err != nil || uid == "00000000-0000-0000-0000-000000000000" {
		return nil, err
	}

	return s.uc.GenerateToken(ctx, uid, s.key)

	//now := time.Now()
	//
	//// 生成accessToken
	//claims := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, jwtv5.MapClaims{
	//	"user_id": g,
	//	"sub":     g,
	//	"exp":     now.Add(30 * time.Minute).Unix(), // 过期时间（30分钟后过期）
	//	"nbf":     now.Unix(),                       // 生效时间
	//	"iat":     now.Unix(),                       // 颁发时间
	//})
	//accessToken, err := claims.SignedString([]byte(s.key))
	//if err != nil {
	//	return nil, ErrLoginFailed
	//}
	//
	//// 生成refreshToken，提前5分钟生效
	//claims = jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, jwtv5.MapClaims{
	//	"user_id": g,
	//	"sub":     g,
	//	"exp":     now.Add(60 * time.Minute).Unix(), // 过期时间（30分钟后过期）
	//	"nbf":     now.Add(25 * time.Minute).Unix(), // 生效时间
	//	"iat":     now.Unix(),                       // 颁发时间
	//	"jti":     uuid.New().String(),              // 唯一标识符，主要用来作为一次性 token，从而回避重放（replay）攻击
	//})
	//refreshToken, err := claims.SignedString([]byte(s.key))
	//if err != nil {
	//	return nil, ErrLoginFailed
	//}
	//
	//return &pb.LoginReply{
	//	UserId:       g,
	//	AccessToken:  accessToken,
	//	RefreshToken: refreshToken,
	//}, nil
}
func (s *BaseService) GetUserInfo(ctx context.Context, req *emptypb.Empty) (*pb.GetUserInfoReply, error) {
	uid := ""
	if claims, ok := jwt.FromContext(ctx); ok {
		uid = (*claims.(*jwtv5.MapClaims))["user_id"].(string)
	}
	user, err := s.uc.GetUserInfo(ctx, uid)
	if err != nil {
		return nil, err
	}
	var res pb.GetUserInfoReply
	copier.Copy(&res, user)
	res.UserId = user.ID.String()
	//res.HomePath = "Dashboard"
	// /pages
	return &res, nil
}
func (s *BaseService) GetAccessCodes(ctx context.Context, req *emptypb.Empty) (*pb.GetAccessCodesReply, error) {
	uid := tools.GetUserId(ctx)
	user, err := s.uc.GetUserInfo(ctx, uid)
	if err != nil {
		return nil, err
	}
	var res pb.GetUserInfoReply
	copier.Copy(&res, user)
	res.UserId = user.ID.String()
	accessCodeList := strings.Split(user.Extension, ",")
	return &pb.GetAccessCodesReply{AccessCodeList: accessCodeList}, nil
}
func (s *BaseService) Logout(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	// 仅靠jwt无法实现退出功能
	return &emptypb.Empty{}, nil
}
func (s *BaseService) GetMenuList(ctx context.Context, req *emptypb.Empty) (*pb.GetMenuListReply, error) {
	return s.uc.CreateMenuTree(ctx)
}
func (s *BaseService) RefreshToken(ctx context.Context, req *emptypb.Empty) (*pb.LoginReply, error) {
	uid := ""
	aud := ""
	if claims, ok := jwt.FromContext(ctx); ok {
		uid = (*claims.(*jwtv5.MapClaims))["user_id"].(string)
		aud = (*claims.(*jwtv5.MapClaims))["aud"].(string)
	}

	if aud == "refresh" {
		return s.uc.GenerateToken(ctx, uid, s.key)
	}

	return &pb.LoginReply{}, nil
}

/////////////////////////

func (s *BaseService) GetDeptList(ctx context.Context, req *emptypb.Empty) (*pb.GetDeptListReply, error) {
	return s.uc.CreateDeptTree(ctx)
}
func (s *BaseService) AddDept(ctx context.Context, req *pb.DeptListItem) (*pb.DeptListItem, error) {
	err := s.uc.AddDept(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.DeptListItem{}, nil
}
func (s *BaseService) UpdateDept(ctx context.Context, req *pb.DeptListItem) (*pb.DeptListItem, error) {
	err := s.uc.UpdateDept(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.DeptListItem{}, nil
}
func (s *BaseService) DelDept(ctx context.Context, req *pb.DeleteDept) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.DelDept(ctx, req.Id)
}
func (s *BaseService) GetRoleListByPage(ctx context.Context, req *pb.RolePageParams) (*pb.GetRoleListByPageReply, error) {
	return s.uc.GetAllRoleList(ctx, req)
}
func (s *BaseService) AddRole(ctx context.Context, req *pb.RoleListItem) (*pb.RoleListItem, error) {
	err := s.uc.AddRole(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.RoleListItem{}, nil
}
func (s *BaseService) DelRole(ctx context.Context, req *pb.DeleteRole) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.DelRole(ctx, req.Id)
}
func (s *BaseService) UpdateRole(ctx context.Context, req *pb.RoleListItem) (*pb.RoleListItem, error) {
	err := s.uc.UpdateRole(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.RoleListItem{}, nil
}
func (s *BaseService) GetSysMenuList(ctx context.Context, req *pb.MenuParams) (*pb.GetSysMenuListReply, error) {
	return s.uc.GetSysMenuList(ctx)
}
func (s *BaseService) GetAccountList(ctx context.Context, req *pb.AccountParams) (*pb.GetAccountListReply, error) {
	return s.uc.GetAccountList(ctx, req)
}
func (s *BaseService) AddUser(ctx context.Context, req *pb.AccountListItem) (*pb.AccountListItem, error) {
	return s.uc.AddUser(ctx, req)
}
func (s *BaseService) DelUser(ctx context.Context, req *pb.DeleteUser) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.DelUser(ctx, req.Id)
}
func (s *BaseService) GetAllRoleList(ctx context.Context, req *pb.RoleParams) (*pb.GetRoleListByPageReply, error) {
	return s.uc.GetAllRoleList(ctx, &pb.RolePageParams{
		RoleNme: req.RoleName,
		Status:  req.Status,
		DeptId:  "",
	})
}
func (s *BaseService) SetRoleStatus(ctx context.Context, req *pb.SetRoleStatusRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
func (s *BaseService) IsAccountExist(ctx context.Context, req *pb.IsAccountRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
func (s *BaseService) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*emptypb.Empty, error) {
	uid := ""
	if claims, ok := jwt.FromContext(ctx); ok {
		uid = (*claims.(*jwtv5.MapClaims))["user_id"].(string)
	}
	return &emptypb.Empty{}, s.uc.ChangePassword(ctx, uid, req.PasswordOld, req.PasswordNew)
}
