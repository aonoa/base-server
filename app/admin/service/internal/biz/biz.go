package biz

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	v1 "base-server/api/gen/go/admin/service/v1"
	userv1 "base-server/api/gen/go/user/service/v1"
	"base-server/pkg/authx"
	"base-server/pkg/data/ent"
	"base-server/pkg/tools"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewAdminUsecase)

type AdminRepo interface {
	GetMenuList(context.Context) ([]*ent.Menu, error)
	GetUserAuthInfo(context.Context, string) (*userv1.GetUserAuthInfoReply, error)
	ListRoles(context.Context, *v1.RolePageParams) ([]*ent.Role, error)
	ListAllRoles(context.Context) ([]*ent.Role, error)
	GetRole(context.Context, int64) (*ent.Role, error)
	ResolveRoleValues(context.Context, []int64) (map[int64]string, error)
	GetUserRoleBinding(context.Context, string) (*ent.UserRoleBinding, error)
	ListUserRoleBindings(context.Context) ([]*ent.UserRoleBinding, error)
	UpsertUserRoleBinding(context.Context, string, int64) (*ent.UserRoleBinding, error)
	DeleteUserRoleBinding(context.Context, string) error
	AddRole(context.Context, *v1.RoleListItem) (*ent.Role, error)
	UpdateRole(context.Context, int64, *v1.RoleListItem) (*ent.Role, error)
	DelRole(context.Context, int64) error
	GetApiList(context.Context, *v1.GetApiPageParams) ([]*ent.ApiResources, int64, error)
	GetApi(context.Context, string) (*ent.ApiResources, error)
	AddApi(context.Context, *ent.ApiResources) (*ent.ApiResources, error)
	UpdateApi(context.Context, *ent.ApiResources) (*ent.ApiResources, error)
	DelApi(context.Context, string) error
	GetResourceList(context.Context, *v1.GetResourcePageParams) ([]*ent.Resource, int64, error)
	GetResource(context.Context, string) (*ent.Resource, error)
	AddResource(context.Context, *ent.Resource) (*ent.Resource, error)
	UpdateResource(context.Context, *ent.Resource) (*ent.Resource, error)
	DelResource(context.Context, string) error
	RegisterPermissionSnapshot(context.Context) error
	ApplyAuthRoleDelta(context.Context, *ent.Role, *ent.Role) error
	ApplyAuthApiDelta(context.Context, *ent.ApiResources, *ent.ApiResources) error
	ApplyAuthUserRoleBindingDelta(context.Context, *ent.UserRoleBinding, *ent.UserRoleBinding) error
	ListAuthWalkRoutes(context.Context) ([]tools.WalkRouteItem, error)
	ListUserWalkRoutes(context.Context) ([]tools.WalkRouteItem, error)
	ListCommonWalkRoutes(context.Context) ([]tools.WalkRouteItem, error)
	CreateMenu(context.Context, *ent.Menu) (*ent.Menu, error)
	UpdateMenu(context.Context, int64, *ent.Menu) (*ent.Menu, error)
	DeleteMenu(context.Context, int64) error
	GetDeptList(context.Context) ([]*ent.Dept, error)
	AddDept(context.Context, *v1.DeptListItem) (*ent.Dept, error)
	UpdateDept(context.Context, int64, *v1.DeptListItem) (*ent.Dept, error)
	DelDept(context.Context, int64) error
	GetDeptLeafsChildren(context.Context, int64) ([]*ent.Dept, error)
	GetDeptById(context.Context, int64) (*ent.Dept, error)
	CreateSysLog(context.Context, *ent.SysLogRecord) error
	GetSysLogList(context.Context, *v1.GetSysLogListParams) ([]*ent.SysLogRecord, int64, error)
	GetSysLogInfo(context.Context, string) (*ent.SysLogRecord, error)
}

type AdminUsecase struct {
	repo AdminRepo
	log  *log.Helper
}

type currentUserMenuAuthority struct {
	isRoot  bool
	menuIDs []int32
}

func NewAdminUsecase(repo AdminRepo, logger log.Logger) *AdminUsecase {
	return &AdminUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *AdminUsecase) RegisterPermissionSnapshot(ctx context.Context) error {
	if err := uc.repo.RegisterPermissionSnapshot(ctx); err != nil {
		return err
	}
	uc.log.Info("auth projection snapshot registered")
	return nil
}

func (uc *AdminUsecase) syncAuthProjection(ctx context.Context, label string, apply func() error) error {
	if err := apply(); err != nil {
		uc.log.Warnf("auth projection delta failed, fallback to snapshot: %s err=%v", label, err)
		if snapshotErr := uc.repo.RegisterPermissionSnapshot(ctx); snapshotErr != nil {
			return fmt.Errorf("apply auth projection update: %w; snapshot fallback failed: %v", err, snapshotErr)
		}
		uc.log.Infof("auth projection snapshot fallback succeeded: %s", label)
		return nil
	}
	uc.log.Infof("auth projection delta applied: %s", label)
	return nil
}

