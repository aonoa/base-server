package service

import (
	pb "base-server/api/base_api/v1"
	"base-server/internal/biz"
	"base-server/internal/conf"
	"base-server/internal/tools"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"os"
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

func (s *BaseService) ReLoadPolicy(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *BaseService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	//req.Password = tools.UserPasswdEncrypt(req.Password, "")
	// 检查是否有这个人
	uid, err := s.uc.Login(ctx, req)
	if err != nil || uid == uuid.Nil.String() {
		return nil, err
	}

	return s.uc.GenerateToken(ctx, uid, s.key)
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

func (s *BaseService) GetAccountList(ctx context.Context, req *pb.AccountParams) (*pb.GetAccountListReply, error) {
	return s.uc.GetAccountList(ctx, req)
}

func (s *BaseService) AddUser(ctx context.Context, req *pb.AccountListItem) (*pb.AccountListItem, error) {
	return s.uc.AddUser(ctx, req)
}

func (s *BaseService) DelUser(ctx context.Context, req *pb.DeleteUser) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.DelUser(ctx, req.Id)
}

func (s *BaseService) GetRoleListByPage(ctx context.Context, req *pb.RolePageParams) (*pb.GetRoleListByPageReply, error) {
	return s.uc.GetAllRoleList(ctx, req)
}

////////////////////////////////////////

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

func (s *BaseService) UploadFileHttp(ctx context.Context, reqFile *pb.File, opts ...grpc.CallOption) (*pb.UploadResponse, error) {
	log.Infof("文件:%s,大小:%d", reqFile.FileName, reqFile.FileSize)
	if reqFile.FileSize <= 0 {
		reqFile.FileSize = int64(len(reqFile.File))
	}
	fileName := fmt.Sprintf("%s", reqFile.FileName)
	if fileName == "" {
		fileName = "aaa.png"
	}
	if reqFile.FileSize <= 0 {
		reqFile.FileSize = int64(len(reqFile.File))
	}

	//err := s.minio.Save(bucket, fileName, bytes.NewReader(reqFile.File), reqFile.FileSize) // 文件上传到minio
	//if err != nil {
	//	log.Error(err.Error())
	//}
	//fileType := file.CheckFileType(reqFile.FileName)

	// 创建一个新文件用于写入
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Failed to create file: %s", err)
	}
	defer file.Close()

	// 使用 bytes.NewReader 创建一个 Reader
	reader := bytes.NewReader(reqFile.File)

	// 将 reader 的内容写入文件
	_, err = io.Copy(file, reader)
	if err != nil {
		log.Fatalf("Failed to write to file: %s", err)
	}

	// 文件写入成功
	log.Infof("File '%s' saved successfully.", fileName)

	return &pb.UploadResponse{
		FileInfoId: "456465",
		FullUrl:    "aaa",
		Url:        "https://q1.qlogo.cn/g?b=qq&nk=190848757&s=640",
	}, nil
}
