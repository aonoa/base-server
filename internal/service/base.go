package service

import (
	"base-server/api/gen/go/base_api/v1"
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
	v1.UnimplementedBaseServer
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
	return nil, s.uc.ReLoadPolicy(ctx)
}

func (s *BaseService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	//req.Password = tools.UserPasswdEncrypt(req.Password, "")
	// 检查是否有这个人
	uid, err := s.uc.Login(ctx, req)
	if err != nil || uid == uuid.Nil.String() {
		return nil, err
	}

	return s.uc.GenerateToken(ctx, uid, s.key)
}
func (s *BaseService) GetUserInfo(ctx context.Context, req *emptypb.Empty) (*v1.GetUserInfoReply, error) {
	uid := ""
	if claims, ok := jwt.FromContext(ctx); ok {
		uid = (*claims.(*jwtv5.MapClaims))["user_id"].(string)
	}
	user, err := s.uc.GetUserInfo(ctx, uid)
	if err != nil {
		return nil, err
	}
	var res v1.GetUserInfoReply
	copier.Copy(&res, user)
	res.UserId = user.ID.String()
	//res.HomePath = "Dashboard"
	// /pages
	return &res, nil
}
func (s *BaseService) GetAccessCodes(ctx context.Context, req *emptypb.Empty) (*v1.GetAccessCodesReply, error) {
	uid := tools.GetUserId(ctx)
	user, err := s.uc.GetUserInfo(ctx, uid)
	if err != nil {
		return nil, err
	}
	var res v1.GetUserInfoReply
	copier.Copy(&res, user)
	res.UserId = user.ID.String()
	accessCodeList := strings.Split(user.Extension, ",")
	return &v1.GetAccessCodesReply{AccessCodeList: accessCodeList}, nil
}
func (s *BaseService) Logout(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	// 仅靠jwt无法实现退出功能
	return &emptypb.Empty{}, nil
}

// GetMenuList 获取路由菜单
func (s *BaseService) GetMenuList(ctx context.Context, req *emptypb.Empty) (*v1.GetSysMenuListReply, error) { // *pb.GetMenuListReply
	return s.uc.CreateMenuTree(ctx)
}
func (s *BaseService) RefreshToken(ctx context.Context, req *emptypb.Empty) (*v1.LoginReply, error) {
	uid := ""
	aud := ""
	if claims, ok := jwt.FromContext(ctx); ok {
		uid = (*claims.(*jwtv5.MapClaims))["user_id"].(string)
		aud = (*claims.(*jwtv5.MapClaims))["aud"].(string)
	}

	if aud == "refresh" {
		return s.uc.GenerateToken(ctx, uid, s.key)
	}

	return &v1.LoginReply{}, nil
}

/////////////////////////

func (s *BaseService) GetUserList(ctx context.Context, req *v1.GetUserParams) (*v1.GetUserListReply, error) {
	return s.uc.GetUserList(ctx, req)
}

func (s *BaseService) AddUser(ctx context.Context, req *v1.UserListItem) (*v1.UserListItem, error) {
	return s.uc.AddUser(ctx, req)
}

func (s *BaseService) UpdateUser(ctx context.Context, req *v1.UserListItem) (*v1.UserListItem, error) {
	err := s.uc.UpdateUser(ctx, req)
	if err != nil {
		return nil, err
	}
	return &v1.UserListItem{}, nil
}

func (s *BaseService) DelUser(ctx context.Context, req *v1.DeleteUser) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.DelUser(ctx, req.Id)
}

func (s *BaseService) IsUserExist(ctx context.Context, req *v1.IsUserExistsRequest) (*v1.IsUserExistsReply, error) {
	res := &v1.IsUserExistsReply{}
	if ok, err := s.uc.IsUserExist(ctx, req); err == nil {
		res.Data = ok
	}
	return res, nil
}

////////////////////////////////////////