func (uc *AdminUsecase) loadRoleWithResources(ctx context.Context, roleID int64) (*ent.Role, error) {
	roleItem, err := uc.repo.GetRole(ctx, roleID)
	if err != nil {
		return nil, err
	}
	resources, err := roleItem.QueryResource().All(ctx)
	if err != nil {
		return nil, err
	}
	roleItem.Edges.Resource = resources
	return roleItem, nil
}

func (uc *AdminUsecase) GetAuthRoleCatalog(ctx context.Context) (*v1.GetAuthRoleCatalogReply, error) {
	roleList, err := uc.repo.ListAllRoles(ctx)
	if err != nil {
		return nil, err
	}
	res := &v1.GetAuthRoleCatalogReply{Items: make([]*v1.AuthRoleItem, 0, len(roleList))}
	for _, item := range roleList {
		res.Items = append(res.Items, roleToAuthReply(item))
	}
	return res, nil
}

func (uc *AdminUsecase) GetAuthApiCatalog(ctx context.Context) (*v1.GetAuthApiCatalogReply, error) {
	list, _, err := uc.repo.GetApiList(ctx, &v1.GetApiPageParams{})
	if err != nil {
		return nil, err
	}
	res := &v1.GetAuthApiCatalogReply{Items: make([]*v1.AuthApiItem, 0, len(list))}
	for _, item := range list {
		res.Items = append(res.Items, apiToAuthReply(item))
	}
	return res, nil
}

func (uc *AdminUsecase) GetAuthRole(ctx context.Context, req *v1.GetAuthRoleRequest) (*v1.AuthRoleItem, error) {
	roleItem, err := uc.repo.GetRole(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	resources, err := roleItem.QueryResource().All(ctx)
	if err != nil {
		return nil, err
	}
	roleItem.Edges.Resource = resources
	return roleToAuthReply(roleItem), nil
}

func (uc *AdminUsecase) ResolveRoleValues(ctx context.Context, req *v1.ResolveRoleValuesRequest) (*v1.ResolveRoleValuesReply, error) {
	values, err := uc.repo.ResolveRoleValues(ctx, req.RoleIds)
	if err != nil {
		return nil, err
	}
	res := &v1.ResolveRoleValuesReply{Items: make([]*v1.ResolveRoleValueItem, 0, len(values))}
	for _, roleID := range req.RoleIds {
		if value, ok := values[roleID]; ok {
			res.Items = append(res.Items, &v1.ResolveRoleValueItem{RoleId: roleID, Value: value})
		}
	}
	return res, nil
}

func (uc *AdminUsecase) GetUserRoleBinding(ctx context.Context, req *v1.GetUserRoleBindingRequest) (*v1.UserRoleBindingItem, error) {
	item, err := uc.repo.GetUserRoleBinding(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return userRoleBindingToReply(item), nil
}

func (uc *AdminUsecase) ListUserRoleBindings(ctx context.Context) (*v1.ListUserRoleBindingsReply, error) {
	list, err := uc.repo.ListUserRoleBindings(ctx)
	if err != nil {
		return nil, err
	}
	res := &v1.ListUserRoleBindingsReply{Items: make([]*v1.UserRoleBindingItem, 0, len(list))}
	for _, item := range list {
		res.Items = append(res.Items, userRoleBindingToReply(item))
	}
	return res, nil
}

func (uc *AdminUsecase) UpsertUserRoleBinding(ctx context.Context, req *v1.UserRoleBindingItem) (*v1.UserRoleBindingItem, error) {
	if _, err := uc.repo.GetRole(ctx, req.RoleId); err != nil {
		return nil, err
	}
	var before *ent.UserRoleBinding
	before, err := uc.repo.GetUserRoleBinding(ctx, req.UserId)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}
	if ent.IsNotFound(err) {
		before = nil
	}
	item, err := uc.repo.UpsertUserRoleBinding(ctx, req.UserId, req.RoleId)
	if err != nil {
		return nil, err
	}
	if err := uc.syncAuthProjection(ctx, fmt.Sprintf("user-role-binding upsert user=%s role_id=%d", req.UserId, req.RoleId), func() error {
		return uc.repo.ApplyAuthUserRoleBindingDelta(ctx, before, item)
	}); err != nil {
		return nil, err
	}
	return userRoleBindingToReply(item), nil
}

func (uc *AdminUsecase) DeleteUserRoleBinding(ctx context.Context, req *v1.DeleteUserRoleBindingRequest) error {
	before, err := uc.repo.GetUserRoleBinding(ctx, req.UserId)
	if err != nil {
		return err
	}
	if err := uc.repo.DeleteUserRoleBinding(ctx, req.UserId); err != nil {
		return err
	}
	return uc.syncAuthProjection(ctx, fmt.Sprintf("user-role-binding delete user=%s", req.UserId), func() error {
		return uc.repo.ApplyAuthUserRoleBindingDelta(ctx, before, nil)
	})
}

func (uc *AdminUsecase) GetRoleList(ctx context.Context, req *v1.RolePageParams) (*v1.GetRoleListByPageReply, error) {
	roleList, err := uc.repo.ListRoles(ctx, req)
	if err != nil {
		return nil, err
	}
	res := &v1.GetRoleListByPageReply{Items: make([]*v1.RoleListItem, 0, len(roleList)), Total: int64(len(roleList))}
	for i, item := range roleList {
		res.Items = append(res.Items, roleToReply(item, i))
	}
	return res, nil
}

func (uc *AdminUsecase) AddRole(ctx context.Context, req *v1.RoleListItem) (*v1.RoleListItem, error) {
	roleItem, err := uc.repo.AddRole(ctx, req)
	if err != nil {
		return nil, err
	}
	resources, err := roleItem.QueryResource().All(ctx)
	if err != nil {
		return nil, err
	}
	roleItem.Edges.Resource = resources
	if err := uc.syncAuthProjection(ctx, fmt.Sprintf("role add id=%d value=%s", roleItem.ID, roleItem.Value), func() error {
		return uc.repo.ApplyAuthRoleDelta(ctx, nil, roleItem)
	}); err != nil {
		return nil, err
	}
	return roleToReply(roleItem, 0), nil
}

func (uc *AdminUsecase) UpdateRole(ctx context.Context, req *v1.RoleListItem) (*v1.RoleListItem, error) {
	roleID, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return nil, err
	}
	before, err := uc.loadRoleWithResources(ctx, roleID)
	if err != nil {
		return nil, err
	}
	roleItem, err := uc.repo.UpdateRole(ctx, roleID, req)
	if err != nil {
		return nil, err
	}
	resources, err := roleItem.QueryResource().All(ctx)
	if err != nil {
		return nil, err
	}
	roleItem.Edges.Resource = resources
	if err := uc.syncAuthProjection(ctx, fmt.Sprintf("role update id=%d value=%s", roleItem.ID, roleItem.Value), func() error {
		return uc.repo.ApplyAuthRoleDelta(ctx, before, roleItem)
	}); err != nil {
		return nil, err
	}
	return roleToReply(roleItem, 0), nil
}

