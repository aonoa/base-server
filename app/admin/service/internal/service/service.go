package service

import (
	"context"

	v1 "base-server/api/gen/go/admin/service/v1"
	"base-server/app/admin/service/internal/biz"
	"base-server/pkg/authx"
	"base-server/pkg/tools"

	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewAdminService)

type AdminService struct {
	v1.UnimplementedAdminServiceServer

	uc         *biz.AdminUsecase
	RestServer *kratoshttp.Server
}

func NewAdminService(uc *biz.AdminUsecase) *AdminService {
	return &AdminService{uc: uc}
}

func (s *AdminService) GetDeptList(ctx context.Context, req *emptypb.Empty) (*v1.GetDeptListReply, error) {
	return s.uc.GetDeptList(ctx)
}

func (s *AdminService) AddDept(ctx context.Context, req *v1.DeptListItem) (*v1.DeptListItem, error) {
	return s.uc.AddDept(ctx, req)
}

func (s *AdminService) UpdateDept(ctx context.Context, req *v1.DeptListItem) (*v1.DeptListItem, error) {
	return s.uc.UpdateDept(ctx, req)
}

func (s *AdminService) DelDept(ctx context.Context, req *v1.DeleteDept) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.DelDept(ctx, req.Id)
}

func (s *AdminService) GetCurrentUserMenus(ctx context.Context, req *emptypb.Empty) (*v1.GetCurrentUserMenusReply, error) {
	return s.uc.GetCurrentUserMenus(ctx, authx.UserID(ctx))
}

func (s *AdminService) GetSysMenuList(ctx context.Context, req *v1.MenuParams) (*v1.GetSysMenuListReply, error) {
	return s.uc.GetSysMenuList(ctx)
}

func (s *AdminService) ListMenus(ctx context.Context, req *emptypb.Empty) (*v1.ListMenusReply, error) {
	return s.uc.ListMenus(ctx)
}

func (s *AdminService) GetWalkRoute(ctx context.Context, req *emptypb.Empty) (*v1.GetWalkRouteReply, error) {
	return s.uc.GetWalkRoute(ctx)
}

func (s *AdminService) GetSelfWalkRoute(ctx context.Context, req *emptypb.Empty) (*v1.GetWalkRouteReply, error) {
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

func (s *AdminService) IsMenuNameExists(ctx context.Context, req *v1.IsMenuNameExistsRequest) (*v1.IsMenuNameExistsReply, error) {
	ok, err := s.uc.IsMenuNameExists(ctx, req)
	if err != nil {
		return nil, err
	}
	return &v1.IsMenuNameExistsReply{Data: ok}, nil
}

func (s *AdminService) IsMenuPathExists(ctx context.Context, req *v1.IsMenuPathExistsRequest) (*v1.IsMenuPathExistsReply, error) {
	ok, err := s.uc.IsMenuPathExists(ctx, req)
	if err != nil {
		return nil, err
	}
	return &v1.IsMenuPathExistsReply{Data: ok}, nil
}

func (s *AdminService) CreateMenu(ctx context.Context, req *v1.SysMenuListItem) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.CreateMenu(ctx, req)
}

func (s *AdminService) UpdateMenu(ctx context.Context, req *v1.SysMenuListItem) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.UpdateMenu(ctx, req)
}

func (s *AdminService) DeleteMenu(ctx context.Context, req *v1.DeleteMenuRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.DeleteMenu(ctx, req)
}

func (s *AdminService) CreateSysLog(ctx context.Context, req *v1.CreateSysLogRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.uc.CreateSysLog(ctx, req)
}

func (s *AdminService) GetSysLogList(ctx context.Context, req *v1.GetSysLogListParams) (*v1.GetSysLogListReply, error) {
	return s.uc.GetSysLogList(ctx, req)
}

func (s *AdminService) GetSysLogInfo(ctx context.Context, req *v1.GetSysLogInfoParams) (*v1.GetSysLogInfoReply, error) {
	return s.uc.GetSysLogInfo(ctx, req.Id)
}
