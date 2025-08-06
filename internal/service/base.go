package service

import (
	pb "base-server/api/gen/go/base_api/v1"
	"base-server/internal/biz"
	"base-server/internal/conf"
	"base-server/internal/tools"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/transport/http"
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

	RestServer *http.Server
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

// GetMenuList 获取路由菜单
func (s *BaseService) GetMenuList(ctx context.Context, req *emptypb.Empty) (*pb.GetSysMenuListReply, error) { // *pb.GetMenuListReply
	return s.uc.CreateRouteMenuTree(ctx)
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

func (s *BaseService) GetUserList(ctx context.Context, req *pb.GetUserParams) (*pb.GetUserListReply, error) {
	return s.uc.GetUserList(ctx, req)
}

func (s *BaseService) AddUser(ctx context.Context, req *pb.UserListItem) (*pb.UserListItem, error) {
	return s.uc.AddUser(ctx, req)
}

func (s *BaseService) UpdateUser(ctx context.Context, req *pb.UserListItem) (*pb.UserListItem, error) {
	err := s.uc.UpdateUser(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.UserListItem{}, nil
}

func (s *BaseService) DelUser(ctx context.Context, req *pb.DeleteUser) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.DelUser(ctx, req.Id)
}

func (s *BaseService) IsUserExist(ctx context.Context, req *pb.IsUserExistsRequest) (*pb.IsUserExistsReply, error) {
	res := &pb.IsUserExistsReply{}
	if ok, err := s.uc.IsUserExist(ctx, req); err == nil {
		res.Data = ok
	}
	return res, nil
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

///////////////////////////////////////////////////// 系统菜单管理

func (s *BaseService) GetSysMenuList(ctx context.Context, req *pb.MenuParams) (*pb.GetSysMenuListReply, error) {
	return s.uc.GetSysMenuList(ctx)
}

func (s *BaseService) IsMenuNameExists(ctx context.Context, req *pb.IsMenuNameExistsRequest) (*pb.IsMenuNameExistsReply, error) {
	res := &pb.IsMenuNameExistsReply{}
	if ok, err := s.uc.IsMenuNameExists(ctx, req); err == nil {
		res.Data = ok
	}
	return res, nil
}
func (s *BaseService) IsMenuPathExists(ctx context.Context, req *pb.IsMenuPathExistsRequest) (*pb.IsMenuPathExistsReply, error) {
	res := &pb.IsMenuPathExistsReply{}
	if ok, err := s.uc.IsMenuPathExists(ctx, req); err == nil {
		res.Data = ok
	}
	return res, nil
}
func (s *BaseService) CreateMenu(ctx context.Context, req *pb.SysMenuListItem) (*emptypb.Empty, error) {
	return s.uc.CreateMenu(ctx, req)
}
func (s *BaseService) UpdateMenu(ctx context.Context, req *pb.SysMenuListItem) (*emptypb.Empty, error) {
	return s.uc.UpdateMenu(ctx, req)
}
func (s *BaseService) DeleteMenu(ctx context.Context, req *pb.DeleteMenuRequest) (*emptypb.Empty, error) {
	return s.uc.DeleteMenu(ctx, req)
}

/////////////////////////////////////////////////////

func (s *BaseService) GetRoleList(ctx context.Context, req *pb.RolePageParams) (*pb.GetRoleListByPageReply, error) {
	return s.uc.GetAllRoleList(ctx, req)
}

func (s *BaseService) SetRoleStatus(ctx context.Context, req *pb.SetRoleStatusRequest) (*emptypb.Empty, error) {
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

//////////////////////////////////////////////////

func (s *BaseService) GetWalkRoute(ctx context.Context, req *emptypb.Empty) (*pb.GetWalkRouteReply, error) {
	if s.RestServer == nil {
		return nil, fmt.Errorf("RestServer is nil")
	}

	res := &pb.GetWalkRouteReply{
		Items: []*pb.WalkRouteItem{},
	}

	var count uint32 = 0
	if err := s.RestServer.WalkRoute(func(info http.RouteInfo) error {
		//log.Infof("Path[%s] Method[%s]", info.Path, info.Method)
		count++
		res.Items = append(res.Items, &pb.WalkRouteItem{
			Url:    info.Path,
			Method: info.Method,
		})

		return nil
	}); err != nil {
		log.Errorf("failed to walk route: %v", err)
	}

	return res, nil
}

func (s *BaseService) GetApiList(ctx context.Context, req *pb.GetApiPageParams) (*pb.GetApiListByPageReply, error) {
	return s.uc.GetApiList(ctx, req)
}

func (s *BaseService) AddApi(ctx context.Context, req *pb.ApiListItem) (*pb.ApiListItem, error) {
	return s.uc.AddApi(ctx, req)
}

func (s *BaseService) UpdateApi(ctx context.Context, req *pb.ApiListItem) (*pb.ApiListItem, error) {
	_, err := s.uc.UpdateApi(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.ApiListItem{}, nil
}

func (s *BaseService) DelApi(ctx context.Context, req *pb.DeleteApi) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.DelApi(ctx, req)
}

func (s *BaseService) GetResourceList(ctx context.Context, req *pb.GetResourcePageParams) (*pb.GetResourceListByPageReply, error) {
	return s.uc.GetResourceList(ctx, req)
}

func (s *BaseService) AddResource(ctx context.Context, req *pb.ResourceListItem) (*pb.ResourceListItem, error) {
	return s.uc.AddResource(ctx, req)
}

func (s *BaseService) UpdateResource(ctx context.Context, req *pb.ResourceListItem) (*pb.ResourceListItem, error) {
	_, err := s.uc.UpdateResource(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.ResourceListItem{}, nil
}

func (s *BaseService) DelResource(ctx context.Context, req *pb.DeleteResource) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.DelResource(ctx, req)
}