func (uc *AdminUsecase) DelRole(ctx context.Context, roleID string) error {
	id, err := strconv.ParseInt(roleID, 10, 64)
	if err != nil {
		return err
	}
	before, err := uc.loadRoleWithResources(ctx, id)
	if err != nil {
		return err
	}
	if _, err := uc.repo.GetRole(ctx, id); err != nil {
		return err
	}
	if err := uc.repo.DelRole(ctx, id); err != nil {
		return err
	}
	return uc.syncAuthProjection(ctx, fmt.Sprintf("role delete id=%d value=%s", before.ID, before.Value), func() error {
		return uc.repo.ApplyAuthRoleDelta(ctx, before, nil)
	})
}

func (uc *AdminUsecase) GetApiList(ctx context.Context, req *v1.GetApiPageParams) (*v1.GetApiListByPageReply, error) {
	list, count, err := uc.repo.GetApiList(ctx, req)
	if err != nil {
		return nil, err
	}
	res := &v1.GetApiListByPageReply{Items: make([]*v1.ApiListItem, 0, len(list)), Total: count}
	for _, item := range list {
		res.Items = append(res.Items, apiToReply(item))
	}
	return res, nil
}

func (uc *AdminUsecase) AddApi(ctx context.Context, req *v1.ApiListItem) (*v1.ApiListItem, error) {
	apiItem := &ent.ApiResources{}
	copier.Copy(apiItem, req)
	created, err := uc.repo.AddApi(ctx, apiItem)
	if err != nil {
		return nil, err
	}
	if err := uc.syncAuthProjection(ctx, fmt.Sprintf("api add id=%s path=%s group=%s", created.ID, created.Path, created.ResourcesGroup), func() error {
		return uc.repo.ApplyAuthApiDelta(ctx, nil, created)
	}); err != nil {
		return nil, err
	}
	return apiToReply(created), nil
}