func (s *BaseService) GetDeptList(ctx context.Context, req *emptypb.Empty) (*v1.GetDeptListReply, error) {
	return s.uc.CreateDeptTree(ctx)
}
func (s *BaseService) AddDept(ctx context.Context, req *v1.DeptListItem) (*v1.DeptListItem, error) {
	err := s.uc.AddDept(ctx, req)
	if err != nil {
		return nil, err
	}
	return &v1.DeptListItem{}, nil
}
func (s *BaseService) UpdateDept(ctx context.Context, req *v1.DeptListItem) (*v1.DeptListItem, error) {
	err := s.uc.UpdateDept(ctx, req)
	if err != nil {
		return nil, err
	}
	return &v1.DeptListItem{}, nil
}
func (s *BaseService) DelDept(ctx context.Context, req *v1.DeleteDept) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.DelDept(ctx, req.Id)
}

func (s *BaseService) AddRole(ctx context.Context, req *v1.RoleListItem) (*v1.RoleListItem, error) {
	err := s.uc.AddRole(ctx, req)
	if err != nil {
		return nil, err
	}
	return &v1.RoleListItem{}, nil
}
func (s *BaseService) DelRole(ctx context.Context, req *v1.DeleteRole) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.DelRole(ctx, req.Id)
}
func (s *BaseService) UpdateRole(ctx context.Context, req *v1.RoleListItem) (*v1.RoleListItem, error) {
	err := s.uc.UpdateRole(ctx, req)
	if err != nil {
		return nil, err
	}
	return &v1.RoleListItem{}, nil
}

///////////////////////////////////////////////////// 系统菜单管理

func (s *BaseService) GetSysMenuList(ctx context.Context, req *v1.MenuParams) (*v1.GetSysMenuListReply, error) {
	return s.uc.GetSysMenuList(ctx)
}

func (s *BaseService) IsMenuNameExists(ctx context.Context, req *v1.IsMenuNameExistsRequest) (*v1.IsMenuNameExistsReply, error) {
	res := &v1.IsMenuNameExistsReply{}
	if ok, err := s.uc.IsMenuNameExists(ctx, req); err == nil {
		res.Data = ok
	}
	return res, nil
}
func (s *BaseService) IsMenuPathExists(ctx context.Context, req *v1.IsMenuPathExistsRequest) (*v1.IsMenuPathExistsReply, error) {
	res := &v1.IsMenuPathExistsReply{}
	if ok, err := s.uc.IsMenuPathExists(ctx, req); err == nil {
		res.Data = ok
	}
	return res, nil
}
func (s *BaseService) CreateMenu(ctx context.Context, req *v1.SysMenuListItem) (*emptypb.Empty, error) {
	return s.uc.CreateMenu(ctx, req)
}
func (s *BaseService) UpdateMenu(ctx context.Context, req *v1.SysMenuListItem) (*emptypb.Empty, error) {
	return s.uc.UpdateMenu(ctx, req)
}
func (s *BaseService) DeleteMenu(ctx context.Context, req *v1.DeleteMenuRequest) (*emptypb.Empty, error) {
	return s.uc.DeleteMenu(ctx, req)
}

/////////////////////////////////////////////////////

func (s *BaseService) GetRoleList(ctx context.Context, req *v1.RolePageParams) (*v1.GetRoleListByPageReply, error) {
	return s.uc.GetAllRoleList(ctx, req)
}

func (s *BaseService) SetRoleStatus(ctx context.Context, req *v1.SetRoleStatusRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *BaseService) ChangePassword(ctx context.Context, req *v1.ChangePasswordRequest) (*emptypb.Empty, error) {
	uid := ""
	if claims, ok := jwt.FromContext(ctx); ok {
		uid = (*claims.(*jwtv5.MapClaims))["user_id"].(string)
	}
	return &emptypb.Empty{}, s.uc.ChangePassword(ctx, uid, req.PasswordOld, req.PasswordNew)
}

func (s *BaseService) UploadFileHttp(ctx context.Context, reqFile *v1.File, opts ...grpc.CallOption) (*v1.UploadResponse, error) {
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

	return &v1.UploadResponse{
		FileInfoId: "456465",
		FullUrl:    "aaa",
		Url:        "https://q1.qlogo.cn/g?b=qq&nk=190848757&s=640",
	}, nil
}
