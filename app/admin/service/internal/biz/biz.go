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
	GetDefaultOrganizationID(context.Context) (string, error)
	ResolveRoleValues(context.Context, []int64) (map[int64]string, error)
	GetUserRoleBindings(context.Context, string, string) ([]*ent.UserRoleBinding, error)
	ListUserRoleBindings(context.Context) ([]*ent.UserRoleBinding, error)
	UpsertUserRoleBinding(context.Context, string, string, []int64) ([]*ent.UserRoleBinding, error)
	DeleteUserRoleBinding(context.Context, string, string) error
	GetUserDeptBinding(context.Context, string, string) (*ent.UserDeptMembership, error)
	UpsertUserDeptBinding(context.Context, string, string, int64) (*ent.UserDeptMembership, error)
	DeleteUserDeptBinding(context.Context, string, string) error
	ListOrganizations(context.Context, *v1.GetOrganizationListParams) ([]*ent.Organization, int64, error)
	GetOrganization(context.Context, string) (*ent.Organization, error)
	AddOrganization(context.Context, *v1.OrganizationItem) (*ent.Organization, error)
	UpdateOrganization(context.Context, string, *v1.OrganizationItem) (*ent.Organization, error)
	DelOrganization(context.Context, string) error
	CountOrganizationDepts(context.Context, string) (int64, error)
	CountOrganizationMembers(context.Context, string) (int64, error)
	ListOrganizationMembers(context.Context, string) ([]*v1.OrganizationMemberItem, error)
	SaveOrganizationMembers(context.Context, string, []string) ([]*v1.OrganizationMemberItem, error)
	EnsureAllUsersInDefaultOrganization(context.Context) error
	EnsureUserInDefaultOrganization(context.Context, string) error
	ListUserOrganizations(context.Context, string) ([]*ent.UserOrganization, error)
	GetCurrentOrganizationID(context.Context, string) (string, error)
	UserBelongsToOrganization(context.Context, string, string) (bool, error)
	SwitchCurrentOrganization(context.Context, string, string) error
	RemoveOrganizationRoleBindings(context.Context, string, string) error
	RemoveOrganizationDeptBinding(context.Context, string, string) error
	ListOrganizationPermissionScopes(context.Context, string) ([]*ent.OrganizationPermissionScope, error)
	SaveOrganizationPermissionScopes(context.Context, string, []int32, []string, string) ([]*ent.OrganizationPermissionScope, error)
	ListUserOrganizationRoles(context.Context, string, string) ([]*ent.Role, error)
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
	EnsurePermissionBootstrap(context.Context) error
	RegisterPermissionSnapshot(context.Context) error
	ApplyAuthRoleDelta(context.Context, *ent.Role, *ent.Role) error
	ApplyAuthApiDelta(context.Context, *ent.ApiResources, *ent.ApiResources) error
	ListAuthWalkRoutes(context.Context) ([]tools.WalkRouteItem, error)
	ListUserWalkRoutes(context.Context) ([]tools.WalkRouteItem, error)
	ListCommonWalkRoutes(context.Context) ([]tools.WalkRouteItem, error)
	CreateMenu(context.Context, *ent.Menu) (*ent.Menu, error)
	UpdateMenu(context.Context, int64, *ent.Menu) (*ent.Menu, error)
	DeleteMenu(context.Context, int64) error
	GetDeptList(context.Context, string) ([]*ent.Dept, error)
	AddDept(context.Context, *v1.DeptListItem) (*ent.Dept, error)
	UpdateDept(context.Context, int64, *v1.DeptListItem) (*ent.Dept, error)
	DelDept(context.Context, int64) error
	GetDeptLeafsChildren(context.Context, int64) ([]*ent.Dept, error)
	GetDeptById(context.Context, int64) (*ent.Dept, error)
	RemoveDeptBindings(context.Context, int64) error
	CreateSysLog(context.Context, *ent.SysLogRecord) error
	GetSysLogList(context.Context, *v1.GetSysLogListParams) ([]*ent.SysLogRecord, int64, error)
	GetSysLogInfo(context.Context, string) (*ent.SysLogRecord, error)
	ListServiceRegistries(context.Context) ([]*ent.ServiceRegistry, error)
	ListProjectionSourceStatuses(context.Context) ([]*ent.ProjectionSourceStatus, error)
	SaveProjectionSourceStatus(context.Context, authx.ProjectionStatus) error
}

type AdminUsecase struct {
	repo AdminRepo
	log  *log.Helper
}

type currentUserMenuAuthority struct {
	isRoot  bool
	menuIDs []int32
}

type actorPermissionGrant struct {
	isPlatformRoot bool
	menuIDs        map[int32]struct{}
	resourceIDs    map[string]struct{}
	dataScopes     map[string]struct{}
	customDeptIDs  map[int64]struct{}
}

func NewAdminUsecase(repo AdminRepo, logger log.Logger) *AdminUsecase {
	return &AdminUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *AdminUsecase) RegisterPermissionSnapshot(ctx context.Context) error {
	if err := uc.repo.EnsurePermissionBootstrap(ctx); err != nil {
		return fmt.Errorf("ensure permission bootstrap: %w", err)
	}
	if err := uc.ensureBuiltinSiteMessageBootstrap(ctx); err != nil {
		return fmt.Errorf("ensure site-message bootstrap: %w", err)
	}
	if err := uc.repo.RegisterPermissionSnapshot(ctx); err != nil {
		return fmt.Errorf("register permission snapshot: %w", err)
	}
	uc.log.Info("auth projection snapshot registered")
	return nil
}

func (uc *AdminUsecase) syncAuthProjection(ctx context.Context, label string, apply func() error) error {
	return authx.SyncProjectionDelta(ctx, uc.log, label, apply, uc.RegisterPermissionSnapshot)
}