func (uc *AdminUsecase) UpdateApi(ctx context.Context, req *v1.ApiListItem) (*v1.ApiListItem, error) {
	before, err := uc.repo.GetApi(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	apiItem := &ent.ApiResources{}
	copier.Copy(apiItem, req)
	updated, err := uc.repo.UpdateApi(ctx, apiItem)
	if err != nil {
		return nil, err
	}
	if err := uc.syncAuthProjection(ctx, fmt.Sprintf("api update id=%s path=%s group=%s", updated.ID, updated.Path, updated.ResourcesGroup), func() error {
		return uc.repo.ApplyAuthApiDelta(ctx, before, updated)
	}); err != nil {
		return nil, err
	}
	return apiToReply(updated), nil
}

func (uc *AdminUsecase) DelApi(ctx context.Context, apiID string) error {
	before, err := uc.repo.GetApi(ctx, apiID)
	if err != nil {
		return err
	}
	if err := uc.repo.DelApi(ctx, apiID); err != nil {
		return err
	}
	return uc.syncAuthProjection(ctx, fmt.Sprintf("api delete id=%s path=%s group=%s", before.ID, before.Path, before.ResourcesGroup), func() error {
		return uc.repo.ApplyAuthApiDelta(ctx, before, nil)
	})
}

func (uc *AdminUsecase) GetResourceList(ctx context.Context, req *v1.GetResourcePageParams) (*v1.GetResourceListByPageReply, error) {
	list, count, err := uc.repo.GetResourceList(ctx, req)
	if err != nil {
		return nil, err
	}
	res := &v1.GetResourceListByPageReply{Items: make([]*v1.ResourceListItem, 0, len(list)), Total: count}
	for _, item := range list {
		res.Items = append(res.Items, resourceToReply(item))
	}
	return res, nil
}

func (uc *AdminUsecase) AddResource(ctx context.Context, req *v1.ResourceListItem) (*v1.ResourceListItem, error) {
	resourceItem := &ent.Resource{}
	copier.Copy(resourceItem, req)
	created, err := uc.repo.AddResource(ctx, resourceItem)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.RegisterPermissionSnapshot(ctx); err != nil {
		return nil, err
	}
	return resourceToReply(created), nil
}

func (uc *AdminUsecase) UpdateResource(ctx context.Context, req *v1.ResourceListItem) (*v1.ResourceListItem, error) {
	if _, err := uc.repo.GetResource(ctx, req.Id); err != nil {
		return nil, err
	}
	resourceItem := &ent.Resource{}
	copier.Copy(resourceItem, req)
	updated, err := uc.repo.UpdateResource(ctx, resourceItem)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.RegisterPermissionSnapshot(ctx); err != nil {
		return nil, err
	}
	return resourceToReply(updated), nil
}

func (uc *AdminUsecase) DelResource(ctx context.Context, resourceID string) error {
	if _, err := uc.repo.GetResource(ctx, resourceID); err != nil {
		return err
	}
	if err := uc.repo.DelResource(ctx, resourceID); err != nil {
		return err
	}
	return uc.repo.RegisterPermissionSnapshot(ctx)
}

func (uc *AdminUsecase) GetDeptList(ctx context.Context) (*v1.GetDeptListReply, error) {
	deptList, err := uc.repo.GetDeptList(ctx)
	if err != nil {
		return nil, err
	}
	forest := make([]*deptNode, 0)
	for _, dept := range deptList {
		if dept.ID == 0 {
			continue
		}
		uc.buildDeptTree(&forest, dept, true)
	}
	return &v1.GetDeptListReply{
		Items: toDeptTree(forest, ""),
		Total: int64(len(deptList)),
	}, nil
}

func (uc *AdminUsecase) AddDept(ctx context.Context, req *v1.DeptListItem) (*v1.DeptListItem, error) {
	dept, err := uc.repo.AddDept(ctx, req)
	if err != nil {
		return nil, err
	}
	return deptToReply(dept, req.Pid), nil
}

func (uc *AdminUsecase) UpdateDept(ctx context.Context, req *v1.DeptListItem) (*v1.DeptListItem, error) {
	deptID, err := tools.DeptStrSplitToInt(req.Id)
	if err != nil {
		return nil, err
	}
	dept, err := uc.repo.UpdateDept(ctx, deptID, req)
	if err != nil {
		return nil, err
	}
	return deptToReply(dept, req.Pid), nil
}

func (uc *AdminUsecase) DelDept(ctx context.Context, deptID string) error {
	id, err := tools.DeptStrSplitToInt(deptID)
	if err != nil {
		return err
	}
	dept, err := uc.repo.GetDeptById(ctx, id)
	if err != nil {
		return err
	}
	for {
		childrenList, err := uc.repo.GetDeptLeafsChildren(ctx, id)
		if err != nil {
			return err
		}
		if len(childrenList) == 0 {
			break
		}
		for _, child := range childrenList {
			if err := uc.repo.DelDept(ctx, child.ID); err != nil {
				return err
			}
		}
	}
	_ = dept
	return uc.repo.DelDept(ctx, id)
}

func (uc *AdminUsecase) GetCurrentUserMenus(ctx context.Context, userID string) (*v1.GetCurrentUserMenusReply, error) {
	_, err := uc.repo.GetUserAuthInfo(ctx, userID)
	if err != nil {
		return nil, err
	}
	binding, err := uc.repo.GetUserRoleBinding(ctx, userID)
	if err != nil {
		return nil, err
	}
	roleItem, err := uc.repo.GetRole(ctx, binding.RoleID)
	if err != nil {
		return nil, err
	}
	authority := &currentUserMenuAuthority{
		isRoot:  roleItem.Value == "root",
		menuIDs: append([]int32(nil), roleItem.Menus...),
	}
	menuList, err := uc.repo.GetMenuList(ctx)
	if err != nil {
		return nil, err
	}
	visibleMenus := filterVisibleMenus(menuList, authority)
	return &v1.GetCurrentUserMenusReply{Items: createCurrentUserMenuTree(visibleMenus)}, nil
}

func (uc *AdminUsecase) GetSysMenuList(ctx context.Context) (*v1.GetSysMenuListReply, error) {
	menuList, err := uc.repo.GetMenuList(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetSysMenuListReply{Items: uc.createMenuTree(menuList)}, nil
}

func (uc *AdminUsecase) ListMenus(ctx context.Context) (*v1.ListMenusReply, error) {
	menuList, err := uc.repo.GetMenuList(ctx)
	if err != nil {
		return nil, err
	}
	res := &v1.ListMenusReply{Items: make([]*v1.MenuRecord, 0, len(menuList))}
	for _, item := range menuList {
		res.Items = append(res.Items, entMenuToRecord(item))
	}
	return res, nil
}

func (uc *AdminUsecase) GetWalkRoute(ctx context.Context, selfItems ...tools.WalkRouteItem) (*v1.GetWalkRouteReply, error) {
	type walkRouteResult struct {
		items []tools.WalkRouteItem
		err   error
	}

	results := make(chan walkRouteResult, 3)
	var wg sync.WaitGroup
	fetchers := []func(context.Context) ([]tools.WalkRouteItem, error){
		uc.repo.ListAuthWalkRoutes,
		uc.repo.ListUserWalkRoutes,
		uc.repo.ListCommonWalkRoutes,
	}
	for _, fetch := range fetchers {
		wg.Add(1)
		go func(fetch func(context.Context) ([]tools.WalkRouteItem, error)) {
			defer wg.Done()
			items, err := fetch(ctx)
			results <- walkRouteResult{items: items, err: err}
		}(fetch)
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	items := make([]tools.WalkRouteItem, 0)
	for result := range results {
		if result.err != nil {
			return nil, result.err
		}
		items = append(items, result.items...)
	}
	items = append(items, selfItems...)
	return uc.buildWalkRouteReply(items), nil
}

func (uc *AdminUsecase) ListSelfWalkRoute(_ context.Context, items []tools.WalkRouteItem) ([]tools.WalkRouteItem, error) {
	return tools.SortAndUniqueWalkRoutes(items), nil
}

func (uc *AdminUsecase) BuildWalkRouteReply(items []tools.WalkRouteItem) *v1.GetWalkRouteReply {
	return uc.buildWalkRouteReply(items)
}

func (uc *AdminUsecase) buildWalkRouteReply(items []tools.WalkRouteItem) *v1.GetWalkRouteReply {
	items = tools.SortAndUniqueWalkRoutes(items)
	res := &v1.GetWalkRouteReply{Items: make([]*v1.WalkRouteItem, 0, len(items))}
	for _, item := range items {
		res.Items = append(res.Items, &v1.WalkRouteItem{Url: item.URL, Method: item.Method})
	}
	return res
}

func (uc *AdminUsecase) IsMenuNameExists(ctx context.Context, req *v1.IsMenuNameExistsRequest) (bool, error) {
	menuList, err := uc.repo.GetMenuList(ctx)
	if err != nil {
		return false, err
	}
	for _, menu := range menuList {
		if req.Name == menu.Name && menu.ID != req.Id {
			return true, nil
		}
	}
	return false, nil
}

func (uc *AdminUsecase) IsMenuPathExists(ctx context.Context, req *v1.IsMenuPathExistsRequest) (bool, error) {
	menuList, err := uc.repo.GetMenuList(ctx)
	if err != nil {
		return false, err
	}
	for _, menu := range menuList {
		if req.Path == menu.Path && menu.ID != req.Id {
			return true, nil
		}
	}
	return false, nil
}

func (uc *AdminUsecase) CreateMenu(ctx context.Context, req *v1.SysMenuListItem) error {
	_, err := uc.repo.CreateMenu(ctx, menuToEntMenu(req))
	return err
}

func (uc *AdminUsecase) UpdateMenu(ctx context.Context, req *v1.SysMenuListItem) error {
	_, err := uc.repo.UpdateMenu(ctx, int64(req.Id), menuToEntMenu(req))
	return err
}

func (uc *AdminUsecase) DeleteMenu(ctx context.Context, req *v1.DeleteMenuRequest) error {
	return uc.repo.DeleteMenu(ctx, req.Id)
}

func (uc *AdminUsecase) CreateSysLog(ctx context.Context, req *v1.CreateSysLogRequest) error {
	requestTime, err := time.Parse(time.DateTime, req.RequestTime)
	if err != nil {
		return err
	}
	return uc.repo.CreateSysLog(ctx, &ent.SysLogRecord{
		UserID:      req.UserId,
		UserName:    req.UserName,
		IsLogin:     req.IsLogin,
		SessionID:   req.SessionId,
		Method:      req.Method,
		Path:        req.Path,
		RequestTime: requestTime,
		IPAddress:   req.IpAddress,
		IPLocation:  req.IpLocation,
		Latency:     req.Latency,
		Os:          req.Os,
		Browser:     req.Browser,
		UserAgent:   req.UserAgent,
		Header:      req.Header,
		GetParams:   req.GetParams,
		PostData:    req.PostData,
		ResCode:     req.ResCode,
		Reason:      req.Reason,
		ResStatus:   req.ResStatus,
		Stack:       req.Stack,
	})
}

func (uc *AdminUsecase) GetSysLogList(ctx context.Context, req *v1.GetSysLogListParams) (*v1.GetSysLogListReply, error) {
	list, count, err := uc.repo.GetSysLogList(ctx, req)
	if err != nil {
		return nil, err
	}
	res := &v1.GetSysLogListReply{Items: make([]*v1.SysLogItem, 0, len(list)), Total: count}
	for _, item := range list {
		res.Items = append(res.Items, &v1.SysLogItem{
			Id:          item.ID,
			UserId:      item.UserID,
			UserName:    item.UserName,
			Method:      item.Method,
			Path:        item.Path,
			RequestTime: item.RequestTime.Format(time.DateTime),
			IpAddress:   item.IPAddress,
			IpLocation:  item.IPLocation,
			Latency:     item.Latency,
			Os:          item.Os,
			Browser:     item.Browser,
			ResCode:     item.ResCode,
			ResStatus:   item.ResStatus,
		})
	}
	return res, nil
}

func (uc *AdminUsecase) GetSysLogInfo(ctx context.Context, id string) (*v1.GetSysLogInfoReply, error) {
	info, err := uc.repo.GetSysLogInfo(ctx, id)
	if err != nil {
		return nil, err
	}
	return &v1.GetSysLogInfoReply{
		Id:          info.ID,
		UserId:      info.UserID,
		UserName:    info.UserName,
		IsLogin:     info.IsLogin,
		SessionId:   info.SessionID,
		Method:      info.Method,
		Path:        info.Path,
		RequestTime: info.RequestTime.Format(time.DateTime),
		IpAddress:   info.IPAddress,
		IpLocation:  info.IPLocation,
		Latency:     info.Latency,
		Os:          info.Os,
		Browser:     info.Browser,
		UserAgent:   info.UserAgent,
		Header:      info.Header,
		GetParams:   info.GetParams,
		PostData:    info.PostData,
		ResCode:     info.ResCode,
		Reason:      info.Reason,
		ResStatus:   info.ResStatus,
		Stack:       info.Stack,
		CreateTime:  info.CreateTime.Format(time.DateTime),
	}, nil
}

// remaining helper functions unchanged below

type deptNode struct {
	Id       int64
	Pid      int64
	Value    *ent.Dept
	Children []*deptNode
}

func deptToNode(value *ent.Dept) *deptNode {
	return &deptNode{Id: value.ID, Pid: value.Pid, Value: value}
}

func (uc *AdminUsecase) buildDeptTree(forest *[]*deptNode, dept *ent.Dept, top bool) bool {
	if dept.Pid == 0 {
		*forest = append(*forest, deptToNode(dept))
		return true
	}
	for _, item := range *forest {
		if item.Id == dept.Pid {
			item.Children = append(item.Children, deptToNode(dept))
			return true
		}
		if len(item.Children) > 0 && uc.buildDeptTree(&item.Children, dept, false) {
			return true
		}
	}
	if top {
		*forest = append(*forest, deptToNode(dept))
		return true
	}
	return false
}

func toDeptTree(forest []*deptNode, strPid string) []*v1.DeptListItem {
	items := make([]*v1.DeptListItem, 0, len(forest))
	for _, item := range forest {
		id := strconv.FormatInt(item.Id, 10)
		status := int32(0)
		if item.Value.Status {
			status = 1
		}
		items = append(items, &v1.DeptListItem{
			Id:         id,
			Pid:        strPid,
			Name:       item.Value.Name,
			OrderNo:    item.Value.Sort,
			Remark:     item.Value.Desc,
			Status:     status,
			CreateTime: item.Value.CreateTime.Format(time.DateTime),
			Dom:        item.Value.Dom,
			Children:   toDeptTree(item.Children, id),
		})
	}
	return items
}

func deptToReply(dept *ent.Dept, pid string) *v1.DeptListItem {
	status := int32(0)
	if dept.Status {
		status = 1
	}
	return &v1.DeptListItem{
		Id:         strconv.FormatInt(dept.ID, 10),
		Pid:        pid,
		Name:       dept.Name,
		OrderNo:    dept.Sort,
		Remark:     dept.Desc,
		Status:     status,
		CreateTime: dept.CreateTime.Format(time.DateTime),
		Dom:        dept.Dom,
	}
}

func (uc *AdminUsecase) createMenuTree(menuList []*ent.Menu) []*v1.SysMenuListItem {
	items := make([]*v1.SysMenuListItem, 0)
	for _, menu := range menuList {
		if menu.Pid == 0 {
			items = append(items, entMenuToMenu(menu))
			continue
		}
		if !buildMenuTree(&items, menu) {
			items = append(items, entMenuToMenu(menu))
		}
	}
	return items
}

func buildMenuTree(menuList *[]*v1.SysMenuListItem, menu *ent.Menu) bool {
	for _, item := range *menuList {
		if len(item.Children) > 0 && buildMenuTree(&item.Children, menu) {
			return true
		}
		if int64(item.Id) == menu.Pid {
			item.Children = append(item.Children, entMenuToMenu(menu))
			return true
		}
	}
	return false
}

func filterVisibleMenus(menuList []*ent.Menu, authority *currentUserMenuAuthority) []*ent.Menu {
	if authority.isRoot {
		items := make([]*ent.Menu, 0, len(menuList))
		for _, item := range menuList {
			if item.Status {
				items = append(items, item)
			}
		}
		sortMenus(items)
		return items
	}
	allowed := make(map[int64]struct{}, len(authority.menuIDs))
	for _, id := range authority.menuIDs {
		allowed[int64(id)] = struct{}{}
	}
	menuByID := make(map[int64]*ent.Menu, len(menuList))
	for _, item := range menuList {
		menuByID[item.ID] = item
	}
	for _, id := range authority.menuIDs {
		item, ok := menuByID[int64(id)]
		if !ok {
			continue
		}
		pid := item.Pid
		for pid > 0 {
			parent, ok := menuByID[pid]
			if !ok {
				break
			}
			allowed[parent.ID] = struct{}{}
			pid = parent.Pid
		}
	}
	items := make([]*ent.Menu, 0, len(menuList))
	for _, item := range menuList {
		if !item.Status {
			continue
		}
		if _, ok := allowed[item.ID]; !ok {
			continue
		}
		items = append(items, item)
	}
	sortMenus(items)
	return items
}

func sortMenus(items []*ent.Menu) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Pid != items[j].Pid {
			return items[i].Pid < items[j].Pid
		}
		if items[i].Order != items[j].Order {
			return items[i].Order < items[j].Order
		}
		return items[i].ID < items[j].ID
	})
}

func createCurrentUserMenuTree(menuList []*ent.Menu) []*v1.CurrentUserMenuItem {
	items := make([]*v1.CurrentUserMenuItem, 0)
	for _, menu := range menuList {
		if menu.Pid == 0 {
			items = append(items, entMenuToCurrentUserMenu(menu))
			continue
		}
		if !buildCurrentUserMenuTree(&items, menu) {
			items = append(items, entMenuToCurrentUserMenu(menu))
		}
	}
	return items
}

func buildCurrentUserMenuTree(menuList *[]*v1.CurrentUserMenuItem, menu *ent.Menu) bool {
	for _, item := range *menuList {
		if len(item.Children) > 0 && buildCurrentUserMenuTree(&item.Children, menu) {
			return true
		}
		if int64(item.Id) == menu.Pid {
			item.Children = append(item.Children, entMenuToCurrentUserMenu(menu))
			return true
		}
	}
	return false
}

func entMenuToMenu(menu *ent.Menu) *v1.SysMenuListItem {
	status := int32(0)
	if menu.Status {
		status = 1
	}
	redirect := menu.Redirect
	activePath := menu.ActivePath
	link := menu.Link
	iframeSrc := menu.IframeSrc
	keepAlive := menu.Keepalive
	maxNumOfOpenTab := int64(menu.MaxNumOfOpenTab)
	hideInMenu := menu.HideInMenu
	hideInTab := menu.HideInTab
	hideInBreadcrumb := menu.HideInBreadcrumb
	hideChildrenInMenu := menu.HideChildrenInMenu
	return &v1.SysMenuListItem{
		Id:         int32(menu.ID),
		Component:  menu.Component,
		Status:     &status,
		AuthCode:   "",
		Name:       menu.Name,
		Path:       menu.Path,
		Pid:        menu.Pid,
		Redirect:   &redirect,
		Type:       menu.Type,
		CreateTime: menu.CreateTime.Format(time.DateTime),
		Meta: &v1.Meta{
			Order:              int64(menu.Order),
			Icon:               menu.Icon,
			Title:              menu.Title,
			ActivePath:         &activePath,
			IframeSrc:          &iframeSrc,
			Link:               &link,
			KeepAlive:          &keepAlive,
			MaxNumOfOpenTab:    &maxNumOfOpenTab,
			IgnoreAccess:       menu.IgnoreAccess,
			HideInMenu:         &hideInMenu,
			HideInTab:          &hideInTab,
			HideInBreadcrumb:   &hideInBreadcrumb,
			HideChildrenInMenu: &hideChildrenInMenu,
			Authority:          splitAuthority(menu.Authority),
		},
	}
}