func (uc *AdminUsecase) loadRoleWithResources(ctx context.Context, roleID int64) (*ent.Role, error) {
	roleItem, err := uc.repo.GetRole(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if roleItem.Edges.Resource != nil {
		return roleItem, nil
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
	organizationID, err := uc.resolveRoleBindingOrganization(ctx, req.GetOrganizationId())
	if err != nil {
		return nil, err
	}
	items, err := uc.repo.GetUserRoleBindings(ctx, req.UserId, organizationID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return &v1.UserRoleBindingItem{
			OrganizationId: organizationID,
			UserId:         req.UserId,
		}, nil
	}
	return userRoleBindingsToReply(req.UserId, organizationID, items), nil
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
	organizationID, err := uc.resolveRoleBindingOrganization(ctx, req.GetOrganizationId())
	if err != nil {
		return nil, err
	}
	if err := uc.ensureUserOrganizationMember(ctx, req.GetUserId(), organizationID); err != nil {
		return nil, err
	}
	roleIDs, err := uc.normalizeUserRoleBindingIDs(ctx, normalizeRoleIDs(req.RoleIds, req.RoleId))
	if err != nil {
		return nil, err
	}
	platformRoot, err := uc.isPlatformRoot(ctx)
	if err != nil {
		return nil, err
	}
	for _, roleID := range roleIDs {
		roleItem, err := uc.repo.GetRole(ctx, roleID)
		if err != nil {
			return nil, err
		}
		if roleItem.OrganizationID != organizationID {
			return nil, fmt.Errorf("role does not belong to organization")
		}
		if roleItem.Value == "root" && !platformRoot {
			return nil, fmt.Errorf("root role can only be assigned by platform root")
		}
	}
	if err := uc.validateUserRoleBindingGrantable(ctx, organizationID, roleIDs); err != nil {
		return nil, err
	}
	items, err := uc.repo.UpsertUserRoleBinding(ctx, req.UserId, organizationID, roleIDs)
	if err != nil {
		return nil, err
	}
	if err := uc.RegisterPermissionSnapshot(ctx); err != nil {
		return nil, err
	}
	uc.log.Infof("auth projection snapshot registered: user-role-binding upsert user=%s organization=%s role_ids=%v", req.UserId, organizationID, roleIDs)
	return userRoleBindingsToReply(req.UserId, organizationID, items), nil
}

func (uc *AdminUsecase) DeleteUserRoleBinding(ctx context.Context, req *v1.DeleteUserRoleBindingRequest) error {
	organizationID, err := uc.resolveRoleBindingOrganization(ctx, req.GetOrganizationId())
	if err != nil {
		return err
	}
	if err := uc.repo.DeleteUserRoleBinding(ctx, req.UserId, organizationID); err != nil {
		return err
	}
	if err := uc.RegisterPermissionSnapshot(ctx); err != nil {
		return err
	}
	uc.log.Infof("auth projection snapshot registered: user-role-binding delete user=%s organization=%s", req.UserId, organizationID)
	return nil
}

func (uc *AdminUsecase) GetUserDeptBinding(ctx context.Context, req *v1.GetUserDeptBindingRequest) (*v1.UserDeptBindingItem, error) {
	organizationID, err := uc.resolveOrganizationID(ctx, req.GetOrganizationId())
	if err != nil {
		return nil, err
	}
	item, err := uc.repo.GetUserDeptBinding(ctx, req.GetUserId(), organizationID)
	if err != nil {
		if ent.IsNotFound(err) {
			return &v1.UserDeptBindingItem{
				OrganizationId: organizationID,
				UserId:         req.GetUserId(),
			}, nil
		}
		return nil, err
	}
	return uc.userDeptBindingToReply(ctx, item)
}

func (uc *AdminUsecase) UpsertUserDeptBinding(ctx context.Context, req *v1.UserDeptBindingItem) (*v1.UserDeptBindingItem, error) {
	organizationID, err := uc.resolveOrganizationID(ctx, req.GetOrganizationId())
	if err != nil {
		return nil, err
	}
	if err := uc.ensureUserOrganizationMember(ctx, req.GetUserId(), organizationID); err != nil {
		return nil, err
	}
	deptID, err := uc.normalizeDeptBindingDeptID(req.GetDeptId())
	if err != nil {
		return nil, err
	}
	if deptID == 0 {
		if err := uc.repo.DeleteUserDeptBinding(ctx, req.GetUserId(), organizationID); err != nil {
			return nil, err
		}
		return &v1.UserDeptBindingItem{
			OrganizationId: organizationID,
			UserId:         req.GetUserId(),
		}, nil
	}
	deptItem, err := uc.repo.GetDeptById(ctx, deptID)
	if err != nil {
		return nil, err
	}
	if err := validateDeptParentOrganization(deptItem, organizationID); err != nil {
		return nil, fmt.Errorf("department does not belong to organization")
	}
	item, err := uc.repo.UpsertUserDeptBinding(ctx, req.GetUserId(), organizationID, deptID)
	if err != nil {
		return nil, err
	}
	return uc.userDeptBindingToReply(ctx, item)
}

func (uc *AdminUsecase) DeleteUserDeptBinding(ctx context.Context, req *v1.DeleteUserDeptBindingRequest) error {
	organizationID, err := uc.resolveOrganizationID(ctx, req.GetOrganizationId())
	if err != nil {
		return err
	}
	return uc.repo.DeleteUserDeptBinding(ctx, req.GetUserId(), organizationID)
}

func (uc *AdminUsecase) GetOrganizationList(ctx context.Context, req *v1.GetOrganizationListParams) (*v1.GetOrganizationListReply, error) {
	if err := uc.repo.EnsureAllUsersInDefaultOrganization(ctx); err != nil {
		return nil, err
	}
	platformRoot, err := uc.isPlatformRoot(ctx)
	if err != nil {
		return nil, err
	}
	if actorUserID := strings.TrimSpace(authx.UserID(ctx)); actorUserID != "" && !platformRoot {
		return uc.getScopedOrganizationList(ctx, actorUserID, req)
	}
	list, count, err := uc.repo.ListOrganizations(ctx, req)
	if err != nil {
		return nil, err
	}
	res := &v1.GetOrganizationListReply{
		Items:                  make([]*v1.OrganizationItem, 0, len(list)),
		Total:                  count,
		CanManageOrganizations: platformRoot,
	}
	for _, item := range list {
		reply, err := uc.organizationToReply(ctx, item)
		if err != nil {
			return nil, err
		}
		res.Items = append(res.Items, reply)
	}
	return res, nil
}

func (uc *AdminUsecase) getScopedOrganizationList(ctx context.Context, actorUserID string, req *v1.GetOrganizationListParams) (*v1.GetOrganizationListReply, error) {
	currentOrganizationID, err := uc.repo.GetCurrentOrganizationID(ctx, actorUserID)
	if err != nil {
		return nil, err
	}
	ids := normalizeStringIDs([]string{currentOrganizationID})
	list := make([]*ent.Organization, 0, len(ids))
	for _, id := range ids {
		item, err := uc.repo.GetOrganization(ctx, id)
		if err != nil {
			if ent.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		if !organizationMatchesListParams(item, req) {
			continue
		}
		list = append(list, item)
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].Sort != list[j].Sort {
			return list[i].Sort < list[j].Sort
		}
		return list[i].ID < list[j].ID
	})
	total := int64(len(list))
	list = paginateOrganizations(list, req)
	res := &v1.GetOrganizationListReply{
		Items:                  make([]*v1.OrganizationItem, 0, len(list)),
		Total:                  total,
		CanManageOrganizations: false,
	}
	for _, item := range list {
		reply, err := uc.organizationToReply(ctx, item)
		if err != nil {
			return nil, err
		}
		res.Items = append(res.Items, reply)
	}
	return res, nil
}

func organizationMatchesListParams(item *ent.Organization, req *v1.GetOrganizationListParams) bool {
	if item == nil {
		return false
	}
	if req == nil {
		return true
	}
	if req.Name != "" && !strings.Contains(item.Name, req.Name) {
		return false
	}
	if req.Code != "" && !strings.Contains(item.Code, req.Code) {
		return false
	}
	if req.Status == 1 && !item.Status {
		return false
	}
	if req.Status == 2 && item.Status {
		return false
	}
	return true
}

func paginateOrganizations(items []*ent.Organization, req *v1.GetOrganizationListParams) []*ent.Organization {
	if req == nil || req.PageSize <= 0 {
		return items
	}
	offset := int64(0)
	if req.CurrentPage > 0 {
		offset = (req.CurrentPage - 1) * req.PageSize
	}
	if offset >= int64(len(items)) {
		return []*ent.Organization{}
	}
	end := offset + req.PageSize
	if end > int64(len(items)) {
		end = int64(len(items))
	}
	return items[offset:end]
}

func (uc *AdminUsecase) AddOrganization(ctx context.Context, req *v1.OrganizationItem) (*v1.OrganizationItem, error) {
	if err := uc.requirePlatformRoot(ctx, "organization can only be created by platform root"); err != nil {
		return nil, err
	}
	item, err := uc.repo.AddOrganization(ctx, req)
	if err != nil {
		return nil, err
	}
	return uc.organizationToReply(ctx, item)
}

func (uc *AdminUsecase) UpdateOrganization(ctx context.Context, req *v1.OrganizationItem) (*v1.OrganizationItem, error) {
	if err := uc.requirePlatformRoot(ctx, "organization can only be updated by platform root"); err != nil {
		return nil, err
	}
	id, err := normalizeOrganizationID(req.Id)
	if err != nil {
		return nil, err
	}
	item, err := uc.repo.UpdateOrganization(ctx, id, req)
	if err != nil {
		return nil, err
	}
	return uc.organizationToReply(ctx, item)
}

func (uc *AdminUsecase) DelOrganization(ctx context.Context, organizationID string) error {
	if err := uc.requirePlatformRoot(ctx, "organization can only be deleted by platform root"); err != nil {
		return err
	}
	id, err := normalizeOrganizationID(organizationID)
	if err != nil {
		return err
	}
	if _, err := uc.repo.GetOrganization(ctx, id); err != nil {
		return err
	}
	deptCount, err := uc.repo.CountOrganizationDepts(ctx, id)
	if err != nil {
		return err
	}
	if deptCount > 0 {
		return fmt.Errorf("organization has departments")
	}
	memberCount, err := uc.repo.CountOrganizationMembers(ctx, id)
	if err != nil {
		return err
	}
	if memberCount > 0 {
		return fmt.Errorf("organization has members")
	}
	return uc.repo.DelOrganization(ctx, id)
}

func (uc *AdminUsecase) GetOrganizationMembers(ctx context.Context, req *v1.GetOrganizationMembersRequest) (*v1.GetOrganizationMembersReply, error) {
	id, err := uc.resolveOrganizationMemberReadID(ctx, req.GetOrganizationId())
	if err != nil {
		return nil, err
	}
	defaultOrganizationID, err := uc.repo.GetDefaultOrganizationID(ctx)
	if err != nil {
		return nil, err
	}
	if id == defaultOrganizationID {
		if err := uc.repo.EnsureAllUsersInDefaultOrganization(ctx); err != nil {
			return nil, err
		}
	}
	items, err := uc.repo.ListOrganizationMembers(ctx, id)
	if err != nil {
		return nil, err
	}
	return &v1.GetOrganizationMembersReply{Items: items, Total: int64(len(items))}, nil
}

func (uc *AdminUsecase) SaveOrganizationMembers(ctx context.Context, req *v1.SaveOrganizationMembersRequest) (*v1.GetOrganizationMembersReply, error) {
	id, err := uc.resolveOrganizationID(ctx, req.GetOrganizationId())
	if err != nil {
		return nil, err
	}
	defaultOrganizationID, err := uc.repo.GetDefaultOrganizationID(ctx)
	if err != nil {
		return nil, err
	}
	if id == defaultOrganizationID {
		if err := uc.repo.EnsureAllUsersInDefaultOrganization(ctx); err != nil {
			return nil, err
		}
	}
	if id != defaultOrganizationID {
		existingMembers, err := uc.repo.ListOrganizationMembers(ctx, id)
		if err != nil {
			return nil, err
		}
		removedUserIDs := diffRemovedOrganizationMembers(existingMembers, req.UserIds)
		for _, userID := range removedUserIDs {
			if err := uc.repo.RemoveOrganizationRoleBindings(ctx, userID, id); err != nil {
				return nil, err
			}
			if err := uc.repo.RemoveOrganizationDeptBinding(ctx, userID, id); err != nil {
				return nil, err
			}
		}
	}
	items, err := uc.repo.SaveOrganizationMembers(ctx, id, req.UserIds)
	if err != nil {
		return nil, err
	}
	if id != defaultOrganizationID {
		if err := uc.RegisterPermissionSnapshot(ctx); err != nil {
			return nil, err
		}
	}
	return &v1.GetOrganizationMembersReply{Items: items, Total: int64(len(items))}, nil
}

func (uc *AdminUsecase) GetOrganizationPermissionScope(ctx context.Context, req *v1.GetOrganizationPermissionScopeRequest) (*v1.OrganizationPermissionScopeReply, error) {
	organizationID, err := uc.resolveOrganizationID(ctx, req.GetOrganizationId())
	if err != nil {
		return nil, err
	}
	scopes, err := uc.repo.ListOrganizationPermissionScopes(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	return organizationScopesToReply(organizationID, scopes), nil
}

func (uc *AdminUsecase) SaveOrganizationPermissionScope(ctx context.Context, req *v1.SaveOrganizationPermissionScopeRequest) (*v1.OrganizationPermissionScopeReply, error) {
	if err := uc.requirePlatformRoot(ctx, "organization permission scope can only be updated by platform root"); err != nil {
		return nil, err
	}
	organizationID, err := uc.resolveOrganizationID(ctx, req.GetOrganizationId())
	if err != nil {
		return nil, err
	}
	menuIDs := normalizeMenuIDs(req.GetMenuIds())
	resourceIDs := normalizeStringIDs(req.GetResourceIds())
	if err := uc.validatePermissionScopeRefs(ctx, menuIDs, resourceIDs); err != nil {
		return nil, err
	}
	if err := uc.validateOrganizationPermissionScopeShrink(ctx, organizationID, menuIDs, resourceIDs); err != nil {
		return nil, err
	}
	items, err := uc.repo.SaveOrganizationPermissionScopes(ctx, organizationID, menuIDs, resourceIDs, authx.UserID(ctx))
	if err != nil {
		return nil, err
	}
	return organizationScopesToReply(organizationID, items), nil
}

func (uc *AdminUsecase) GetCurrentPermissionCatalog(ctx context.Context) (*v1.OrganizationPermissionCatalogReply, error) {
	organizationID, err := uc.resolveOrganizationID(ctx, "")
	if err != nil {
		return nil, err
	}
	return uc.getOrganizationPermissionCatalog(ctx, organizationID)
}

func (uc *AdminUsecase) GetOrganizationPermissionCatalog(ctx context.Context, req *v1.GetOrganizationPermissionCatalogRequest) (*v1.OrganizationPermissionCatalogReply, error) {
	organizationID, err := uc.resolveOrganizationID(ctx, req.GetOrganizationId())
	if err != nil {
		return nil, err
	}
	return uc.getOrganizationPermissionCatalog(ctx, organizationID)
}

func (uc *AdminUsecase) getOrganizationPermissionCatalog(ctx context.Context, organizationID string) (*v1.OrganizationPermissionCatalogReply, error) {
	scopes, err := uc.repo.ListOrganizationPermissionScopes(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	scope := organizationScopesToReply(organizationID, scopes)
	menuList, err := uc.repo.GetMenuList(ctx)
	if err != nil {
		return nil, err
	}
	resourceList, _, err := uc.repo.GetResourceList(ctx, &v1.GetResourcePageParams{})
	if err != nil {
		return nil, err
	}
	menus := filterMenusByScope(menuList, scope.MenuIds)
	resources := filterResourcesByScope(resourceList, scope.ResourceIds)
	res := &v1.OrganizationPermissionCatalogReply{
		OrganizationId:    organizationID,
		Menus:             uc.createMenuTree(menus),
		Resources:         make([]*v1.ResourceListItem, 0, len(resources)),
		ScopedMenuIds:     append([]int32(nil), scope.MenuIds...),
		ScopedResourceIds: append([]string(nil), scope.ResourceIds...),
	}
	for _, item := range resources {
		res.Resources = append(res.Resources, resourceToReply(item))
	}
	return res, nil
}

func (uc *AdminUsecase) GetMyOrganizations(ctx context.Context, userID string) (*v1.GetMyOrganizationsReply, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}
	if err := uc.repo.EnsureUserInDefaultOrganization(ctx, userID); err != nil {
		return nil, err
	}
	currentOrganizationID, err := uc.repo.GetCurrentOrganizationID(ctx, userID)
	if err != nil {
		return nil, err
	}
	memberships, err := uc.repo.ListUserOrganizations(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]*v1.MyOrganizationItem, 0, len(memberships))
	for _, membership := range memberships {
		if membership == nil {
			continue
		}
		organizationItem, err := uc.repo.GetOrganization(ctx, membership.OrganizationID)
		if err != nil {
			if ent.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		if !organizationItem.Status {
			continue
		}
		items = append(items, myOrganizationToReply(organizationItem, membership.OrganizationID == currentOrganizationID))
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Current != items[j].Current {
			return items[i].Current
		}
		if items[i].OrderNo != items[j].OrderNo {
			return items[i].OrderNo < items[j].OrderNo
		}
		if items[i].Name != items[j].Name {
			return items[i].Name < items[j].Name
		}
		return items[i].Id < items[j].Id
	})
	return &v1.GetMyOrganizationsReply{
		Items:                 items,
		CurrentOrganizationId: currentOrganizationID,
	}, nil
}

func (uc *AdminUsecase) SwitchCurrentOrganization(ctx context.Context, userID string, req *v1.SwitchCurrentOrganizationRequest) (*v1.CurrentOrganizationReply, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}
	organizationID, err := normalizeOrganizationID(req.GetOrganizationId())
	if err != nil {
		return nil, err
	}
	organizationItem, err := uc.repo.GetOrganization(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if !organizationItem.Status {
		return nil, fmt.Errorf("organization is disabled")
	}
	if err := uc.repo.EnsureUserInDefaultOrganization(ctx, userID); err != nil {
		return nil, err
	}
	belongs, err := uc.repo.UserBelongsToOrganization(ctx, userID, organizationID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, fmt.Errorf("user is not a member of organization")
	}
	if err := uc.repo.SwitchCurrentOrganization(ctx, userID, organizationID); err != nil {
		return nil, err
	}
	return &v1.CurrentOrganizationReply{
		Current:               myOrganizationToReply(organizationItem, true),
		CurrentOrganizationId: organizationID,
	}, nil
}

func (uc *AdminUsecase) GetRoleList(ctx context.Context, req *v1.RolePageParams) (*v1.GetRoleListByPageReply, error) {
	if req == nil {
		req = &v1.RolePageParams{}
	}
	organizationID, err := uc.resolveOrganizationID(ctx, req.GetOrganizationId())
	if err != nil {
		return nil, err
	}
	req.OrganizationId = organizationID
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
	if err := uc.normalizeRoleOrganization(ctx, req); err != nil {
		return nil, err
	}
	if err := uc.normalizeRoleDataScope(ctx, req); err != nil {
		return nil, err
	}
	if err := uc.normalizeSiteMessageRoleRequest(ctx, req); err != nil {
		return nil, err
	}
	if req.GetValue() == "root" {
		platformRoot, err := uc.isPlatformRoot(ctx)
		if err != nil {
			return nil, err
		}
		if !platformRoot {
			return nil, fmt.Errorf("root role can only be created by platform root")
		}
	}
	if err := uc.validateRolePermissionScope(ctx, req.GetOrganizationId(), req.GetPermissions(), req.GetApiPermissions()); err != nil {
		return nil, err
	}
	if err := uc.validateRoleGrantable(ctx, req.GetOrganizationId(), req.GetPermissions(), req.GetApiPermissions()); err != nil {
		return nil, err
	}
	if err := uc.validateRoleDataScopeGrantable(ctx, req.GetOrganizationId(), req.GetValue(), req.GetDataScope(), req.GetDataScopeDeptIds()); err != nil {
		return nil, err
	}
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
	if err := uc.normalizeSiteMessageRoleRequest(ctx, req); err != nil {
		return nil, err
	}
	if err := uc.normalizeRoleOrganization(ctx, req); err != nil {
		return nil, err
	}
	if err := uc.normalizeRoleDataScope(ctx, req); err != nil {
		return nil, err
	}
	before, err := uc.loadRoleWithResources(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if before.OrganizationID != req.GetOrganizationId() {
		return nil, fmt.Errorf("role organization cannot be changed")
	}
	if before.Value == "root" {
		platformRoot, err := uc.isPlatformRoot(ctx)
		if err != nil {
			return nil, err
		}
		if !platformRoot {
			return nil, fmt.Errorf("root role can only be updated by platform root")
		}
	}
	if before.Value != "root" && req.GetValue() == "root" {
		return nil, fmt.Errorf("role cannot be upgraded to root")
	}
	if err := uc.validateRolePermissionScope(ctx, req.GetOrganizationId(), req.GetPermissions(), req.GetApiPermissions()); err != nil {
		return nil, err
	}
	if err := uc.validateRoleGrantable(ctx, req.GetOrganizationId(), req.GetPermissions(), req.GetApiPermissions()); err != nil {
		return nil, err
	}
	if err := uc.validateRoleDataScopeGrantable(ctx, req.GetOrganizationId(), req.GetValue(), req.GetDataScope(), req.GetDataScopeDeptIds()); err != nil {
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
	if err := uc.ensureUniqueAPIPathMethod(ctx, req.Id, req.Path, req.Method); err != nil {
		return nil, err
	}
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
	if err := uc.ensureUniqueAPIPathMethod(ctx, req.Id, req.Path, req.Method); err != nil {
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

func (uc *AdminUsecase) ensureUniqueAPIPathMethod(ctx context.Context, currentID, path, method string) error {
	path = strings.TrimSpace(path)
	method = strings.TrimSpace(method)
	if path == "" || method == "" {
		return nil
	}
	list, _, err := uc.repo.GetApiList(ctx, &v1.GetApiPageParams{Path: path, Method: method})
	if err != nil {
		return err
	}
	for _, item := range list {
		if item == nil {
			continue
		}
		if currentID == "" || item.ID != currentID {
			return fmt.Errorf("api path + method already exists: %s %s", method, path)
		}
	}
	return nil
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
	if err := uc.RegisterPermissionSnapshot(ctx); err != nil {
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
	if err := uc.RegisterPermissionSnapshot(ctx); err != nil {
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
	return uc.RegisterPermissionSnapshot(ctx)
}

func (uc *AdminUsecase) GetDeptList(ctx context.Context, req *v1.GetDeptListParams) (*v1.GetDeptListReply, error) {
	organizationID, err := uc.resolveOrganizationID(ctx, req.GetOrganizationId())
	if err != nil {
		return nil, err
	}
	deptList, err := uc.repo.GetDeptList(ctx, organizationID)
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
	organizationID, err := uc.resolveOrganizationID(ctx, req.GetOrganizationId())
	if err != nil {
		return nil, err
	}
	if err := uc.ensureDeptParentInOrganization(ctx, req.GetPid(), organizationID); err != nil {
		return nil, err
	}
	req.OrganizationId = organizationID
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
	if req.GetOrganizationId() == "" {
		dept, err := uc.repo.GetDeptById(ctx, deptID)
		if err != nil {
			return nil, err
		}
		req.OrganizationId = dept.OrganizationID
	} else if _, err := uc.resolveOrganizationID(ctx, req.GetOrganizationId()); err != nil {
		return nil, err
	}
	organizationID := req.GetOrganizationId()
	if err := uc.ensureDeptParentInOrganization(ctx, req.GetPid(), organizationID); err != nil {
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
			if err := uc.repo.RemoveDeptBindings(ctx, child.ID); err != nil {
				return err
			}
			if err := uc.repo.DelDept(ctx, child.ID); err != nil {
				return err
			}
		}
	}
	if err := uc.repo.RemoveDeptBindings(ctx, dept.ID); err != nil {
		return err
	}
	return uc.repo.DelDept(ctx, id)
}

func (uc *AdminUsecase) GetCurrentUserMenus(ctx context.Context, userID string) (*v1.GetCurrentUserMenusReply, error) {
	_, err := uc.repo.GetUserAuthInfo(ctx, userID)
	if err != nil {
		return nil, err
	}
	organizationID, err := uc.repo.GetCurrentOrganizationID(ctx, userID)
	if err != nil {
		return nil, err
	}
	bindings, err := uc.repo.GetUserRoleBindings(ctx, userID, organizationID)
	if err != nil {
		return nil, err
	}
	if len(bindings) == 0 {
		defaultOrganizationID, err := uc.repo.GetDefaultOrganizationID(ctx)
		if err != nil {
			return nil, err
		}
		if organizationID != defaultOrganizationID {
			bindings, err = uc.repo.GetUserRoleBindings(ctx, userID, defaultOrganizationID)
			if err != nil {
				return nil, err
			}
		}
	}
	isRoot := false
	menuIDs := make([]int32, 0)
	seenMenuIDs := make(map[int32]struct{})
	for _, binding := range bindings {
		roleItem, err := uc.repo.GetRole(ctx, binding.RoleID)
		if err != nil {
			return nil, err
		}
		if roleItem.Value == "root" {
			isRoot = true
		}
		for _, menuID := range roleItem.Menus {
			if _, ok := seenMenuIDs[menuID]; ok {
				continue
			}
			seenMenuIDs[menuID] = struct{}{}
			menuIDs = append(menuIDs, menuID)
		}
	}
	authority := &currentUserMenuAuthority{
		isRoot:  isRoot,
		menuIDs: menuIDs,
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

func (uc *AdminUsecase) GetServiceRegistryList(ctx context.Context) (*v1.GetServiceRegistryListReply, error) {
	list, err := uc.repo.ListServiceRegistries(ctx)
	if err != nil {
		return nil, err
	}
	res := &v1.GetServiceRegistryListReply{Items: make([]*v1.ServiceRegistryItem, 0, len(list)), Total: int64(len(list))}
	for _, item := range list {
		res.Items = append(res.Items, serviceRegistryToReply(item))
	}
	return res, nil
}

func (uc *AdminUsecase) GetProjectionSourceStatusList(ctx context.Context) (*v1.GetProjectionSourceStatusListReply, error) {
	list, err := uc.repo.ListProjectionSourceStatuses(ctx)
	if err != nil {
		return nil, err
	}
	res := &v1.GetProjectionSourceStatusListReply{Items: make([]*v1.ProjectionSourceStatusItem, 0, len(list)), Total: int64(len(list))}
	for _, item := range list {
		res.Items = append(res.Items, projectionSourceStatusToReply(item))
	}
	return res, nil
}

func (uc *AdminUsecase) resolveOrganizationID(ctx context.Context, idValue string) (string, error) {
	actorUserID := strings.TrimSpace(authx.UserID(ctx))
	if strings.TrimSpace(idValue) == "" {
		if actorUserID != "" {
			return uc.repo.GetCurrentOrganizationID(ctx, actorUserID)
		}
		return uc.repo.GetDefaultOrganizationID(ctx)
	}
	id, err := normalizeOrganizationID(idValue)
	if err != nil {
		return "", err
	}
	platformRoot, err := uc.isPlatformRoot(ctx)
	if err != nil {
		return "", err
	}
	if actorUserID != "" && !platformRoot {
		currentOrganizationID, err := uc.repo.GetCurrentOrganizationID(ctx, actorUserID)
		if err != nil {
			return "", err
		}
		if id != currentOrganizationID {
			return "", fmt.Errorf("organization id must match current organization")
		}
	}
	if _, err := uc.repo.GetOrganization(ctx, id); err != nil {
		return "", err
	}
	return id, nil
}

func (uc *AdminUsecase) resolveOrganizationMemberReadID(ctx context.Context, idValue string) (string, error) {
	id, err := normalizeOrganizationID(idValue)
	if err != nil {
		return "", err
	}
	actorUserID := strings.TrimSpace(authx.UserID(ctx))
	platformRoot, err := uc.isPlatformRoot(ctx)
	if err != nil {
		return "", err
	}
	if actorUserID != "" && !platformRoot {
		currentOrganizationID, err := uc.repo.GetCurrentOrganizationID(ctx, actorUserID)
		if err != nil {
			return "", err
		}
		defaultOrganizationID, err := uc.repo.GetDefaultOrganizationID(ctx)
		if err != nil {
			return "", err
		}
		if id != currentOrganizationID && id != defaultOrganizationID {
			return "", fmt.Errorf("organization id must match current or default organization")
		}
	}
	if _, err := uc.repo.GetOrganization(ctx, id); err != nil {
		return "", err
	}
	return id, nil
}

func (uc *AdminUsecase) isPlatformRoot(ctx context.Context) (bool, error) {
	userID := strings.TrimSpace(authx.UserID(ctx))
	if userID == "" {
		return false, nil
	}
	defaultOrganizationID, err := uc.repo.GetDefaultOrganizationID(ctx)
	if err != nil {
		return false, err
	}
	currentOrganizationID, err := uc.repo.GetCurrentOrganizationID(ctx, userID)
	if err != nil {
		return false, err
	}
	if currentOrganizationID != defaultOrganizationID {
		return false, nil
	}
	return uc.userHasRoleInOrganization(ctx, userID, "root", defaultOrganizationID)
}

func (uc *AdminUsecase) userHasRole(ctx context.Context, userID, roleValue string) (bool, error) {
	userID = strings.TrimSpace(userID)
	roleValue = strings.TrimSpace(roleValue)
	if userID == "" || roleValue == "" {
		return false, nil
	}
	bindings, err := uc.repo.ListUserRoleBindings(ctx)
	if err != nil {
		return false, err
	}
	roleIDs := make([]int64, 0, len(bindings))
	seen := make(map[int64]struct{}, len(bindings))
	for _, binding := range bindings {
		if binding == nil || binding.UserID != userID || binding.RoleID <= 0 {
			continue
		}
		if _, ok := seen[binding.RoleID]; ok {
			continue
		}
		seen[binding.RoleID] = struct{}{}
		roleIDs = append(roleIDs, binding.RoleID)
	}
	if len(roleIDs) == 0 {
		return false, nil
	}
	values, err := uc.repo.ResolveRoleValues(ctx, roleIDs)
	if err != nil {
		return false, err
	}
	for _, value := range values {
		if value == roleValue {
			return true, nil
		}
	}
	return false, nil
}

func (uc *AdminUsecase) userHasRoleInOrganization(ctx context.Context, userID, roleValue string, organizationID string) (bool, error) {
	userID = strings.TrimSpace(userID)
	roleValue = strings.TrimSpace(roleValue)
	organizationID = strings.TrimSpace(organizationID)
	if userID == "" || roleValue == "" || organizationID == "" {
		return false, nil
	}
	bindings, err := uc.repo.ListUserRoleBindings(ctx)
	if err != nil {
		return false, err
	}
	roleIDs := make([]int64, 0, len(bindings))
	seen := make(map[int64]struct{}, len(bindings))
	for _, binding := range bindings {
		if binding == nil || binding.UserID != userID || binding.OrganizationID != organizationID || binding.RoleID <= 0 {
			continue
		}
		if _, ok := seen[binding.RoleID]; ok {
			continue
		}
		seen[binding.RoleID] = struct{}{}
		roleIDs = append(roleIDs, binding.RoleID)
	}
	if len(roleIDs) == 0 {
		return false, nil
	}
	values, err := uc.repo.ResolveRoleValues(ctx, roleIDs)
	if err != nil {
		return false, err
	}
	for _, value := range values {
		if value == roleValue {
			return true, nil
		}
	}
	return false, nil
}

func (uc *AdminUsecase) requirePlatformRoot(ctx context.Context, message string) error {
	ok, err := uc.isPlatformRoot(ctx)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return fmt.Errorf("%s", message)
}

func (uc *AdminUsecase) resolveRoleBindingOrganization(ctx context.Context, idValue string) (string, error) {
	organizationID, err := uc.resolveOrganizationID(ctx, idValue)
	if err != nil {
		return "", err
	}
	if _, err := uc.repo.GetOrganization(ctx, organizationID); err != nil {
		return "", err
	}
	return organizationID, nil
}

func (uc *AdminUsecase) normalizeRoleOrganization(ctx context.Context, req *v1.RoleListItem) error {
	if req == nil {
		return nil
	}
	organizationID, err := uc.resolveOrganizationID(ctx, req.GetOrganizationId())
	if err != nil {
		return err
	}
	req.OrganizationId = organizationID
	return nil
}

func (uc *AdminUsecase) normalizeRoleDataScope(ctx context.Context, req *v1.RoleListItem) error {
	if req == nil {
		return nil
	}
	scope := strings.TrimSpace(req.GetDataScope())
	if scope == "" {
		scope = defaultRoleDataScope(req.GetValue())
	}
	if !isValidRoleDataScope(scope) {
		return fmt.Errorf("invalid role data_scope: %s", scope)
	}
	req.DataScope = scope
	if scope != "custom_depts" {
		req.DataScopeDeptIds = nil
		return nil
	}
	seen := make(map[int64]struct{}, len(req.GetDataScopeDeptIds()))
	deptIDs := make([]int64, 0, len(req.GetDataScopeDeptIds()))
	for _, deptID := range req.GetDataScopeDeptIds() {
		if deptID <= 0 {
			continue
		}
		if _, ok := seen[deptID]; ok {
			continue
		}
		dept, err := uc.repo.GetDeptById(ctx, deptID)
		if err != nil {
			return err
		}
		if dept.OrganizationID != req.GetOrganizationId() {
			return fmt.Errorf("data scope department belongs to another organization")
		}
		seen[deptID] = struct{}{}
		deptIDs = append(deptIDs, deptID)
	}
	req.DataScopeDeptIds = deptIDs
	return nil
}

func isValidRoleDataScope(scope string) bool {
	switch scope {
	case "all", "self_dept", "self_dept_and_child", "self", "custom_depts":
		return true
	default:
		return false
	}
}

func defaultRoleDataScope(roleValue string) string {
	switch strings.TrimSpace(roleValue) {
	case "root", "admin":
		return "all"
	default:
		return "self"
	}
}

func (uc *AdminUsecase) ensureDeptParentInOrganization(ctx context.Context, pid string, organizationID string) error {
	parentID, err := strconv.ParseInt(pid, 10, 64)
	if err != nil || parentID <= 0 {
		return nil
	}
	parent, err := uc.repo.GetDeptById(ctx, parentID)
	if err != nil {
		return err
	}
	return validateDeptParentOrganization(parent, organizationID)
}

func validateDeptParentOrganization(parent *ent.Dept, organizationID string) error {
	if parent == nil || parent.OrganizationID == organizationID {
		return nil
	}
	return fmt.Errorf("parent department belongs to another organization")
}

func (uc *AdminUsecase) ensureUserOrganizationMember(ctx context.Context, userID string, organizationID string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return fmt.Errorf("user_id is required")
	}
	if strings.TrimSpace(organizationID) == "" {
		return fmt.Errorf("organization_id is required")
	}
	belongs, err := uc.repo.UserBelongsToOrganization(ctx, userID, organizationID)
	if err != nil {
		return err
	}
	if !belongs {
		return fmt.Errorf("user does not belong to organization")
	}
	return nil
}

func (uc *AdminUsecase) validatePermissionScopeRefs(ctx context.Context, menuIDs []int32, resourceIDs []string) error {
	if len(menuIDs) > 0 {
		menuList, err := uc.repo.GetMenuList(ctx)
		if err != nil {
			return err
		}
		existing := make(map[int32]struct{}, len(menuList))
		for _, item := range menuList {
			if item == nil {
				continue
			}
			existing[int32(item.ID)] = struct{}{}
		}
		for _, menuID := range menuIDs {
			if _, ok := existing[menuID]; !ok {
				return fmt.Errorf("menu does not exist: %d", menuID)
			}
		}
	}
	for _, resourceID := range resourceIDs {
		if _, err := uc.repo.GetResource(ctx, resourceID); err != nil {
			return err
		}
	}
	return nil
}

func (uc *AdminUsecase) validateOrganizationPermissionScopeShrink(ctx context.Context, organizationID string, menuIDs []int32, resourceIDs []string) error {
	roleList, err := uc.repo.ListRoles(ctx, &v1.RolePageParams{OrganizationId: organizationID})
	if err != nil {
		return err
	}
	menuAllowed := toInt32Set(menuIDs)
	resourceAllowed := toStringSet(resourceIDs)
	if len(menuAllowed) > 0 {
		for _, roleItem := range roleList {
			if roleItem == nil {
				continue
			}
			for _, menuID := range roleItem.Menus {
				if _, ok := menuAllowed[menuID]; !ok {
					return fmt.Errorf("cannot remove menu %d while role %s is using it", menuID, roleItem.Name)
				}
			}
		}
	}
	if len(resourceAllowed) > 0 {
		for _, roleItem := range roleList {
			if roleItem == nil || roleItem.Edges.Resource == nil {
				continue
			}
			for _, resourceItem := range roleItem.Edges.Resource {
				if resourceItem == nil {
					continue
				}
				if _, ok := resourceAllowed[resourceItem.ID]; !ok {
					return fmt.Errorf("cannot remove resource %s while role %s is using it", resourceItem.ID, roleItem.Name)
				}
			}
		}
	}
	return nil
}

func (uc *AdminUsecase) validateRolePermissionScope(ctx context.Context, organizationID string, menuIDs []int32, resourceIDs []string) error {
	scope, err := uc.GetOrganizationPermissionScope(ctx, &v1.GetOrganizationPermissionScopeRequest{OrganizationId: organizationID})
	if err != nil {
		return err
	}
	allowedMenus := toInt32Set(scope.MenuIds)
	for _, menuID := range normalizeMenuIDs(menuIDs) {
		if _, ok := allowedMenus[menuID]; !ok {
			return fmt.Errorf("menu is outside organization permission scope: %d", menuID)
		}
	}
	allowedResources := toStringSet(scope.ResourceIds)
	for _, resourceID := range normalizeStringIDs(resourceIDs) {
		if _, ok := allowedResources[resourceID]; !ok {
			return fmt.Errorf("resource is outside organization permission scope: %s", resourceID)
		}
	}
	return nil
}

func (uc *AdminUsecase) validateRoleGrantable(ctx context.Context, organizationID string, menuIDs []int32, resourceIDs []string) error {
	grant, err := uc.actorPermissionGrant(ctx, organizationID)
	if err != nil {
		return err
	}
	if grant.isPlatformRoot {
		return nil
	}
	for _, menuID := range normalizeMenuIDs(menuIDs) {
		if _, ok := grant.menuIDs[menuID]; !ok {
			return fmt.Errorf("menu is outside actor grantable scope: %d", menuID)
		}
	}
	for _, resourceID := range normalizeStringIDs(resourceIDs) {
		if _, ok := grant.resourceIDs[resourceID]; !ok {
			return fmt.Errorf("resource is outside actor grantable scope: %s", resourceID)
		}
	}
	return nil
}

func (uc *AdminUsecase) validateUserRoleBindingGrantable(ctx context.Context, organizationID string, roleIDs []int64) error {
	grant, err := uc.actorPermissionGrant(ctx, organizationID)
	if err != nil {
		return err
	}
	if grant.isPlatformRoot {
		return nil
	}
	for _, roleID := range normalizeRoleIDs(roleIDs, 0) {
		roleItem, err := uc.loadRoleWithResources(ctx, roleID)
		if err != nil {
			return err
		}
		for _, menuID := range normalizeMenuIDs(roleItem.Menus) {
			if _, ok := grant.menuIDs[menuID]; !ok {
				return fmt.Errorf("role contains menu outside actor grantable scope: %d", menuID)
			}
		}
		resourceIDs := resourceIDsFromRole(roleItem)
		for _, resourceID := range normalizeStringIDs(resourceIDs) {
			if _, ok := grant.resourceIDs[resourceID]; !ok {
				return fmt.Errorf("role contains resource outside actor grantable scope: %s", resourceID)
			}
		}
		if err := validateDataScopeWithinGrant(grant, roleItem.Value, roleItem.DataScope, roleItem.DataScopeDeptIds); err != nil {
			return fmt.Errorf("role contains %w", err)
		}
	}
	return nil
}

func (uc *AdminUsecase) validateRoleDataScopeGrantable(ctx context.Context, organizationID string, roleValue string, dataScope string, deptIDs []int64) error {
	grant, err := uc.actorPermissionGrant(ctx, organizationID)
	if err != nil {
		return err
	}
	return validateDataScopeWithinGrant(grant, roleValue, dataScope, deptIDs)
}

func validateDataScopeWithinGrant(grant *actorPermissionGrant, roleValue string, dataScope string, deptIDs []int64) error {
	if grant == nil {
		return fmt.Errorf("data_scope is outside actor grantable scope")
	}
	if grant.isPlatformRoot {
		return nil
	}
	scope := strings.TrimSpace(dataScope)
	if scope == "" {
		scope = defaultRoleDataScope(roleValue)
	}
	if len(grant.dataScopes) == 0 {
		return fmt.Errorf("data_scope is outside actor grantable scope: %s", scope)
	}
	switch scope {
	case "all":
		if grant.hasDataScope("all") {
			return nil
		}
	case "self_dept_and_child":
		if grant.hasAnyDataScope("all", "self_dept_and_child") {
			return nil
		}
	case "self_dept":
		if grant.hasAnyDataScope("all", "self_dept_and_child", "self_dept") {
			return nil
		}
	case "self":
		return nil
	case "custom_depts":
		if grant.hasDataScope("all") {
			return nil
		}
		for _, deptID := range normalizeDataScopeDeptIDs(deptIDs) {
			if _, ok := grant.customDeptIDs[deptID]; !ok {
				return fmt.Errorf("data_scope department is outside actor grantable scope: %d", deptID)
			}
		}
		return nil
	}
	return fmt.Errorf("data_scope is outside actor grantable scope: %s", scope)
}

func (grant *actorPermissionGrant) hasDataScope(scope string) bool {
	if grant == nil || grant.dataScopes == nil {
		return false
	}
	_, ok := grant.dataScopes[scope]
	return ok
}

func (grant *actorPermissionGrant) hasAnyDataScope(scopes ...string) bool {
	for _, scope := range scopes {
		if grant.hasDataScope(scope) {
			return true
		}
	}
	return false
}

func (uc *AdminUsecase) actorPermissionGrant(ctx context.Context, organizationID string) (*actorPermissionGrant, error) {
	actorUserID := strings.TrimSpace(authx.UserID(ctx))
	platformRoot, err := uc.isPlatformRoot(ctx)
	if err != nil {
		return nil, err
	}
	if platformRoot {
		return &actorPermissionGrant{isPlatformRoot: true}, nil
	}
	if actorUserID == "" {
		return &actorPermissionGrant{
			menuIDs:       map[int32]struct{}{},
			resourceIDs:   map[string]struct{}{},
			dataScopes:    map[string]struct{}{},
			customDeptIDs: map[int64]struct{}{},
		}, nil
	}
	roles, err := uc.repo.ListUserOrganizationRoles(ctx, actorUserID, organizationID)
	if err != nil {
		return nil, err
	}
	grant := &actorPermissionGrant{
		menuIDs:       map[int32]struct{}{},
		resourceIDs:   map[string]struct{}{},
		dataScopes:    map[string]struct{}{},
		customDeptIDs: map[int64]struct{}{},
	}
	for _, roleItem := range roles {
		if roleItem == nil {
			continue
		}
		for _, menuID := range normalizeMenuIDs(roleItem.Menus) {
			grant.menuIDs[menuID] = struct{}{}
		}
		for _, resourceID := range normalizeStringIDs(resourceIDsFromRole(roleItem)) {
			grant.resourceIDs[resourceID] = struct{}{}
		}
		scope := strings.TrimSpace(roleItem.DataScope)
		if scope == "" {
			scope = defaultRoleDataScope(roleItem.Value)
		}
		grant.dataScopes[scope] = struct{}{}
		if scope == "custom_depts" {
			for _, deptID := range normalizeDataScopeDeptIDs(roleItem.DataScopeDeptIds) {
				grant.customDeptIDs[deptID] = struct{}{}
			}
		}
	}
	return grant, nil
}

func normalizeDataScopeDeptIDs(deptIDs []int64) []int64 {
	seen := make(map[int64]struct{}, len(deptIDs))
	items := make([]int64, 0, len(deptIDs))
	for _, deptID := range deptIDs {
		if deptID <= 0 {
			continue
		}
		if _, ok := seen[deptID]; ok {
			continue
		}
		seen[deptID] = struct{}{}
		items = append(items, deptID)
	}
	sort.Slice(items, func(i, j int) bool { return items[i] < items[j] })
	return items
}

func (uc *AdminUsecase) normalizeDeptBindingDeptID(value string) (int64, error) {
	idValue := strings.TrimSpace(value)
	if idValue == "" {
		return 0, nil
	}
	return tools.DeptStrSplitToInt(idValue)
}

func diffRemovedOrganizationMembers(
	existingMembers []*v1.OrganizationMemberItem,
	nextUserIDs []string,
) []string {
	nextSet := make(map[string]struct{}, len(nextUserIDs))
	for _, userID := range nextUserIDs {
		userID = strings.TrimSpace(userID)
		if userID == "" {
			continue
		}
		nextSet[userID] = struct{}{}
	}
	removed := make([]string, 0)
	for _, member := range existingMembers {
		userID := strings.TrimSpace(member.GetUserId())
		if userID == "" {
			continue
		}
		if _, ok := nextSet[userID]; ok {
			continue
		}
		removed = append(removed, userID)
	}
	return removed
}

func organizationScopesToReply(organizationID string, scopes []*ent.OrganizationPermissionScope) *v1.OrganizationPermissionScopeReply {
	menuIDs := make([]int32, 0)
	resourceIDs := make([]string, 0)
	seenMenuIDs := make(map[int32]struct{})
	seenResourceIDs := make(map[string]struct{})
	for _, item := range scopes {
		if item == nil {
			continue
		}
		switch item.PermissionType {
		case "menu":
			id, err := strconv.ParseInt(item.PermissionRef, 10, 32)
			if err != nil || id <= 0 {
				continue
			}
			menuID := int32(id)
			if _, ok := seenMenuIDs[menuID]; ok {
				continue
			}
			seenMenuIDs[menuID] = struct{}{}
			menuIDs = append(menuIDs, menuID)
		case "resource":
			resourceID := strings.TrimSpace(item.PermissionRef)
			if resourceID == "" {
				continue
			}
			if _, ok := seenResourceIDs[resourceID]; ok {
				continue
			}
			seenResourceIDs[resourceID] = struct{}{}
			resourceIDs = append(resourceIDs, resourceID)
		}
	}
	sort.Slice(menuIDs, func(i, j int) bool { return menuIDs[i] < menuIDs[j] })
	sort.Strings(resourceIDs)
	return &v1.OrganizationPermissionScopeReply{
		OrganizationId: organizationID,
		MenuIds:        menuIDs,
		ResourceIds:    resourceIDs,
	}
}

func normalizeMenuIDs(menuIDs []int32) []int32 {
	seen := make(map[int32]struct{}, len(menuIDs))
	items := make([]int32, 0, len(menuIDs))
	for _, id := range menuIDs {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		items = append(items, id)
	}
	sort.Slice(items, func(i, j int) bool { return items[i] < items[j] })
	return items
}

func normalizeStringIDs(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	items := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		items = append(items, id)
	}
	sort.Strings(items)
	return items
}

func toInt32Set(ids []int32) map[int32]struct{} {
	items := make(map[int32]struct{}, len(ids))
	for _, id := range ids {
		if id > 0 {
			items[id] = struct{}{}
		}
	}
	return items
}

func toStringSet(ids []string) map[string]struct{} {
	items := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" {
			items[id] = struct{}{}
		}
	}
	return items
}

func filterMenusByScope(menuList []*ent.Menu, menuIDs []int32) []*ent.Menu {
	allowed := toInt32Set(menuIDs)
	menuByID := make(map[int64]*ent.Menu, len(menuList))
	for _, item := range menuList {
		if item == nil {
			continue
		}
		menuByID[item.ID] = item
	}
	for _, menuID := range menuIDs {
		item, ok := menuByID[int64(menuID)]
		if !ok {
			continue
		}
		pid := item.Pid
		for pid > 0 {
			parent, ok := menuByID[pid]
			if !ok {
				break
			}
			allowed[int32(parent.ID)] = struct{}{}
			pid = parent.Pid
		}
	}
	items := make([]*ent.Menu, 0, len(menuList))
	for _, item := range menuList {
		if item == nil {
			continue
		}
		if _, ok := allowed[int32(item.ID)]; !ok {
			continue
		}
		items = append(items, item)
	}
	sortMenus(items)
	return items
}

func filterResourcesByScope(resourceList []*ent.Resource, resourceIDs []string) []*ent.Resource {
	allowed := toStringSet(resourceIDs)
	items := make([]*ent.Resource, 0, len(resourceList))
	for _, item := range resourceList {
		if item == nil {
			continue
		}
		if _, ok := allowed[item.ID]; !ok {
			continue
		}
		items = append(items, item)
	}
	return items
}

func normalizeOrganizationID(idValue string) (string, error) {
	id := strings.TrimSpace(idValue)
	if id == "" {
		return "", fmt.Errorf("organization id is required")
	}
	return id, nil
}

func (uc *AdminUsecase) organizationToReply(ctx context.Context, item *ent.Organization) (*v1.OrganizationItem, error) {
	memberCount, err := uc.repo.CountOrganizationMembers(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	deptCount, err := uc.repo.CountOrganizationDepts(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	status := int32(0)
	if item.Status {
		status = 1
	}
	return &v1.OrganizationItem{
		Id:          item.ID,
		Name:        item.Name,
		Code:        item.Code,
		OrderNo:     item.Sort,
		Remark:      item.Desc,
		Status:      status,
		CreateTime:  item.CreateTime.Format(time.DateTime),
		UpdateTime:  item.UpdateTime.Format(time.DateTime),
		MemberCount: memberCount,
		DeptCount:   deptCount,
	}, nil
}

func myOrganizationToReply(item *ent.Organization, current bool) *v1.MyOrganizationItem {
	status := int32(0)
	if item.Status {
		status = 1
	}
	return &v1.MyOrganizationItem{
		Id:         item.ID,
		Name:       item.Name,
		Code:       item.Code,
		OrderNo:    item.Sort,
		Remark:     item.Desc,
		Status:     status,
		Current:    current,
		CreateTime: item.CreateTime.Format(time.DateTime),
		UpdateTime: item.UpdateTime.Format(time.DateTime),
	}
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
			Id:             id,
			Pid:            strPid,
			Name:           item.Value.Name,
			OrderNo:        item.Value.Sort,
			Remark:         item.Value.Desc,
			Status:         status,
			CreateTime:     item.Value.CreateTime.Format(time.DateTime),
			OrganizationId: item.Value.OrganizationID,
			Children:       toDeptTree(item.Children, id),
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
		Id:             strconv.FormatInt(dept.ID, 10),
		Pid:            pid,
		Name:           dept.Name,
		OrderNo:        dept.Sort,
		Remark:         dept.Desc,
		Status:         status,
		CreateTime:     dept.CreateTime.Format(time.DateTime),
		OrganizationId: dept.OrganizationID,
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
		OrganizationId: item.OrganizationID,
		DataScope:      item.DataScope,
		DataScopeDeptIds: append([]int64(nil),
			item.DataScopeDeptIds...),
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

func normalizeRoleIDs(roleIDs []int64, fallbackRoleID int64) []int64 {
	if len(roleIDs) == 0 && fallbackRoleID > 0 {
		roleIDs = []int64{fallbackRoleID}
	}
	items := make([]int64, 0, len(roleIDs))
	seen := make(map[int64]struct{}, len(roleIDs))
	for _, roleID := range roleIDs {
		if roleID <= 0 {
			continue
		}
		if _, ok := seen[roleID]; ok {
			continue
		}
		seen[roleID] = struct{}{}
		items = append(items, roleID)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i] < items[j]
	})
	return items
}

func (uc *AdminUsecase) normalizeUserRoleBindingIDs(_ context.Context, roleIDs []int64) ([]int64, error) {
	return normalizeRoleIDs(roleIDs, 0), nil
}

func (uc *AdminUsecase) getDefaultUserRoleID(ctx context.Context) (int64, error) {
	defaultOrganizationID, err := uc.repo.GetDefaultOrganizationID(ctx)
	if err != nil {
		return 0, err
	}
	roleList, err := uc.repo.ListRoles(ctx, &v1.RolePageParams{
		OrganizationId: defaultOrganizationID,
		PageSize:       1000,
	})
	if err != nil {
		return 0, err
	}
	for _, item := range roleList {
		if item != nil && item.Status && item.Value == "default" {
			return item.ID, nil
		}
	}
	return 0, fmt.Errorf("default role not found in default organization")
}

func userRoleBindingToReply(item *ent.UserRoleBinding) *v1.UserRoleBindingItem {
	roleIDs := []int64(nil)
	if item.RoleID >= 0 {
		roleIDs = []int64{item.RoleID}
	}
	reply := &v1.UserRoleBindingItem{
		OrganizationId: item.OrganizationID,
		UserId:         item.UserID,
		RoleId:         item.RoleID,
		RoleIds:        roleIDs,
	}
	if item.ID > 0 {
		reply.Id = strconv.FormatInt(item.ID, 10)
	}
	if !item.CreateTime.IsZero() {
		reply.CreateTime = item.CreateTime.Format(time.DateTime)
	}
	if !item.UpdateTime.IsZero() {
		reply.UpdateTime = item.UpdateTime.Format(time.DateTime)
	}
	return reply
}

func userRoleBindingsToReply(userID string, organizationID string, items []*ent.UserRoleBinding) *v1.UserRoleBindingItem {
	res := &v1.UserRoleBindingItem{
		OrganizationId: organizationID,
		UserId:         userID,
		RoleIds:        make([]int64, 0, len(items)),
	}
	roleIDSet := false
	for _, item := range items {
		if item == nil {
			continue
		}
		if res.Id == "" {
			if item.ID > 0 {
				res.Id = strconv.FormatInt(item.ID, 10)
			}
			if !item.CreateTime.IsZero() {
				res.CreateTime = item.CreateTime.Format(time.DateTime)
			}
			if !item.UpdateTime.IsZero() {
				res.UpdateTime = item.UpdateTime.Format(time.DateTime)
			}
		}
		if res.UserId == "" {
			res.UserId = item.UserID
		}
		if res.OrganizationId == "" {
			res.OrganizationId = item.OrganizationID
		}
		if !roleIDSet {
			res.RoleId = item.RoleID
			roleIDSet = true
		}
		res.RoleIds = append(res.RoleIds, item.RoleID)
	}
	return res
}

func (uc *AdminUsecase) userDeptBindingToReply(ctx context.Context, item *ent.UserDeptMembership) (*v1.UserDeptBindingItem, error) {
	if item == nil {
		return &v1.UserDeptBindingItem{}, nil
	}
	reply := &v1.UserDeptBindingItem{
		Id:             strconv.FormatInt(item.ID, 10),
		UserId:         item.UserID,
		DeptId:         strconv.FormatInt(item.DeptID, 10),
		OrganizationId: item.OrganizationID,
	}
	if !item.CreateTime.IsZero() {
		reply.CreateTime = item.CreateTime.Format(time.DateTime)
	}
	if !item.UpdateTime.IsZero() {
		reply.UpdateTime = item.UpdateTime.Format(time.DateTime)
	}
	deptItem, err := uc.repo.GetDeptById(ctx, item.DeptID)
	if err != nil {
		if ent.IsNotFound(err) {
			return reply, nil
		}
		return nil, err
	}
	reply.DeptName = deptItem.Name
	return reply, nil
}

func serviceRegistryToReply(item *ent.ServiceRegistry) *v1.ServiceRegistryItem {
	status := int32(0)
	if item.Status {
		status = 1
	}
	return &v1.ServiceRegistryItem{
		Id:                item.ID,
		ServiceCode:       item.ServiceCode,
		ServiceName:       item.ServiceName,
		HttpPrefix:        item.HTTPPrefix,
		GrpcService:       item.GrpcService,
		Status:            status,
		ProjectionEnabled: item.ProjectionEnabled,
		Description:       item.Description,
		CreateTime:        item.CreateTime.Format(time.DateTime),
	}
}

func projectionSourceStatusToReply(item *ent.ProjectionSourceStatus) *v1.ProjectionSourceStatusItem {
	return &v1.ProjectionSourceStatusItem{
		Id:                   item.ID,
		SourceService:        item.SourceService,
		SyncMode:             item.SyncMode,
		State:                item.State,
		LastSnapshotRevision: item.LastSnapshotRevision,
		LastSyncTime:         item.LastSyncTime,
		LastError:            item.LastError,
		Description:          item.Description,
		CreateTime:           item.CreateTime.Format(time.DateTime),
	}
}

var _ = emptypb.Empty{}
var _ = authx.UserID