func entMenuToCurrentUserMenu(menu *ent.Menu) *v1.CurrentUserMenuItem {
	item := entMenuToMenu(menu)
	return &v1.CurrentUserMenuItem{
		Id:         item.Id,
		Component:  item.Component,
		Status:     item.Status,
		AuthCode:   item.AuthCode,
		Name:       item.Name,
		Path:       item.Path,
		Pid:        item.Pid,
		Redirect:   item.Redirect,
		Type:       item.Type,
		Meta:       item.Meta,
		CreateTime: item.CreateTime,
	}
}

func entMenuToRecord(menu *ent.Menu) *v1.MenuRecord {
	item := entMenuToMenu(menu)
	return &v1.MenuRecord{
		Id:         item.Id,
		Component:  item.Component,
		Status:     item.Status,
		AuthCode:   item.AuthCode,
		Name:       item.Name,
		Path:       item.Path,
		Pid:        item.Pid,
		Redirect:   item.Redirect,
		Type:       item.Type,
		Meta:       item.Meta,
		CreateTime: item.CreateTime,
	}
}

func menuToEntMenu(menu *v1.SysMenuListItem) *ent.Menu {
	meta := menu.Meta
	if meta == nil {
		meta = &v1.Meta{}
	}
	status := false
	if menu.Status != nil && *menu.Status == 1 {
		status = true
	}
	return &ent.Menu{
		ID:                 int64(menu.Id),
		Pid:                menu.Pid,
		Type:               menu.Type,
		Status:             status,
		Path:               menu.Path,
		Redirect:           ptrToString(menu.Redirect),
		Name:               menu.Name,
		Component:          menu.Component,
		Icon:               meta.Icon,
		Title:              meta.Title,
		Order:              int32(meta.Order),
		Link:               ptrToString(meta.Link),
		IframeSrc:          ptrToString(meta.IframeSrc),
		ActivePath:         ptrToString(meta.ActivePath),
		MaxNumOfOpenTab:    int16(ptrToInt64(meta.MaxNumOfOpenTab)),
		Keepalive:          ptrToBool(meta.KeepAlive),
		IgnoreAccess:       meta.IgnoreAccess,
		Authority:          strings.Join(meta.Authority, ","),
		HideInMenu:         ptrToBool(meta.HideInMenu),
		HideInTab:          ptrToBool(meta.HideInTab),
		HideInBreadcrumb:   ptrToBool(meta.HideInBreadcrumb),
		HideChildrenInMenu: ptrToBool(meta.HideChildrenInMenu),
	}
}

func ptrToString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func ptrToBool(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}

func ptrToInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func splitAuthority(v string) []string {
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	res := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			res = append(res, part)
		}
	}
	return res
}

func replaceBracesIfExists(str string) (bool, string) {
	hasBraces := strings.Contains(str, "{") || strings.Contains(str, "}")
	if !hasBraces {
		return false, str
	}
	re := regexp.MustCompile(`{[^}]*}`)
	return true, re.ReplaceAllString(str, "%")
}

func roleToReply(item *ent.Role, order int) *v1.RoleListItem {
	apiPermissions := make([]string, 0)
	if item.Edges.Resource != nil {
		for _, resourceItem := range item.Edges.Resource {
			if resourceItem.Type == "api" {
				apiPermissions = append(apiPermissions, resourceItem.ID)
			}
		}
	}
	status := int32(0)
	if item.Status {
		status = 1
	}
	return &v1.RoleListItem{
		Id:             strconv.FormatInt(item.ID, 10),
		Name:           item.Name,
		Value:          item.Value,
		Status:         status,
		OrderNo:        strconv.Itoa(order),
		CreateTime:     item.CreateTime.Format(time.DateTime),
		Remark:         item.Desc,
		Permissions:    item.Menus,
		ApiPermissions: apiPermissions,
	}
}

func apiToReply(item *ent.ApiResources) *v1.ApiListItem {
	return &v1.ApiListItem{
		Id:                item.ID,
		Path:              item.Path,
		Method:            item.Method,
		Description:       item.Description,
		Module:            item.Module,
		ModuleDescription: item.ModuleDescription,
		ResourcesGroup:    item.ResourcesGroup,
	}
}

func resourceToReply(item *ent.Resource) *v1.ResourceListItem {
	return &v1.ResourceListItem{
		Id:          item.ID,
		Name:        item.Name,
		Type:        item.Type,
		Value:       item.Value,
		Method:      item.Method,
		Description: item.Description,
	}
}

func roleToAuthReply(item *ent.Role) *v1.AuthRoleItem {
	resources := make([]*v1.AuthRoleResourceItem, 0)
	if item.Edges.Resource != nil {
		for _, resourceItem := range item.Edges.Resource {
			resources = append(resources, &v1.AuthRoleResourceItem{
				Id:     resourceItem.ID,
				Type:   resourceItem.Type,
				Value:  resourceItem.Value,
				Method: resourceItem.Method,
			})
		}
	}
	return &v1.AuthRoleItem{
		Id:        item.ID,
		Name:      item.Name,
		Value:     item.Value,
		Status:    item.Status,
		Remark:    item.Desc,
		MenuIds:   append([]int32(nil), item.Menus...),
		Resources: resources,
	}
}

func apiToAuthReply(item *ent.ApiResources) *v1.AuthApiItem {
	return &v1.AuthApiItem{
		Id:                item.ID,
		Path:              item.Path,
		Method:            item.Method,
		Description:       item.Description,
		Module:            item.Module,
		ModuleDescription: item.ModuleDescription,
		ResourcesGroup:    item.ResourcesGroup,
	}
}

func userRoleBindingToReply(item *ent.UserRoleBinding) *v1.UserRoleBindingItem {
	return &v1.UserRoleBindingItem{
		Id:         strconv.FormatInt(item.ID, 10),
		UserId:     item.UserID,
		RoleId:     item.RoleID,
		CreateTime: item.CreateTime.Format(time.DateTime),
		UpdateTime: item.UpdateTime.Format(time.DateTime),
	}
}

var _ = emptypb.Empty{}
var _ = authx.UserID
