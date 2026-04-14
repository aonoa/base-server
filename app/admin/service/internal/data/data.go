package data

import (
	"context"
	dbsql "database/sql"
	"regexp"
	"strconv"
	"strings"
	"time"

	"ariga.io/entcache"
	v1 "base-server/api/gen/go/admin/service/v1"
	authv1 "base-server/api/gen/go/auth/service/v1"
	commonv1 "base-server/api/gen/go/common/service/v1"
	userv1 "base-server/api/gen/go/user/service/v1"
	"base-server/app/admin/service/internal/biz"
	"base-server/app/admin/service/internal/conf"
	"base-server/pkg/authx"
	"base-server/pkg/data/ent"
	"base-server/pkg/data/ent/apiresources"
	"base-server/pkg/data/ent/dept"
	"base-server/pkg/data/ent/menu"
	"base-server/pkg/data/ent/migrate"
	"base-server/pkg/data/ent/resource"
	"base-server/pkg/data/ent/role"
	"base-server/pkg/data/ent/syslogrecord"
	"base-server/pkg/data/ent/userrolebinding"
	"base-server/pkg/tools"

	"entgo.io/ent/dialect"
	sql "entgo.io/ent/dialect/sql"
	schema "entgo.io/ent/dialect/sql/schema"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewAdminRepo)

// Data .
type Data struct {
	db           *ent.Client
	userConn     *grpc.ClientConn
	userClient   userv1.UserServiceClient
	authConn     *grpc.ClientConn
	authClient   authv1.AuthServiceClient
	commonConn   *grpc.ClientConn
	commonClient commonv1.CommonServiceClient
}

// NewData .
func NewData(c *conf.Data, services *conf.Services, logger log.Logger) (*Data, func(), error) {
	helper := log.NewHelper(logger)
	db, err := dbsql.Open(c.Database.Driver, c.Database.Source)
	if err != nil {
		return nil, nil, err
	}
	drv := sql.OpenDB(toEntDialect(c.Database.Driver), db)
	sqlDrv := dialect.DebugWithContext(drv, func(ctx context.Context, args ...interface{}) {
		helper.WithContext(ctx).Info(args...)
	})
	client := ent.NewClient(ent.Driver(sqlDrv))
	if err := migrate.Create(context.Background(), migrate.NewSchema(sqlDrv), adminTables(), schema.WithForeignKeys(false)); err != nil {
		return nil, nil, err
	}
	userConn, err := grpc.NewClient(services.User.GrpcEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = client.Close()
		return nil, nil, err
	}
	authConn, err := grpc.NewClient(services.Auth.GrpcEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = userConn.Close()
		_ = client.Close()
		return nil, nil, err
	}
	commonConn, err := grpc.NewClient(services.Common.GrpcEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = authConn.Close()
		_ = userConn.Close()
		_ = client.Close()
		return nil, nil, err
	}
	d := &Data{
		db:           client,
		userConn:     userConn,
		userClient:   userv1.NewUserServiceClient(userConn),
		authConn:     authConn,
		authClient:   authv1.NewAuthServiceClient(authConn),
		commonConn:   commonConn,
		commonClient: commonv1.NewCommonServiceClient(commonConn),
	}
	cleanup := func() {
		helper.Info("closing the data resources")
		if d.commonConn != nil {
			_ = d.commonConn.Close()
		}
		if d.authConn != nil {
			_ = d.authConn.Close()
		}
		if d.userConn != nil {
			_ = d.userConn.Close()
		}
		_ = d.db.Close()
	}
	return d, cleanup, nil
}

type adminRepo struct {
	data *Data
	log  *log.Helper
}

func NewAdminRepo(data *Data, logger log.Logger) biz.AdminRepo {
	return &adminRepo{data: data, log: log.NewHelper(logger)}
}

func (r *adminRepo) ListAllRoles(ctx context.Context) ([]*ent.Role, error) {
	return r.data.db.Role.Query().
		Where(role.StatusEQ(true)).
		WithResource().
		All(ctx)
}

func (r *adminRepo) ListRoles(ctx context.Context, req *v1.RolePageParams) ([]*ent.Role, error) {
	query := r.data.db.Role.Query()
	if req.Name != "" {
		query = query.Where(role.NameEQ(req.Name))
	}
	if req.Status == 1 {
		query = query.Where(role.StatusEQ(true))
	}
	query.WithResource(func(query *ent.ResourceQuery) {
		query.Select(resource.FieldID, resource.FieldType, resource.FieldValue, resource.FieldMethod)
	})
	return query.All(ctx)
}

func (r *adminRepo) GetRole(ctx context.Context, id int64) (*ent.Role, error) {
	return r.data.db.Role.Query().Where(role.IDEQ(id)).First(ctx)
}

func (r *adminRepo) ResolveRoleValues(ctx context.Context, roleIDs []int64) (map[int64]string, error) {
	roles, err := r.data.db.Role.Query().
		Where(role.IDIn(roleIDs...)).
		Select(role.FieldID, role.FieldValue).
		All(ctx)
	if err != nil {
		return nil, err
	}
	values := make(map[int64]string, len(roles))
	for _, item := range roles {
		values[item.ID] = item.Value
	}
	return values, nil
}

func (r *adminRepo) GetUserRoleBinding(ctx context.Context, userID string) (*ent.UserRoleBinding, error) {
	return r.data.db.UserRoleBinding.Query().Where(userrolebinding.UserIDEQ(userID)).First(ctx)
}

func (r *adminRepo) ListUserRoleBindings(ctx context.Context) ([]*ent.UserRoleBinding, error) {
	return r.data.db.UserRoleBinding.Query().All(ctx)
}

func (r *adminRepo) UpsertUserRoleBinding(ctx context.Context, userID string, roleID int64) (*ent.UserRoleBinding, error) {
	current, err := r.data.db.UserRoleBinding.Query().Where(userrolebinding.UserIDEQ(userID)).First(ctx)
	if err == nil {
		return r.data.db.UserRoleBinding.UpdateOneID(current.ID).SetRoleID(roleID).Save(ctx)
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}
	return r.data.db.UserRoleBinding.Create().SetUserID(userID).SetRoleID(roleID).Save(ctx)
}

func (r *adminRepo) DeleteUserRoleBinding(ctx context.Context, userID string) error {
	_, err := r.data.db.UserRoleBinding.Delete().Where(userrolebinding.UserIDEQ(userID)).Exec(ctx)
	return err
}

func (r *adminRepo) AddRole(ctx context.Context, req *v1.RoleListItem) (*ent.Role, error) {
	return r.data.db.Role.Create().
		SetName(req.Name).
		SetValue(req.Value).
		SetStatus(req.Status != 0).
		SetDesc(req.Remark).
		SetMenus(req.Permissions).
		AddResourceIDs(req.ApiPermissions...).
		Save(ctx)
}

func (r *adminRepo) UpdateRole(ctx context.Context, roleID int64, req *v1.RoleListItem) (*ent.Role, error) {
	return r.data.db.Role.UpdateOneID(roleID).
		SetName(req.Name).
		SetValue(req.Value).
		SetStatus(req.Status != 0).
		SetDesc(req.Remark).
		SetMenus(req.Permissions).
		ClearResource().
		AddResourceIDs(req.ApiPermissions...).
		Save(ctx)
}

func (r *adminRepo) DelRole(ctx context.Context, id int64) error {
	return r.data.db.Role.DeleteOneID(id).Exec(ctx)
}

func (r *adminRepo) GetMenuList(ctx context.Context) ([]*ent.Menu, error) {
	return r.data.db.Menu.Query().Order(menu.ByPid(), menu.ByOrder()).All(ctx)
}

func (r *adminRepo) GetUserAuthInfo(ctx context.Context, userID string) (*userv1.GetUserAuthInfoReply, error) {
	ctx = authx.ForwardAuthorizationContext(ctx)
	return r.data.userClient.GetUserAuthInfo(ctx, &userv1.GetUserAuthInfoRequest{UserId: userID})
}

func getAPIListQuery(params *v1.GetApiPageParams, isPage bool) func(s *sql.Selector) {
	return func(s *sql.Selector) {
		if params.Path != "" {
			s.Where(sql.EQ(apiresources.FieldPath, params.Path))
		}
		if params.ResourcesGroup != "" {
			s.Where(sql.EQ(apiresources.FieldResourcesGroup, params.ResourcesGroup))
		}
		if params.Method != "" {
			s.Where(sql.EQ(apiresources.FieldMethod, params.Method))
		}
		if params.Description != "" {
			s.Where(sql.Like(apiresources.FieldDescription, "%"+params.Description+"%"))
		}
		if isPage {
			if params.PageSize != 0 {
				s.Limit(int(params.PageSize))
			}
			if params.CurrentPage != 0 {
				s.Offset(int(tools.GetPageOffset(params.CurrentPage, params.PageSize)))
			}
		}
	}
}

func (r *adminRepo) GetApiList(ctx context.Context, req *v1.GetApiPageParams) ([]*ent.ApiResources, int64, error) {
	list, err := r.data.db.ApiResources.Query().Modify(getAPIListQuery(req, true)).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	count, err := r.data.db.ApiResources.Query().Modify(getAPIListQuery(req, false)).Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	return list, int64(count), nil
}

func (r *adminRepo) GetApi(ctx context.Context, id string) (*ent.ApiResources, error) {
	return r.data.db.ApiResources.Query().Where(apiresources.IDEQ(id)).First(ctx)
}

func (r *adminRepo) AddApi(ctx context.Context, req *ent.ApiResources) (*ent.ApiResources, error) {
	return r.data.db.ApiResources.Create().CreateAll(req).Save(ctx)
}

func (r *adminRepo) UpdateApi(ctx context.Context, req *ent.ApiResources) (*ent.ApiResources, error) {
	return r.data.db.ApiResources.UpdateOneID(req.ID).UpdateAll(req).Save(ctx)
}

func (r *adminRepo) DelApi(ctx context.Context, id string) error {
	return r.data.db.ApiResources.DeleteOneID(id).Exec(ctx)
}

func getResourceListQuery(params *v1.GetResourcePageParams, isPage bool) func(s *sql.Selector) {
	return func(s *sql.Selector) {
		if params.Name != "" {
			s.Where(sql.Like(resource.FieldName, "%"+params.Name+"%"))
		}
		if params.Type != "" {
			s.Where(sql.EQ(resource.FieldType, params.Type))
		}
		if params.Value != "" {
			s.Where(sql.EQ(resource.FieldValue, params.Value))
		}
		if params.Method != "" {
			s.Where(sql.Like(resource.FieldMethod, "%"+params.Method+"%"))
		}
		if params.Description != "" {
			s.Where(sql.Like(resource.FieldDescription, "%"+params.Description+"%"))
		}
		if isPage {
			if params.PageSize != 0 {
				s.Limit(int(params.PageSize))
			}
			if params.CurrentPage != 0 {
				s.Offset(int(tools.GetPageOffset(params.CurrentPage, params.PageSize)))
			}
		}
	}
}

func (r *adminRepo) GetResourceList(ctx context.Context, req *v1.GetResourcePageParams) ([]*ent.Resource, int64, error) {
	list, err := r.data.db.Resource.Query().Modify(getResourceListQuery(req, true)).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	count, err := r.data.db.Resource.Query().Modify(getResourceListQuery(req, false)).Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	return list, int64(count), nil
}

func (r *adminRepo) AddResource(ctx context.Context, req *ent.Resource) (*ent.Resource, error) {
	return r.data.db.Resource.Create().CreateAll(req).Save(ctx)
}

func (r *adminRepo) GetResource(ctx context.Context, id string) (*ent.Resource, error) {
	return r.data.db.Resource.Query().Where(resource.IDEQ(id)).First(ctx)
}

func (r *adminRepo) UpdateResource(ctx context.Context, req *ent.Resource) (*ent.Resource, error) {
	return r.data.db.Resource.UpdateOneID(req.ID).UpdateAll(req).Save(ctx)
}

func (r *adminRepo) DelResource(ctx context.Context, id string) error {
	return r.data.db.Resource.DeleteOneID(id).Exec(ctx)
}

func (r *adminRepo) RegisterPermissionSnapshot(ctx context.Context) error {
	req, err := r.buildPermissionSnapshotRequest(ctx)
	if err != nil {
		return err
	}
	ctx = authx.ForwardAuthorizationContext(ctx)
	_, err = r.data.authClient.RegisterPermissionSnapshot(ctx, req)
	return err
}

func (r *adminRepo) ApplyAuthRoleDelta(ctx context.Context, before, after *ent.Role) error {
	beforeRole, err := r.roleToPolicyRole(ctx, before)
	if err != nil {
		return err
	}
	afterRole, err := r.roleToPolicyRole(ctx, after)
	if err != nil {
		return err
	}
	ctx = authx.ForwardAuthorizationContext(ctx)
	_, err = r.data.authClient.ApplyRoleDelta(ctx, &authv1.ApplyRoleDeltaRequest{
		SourceService: "admin",
		Revision:      uint64(time.Now().UnixMilli()),
		Before:        beforeRole,
		After:         afterRole,
	})
	return err
}

func (r *adminRepo) ApplyAuthApiDelta(ctx context.Context, before, after *ent.ApiResources) error {
	ctx = authx.ForwardAuthorizationContext(ctx)
	_, err := r.data.authClient.ApplyApiDelta(ctx, &authv1.ApplyApiDeltaRequest{
		SourceService: "admin",
		Revision:      uint64(time.Now().UnixMilli()),
		Before:        apiToPolicyAPI(before),
		After:         apiToPolicyAPI(after),
	})
	return err
}

func (r *adminRepo) ApplyAuthUserRoleBindingDelta(ctx context.Context, before, after *ent.UserRoleBinding) error {
	beforeBinding, err := r.bindingToPolicyBinding(ctx, before)
	if err != nil {
		return err
	}
	afterBinding, err := r.bindingToPolicyBinding(ctx, after)
	if err != nil {
		return err
	}
	ctx = authx.ForwardAuthorizationContext(ctx)
	_, err = r.data.authClient.ApplyUserRoleBindingDelta(ctx, &authv1.ApplyUserRoleBindingDeltaRequest{
		SourceService: "admin",
		Revision:      uint64(time.Now().UnixMilli()),
		Before:        beforeBinding,
		After:         afterBinding,
	})
	return err
}

func (r *adminRepo) buildPermissionSnapshotRequest(ctx context.Context) (*authv1.RegisterPermissionSnapshotRequest, error) {
	roleList, err := r.ListAllRoles(ctx)
	if err != nil {
		return nil, err
	}
	apiList, _, err := r.GetApiList(ctx, &v1.GetApiPageParams{})
	if err != nil {
		return nil, err
	}
	bindingList, err := r.ListUserRoleBindings(ctx)
	if err != nil {
		return nil, err
	}
	roleValues, err := r.resolveBindingRoleValues(ctx, bindingList)
	if err != nil {
		return nil, err
	}
	req := &authv1.RegisterPermissionSnapshotRequest{
		SourceService: "admin",
		Revision:      uint64(time.Now().UnixMilli()),
		Roles:         make([]*authv1.PolicyRole, 0, len(roleList)),
		Apis:          make([]*authv1.PolicyApi, 0, len(apiList)),
		Bindings:      make([]*authv1.PolicyUserRoleBinding, 0, len(bindingList)),
	}
	for _, roleItem := range roleList {
		req.Roles = append(req.Roles, roleToPolicyRole(roleItem))
	}
	for _, apiItem := range apiList {
		req.Apis = append(req.Apis, apiToPolicyAPI(apiItem))
	}
	for _, binding := range bindingList {
		req.Bindings = append(req.Bindings, userRoleBindingToPolicyBinding(binding, roleValues[binding.RoleID]))
	}
	return req, nil
}

func (r *adminRepo) resolveBindingRoleValues(ctx context.Context, bindings []*ent.UserRoleBinding) (map[int64]string, error) {
	roleIDs := make([]int64, 0, len(bindings))
	seen := make(map[int64]struct{}, len(bindings))
	for _, binding := range bindings {
		if binding == nil || binding.RoleID == 0 {
			continue
		}
		if _, ok := seen[binding.RoleID]; ok {
			continue
		}
		seen[binding.RoleID] = struct{}{}
		roleIDs = append(roleIDs, binding.RoleID)
	}
	if len(roleIDs) == 0 {
		return map[int64]string{}, nil
	}
	return r.ResolveRoleValues(ctx, roleIDs)
}

func (r *adminRepo) roleToPolicyRole(ctx context.Context, item *ent.Role) (*authv1.PolicyRole, error) {
	if item == nil {
		return nil, nil
	}
	if item.Edges.Resource == nil {
		roleItem, err := r.data.db.Role.Query().Where(role.IDEQ(item.ID)).WithResource().Only(ctx)
		if err != nil {
			return nil, err
		}
		item = roleItem
	}
	return roleToPolicyRole(item), nil
}

func (r *adminRepo) bindingToPolicyBinding(ctx context.Context, item *ent.UserRoleBinding) (*authv1.PolicyUserRoleBinding, error) {
	if item == nil {
		return nil, nil
	}
	roleValue := ""
	if item.RoleID > 0 {
		values, err := r.ResolveRoleValues(ctx, []int64{item.RoleID})
		if err != nil {
			return nil, err
		}
		roleValue = values[item.RoleID]
	}
	return userRoleBindingToPolicyBinding(item, roleValue), nil
}

func roleToPolicyRole(item *ent.Role) *authv1.PolicyRole {
	if item == nil {
		return nil
	}
	resources := make([]*authv1.PolicyRoleResource, 0, len(item.Edges.Resource))
	for _, resourceItem := range item.Edges.Resource {
		if resourceItem == nil {
			continue
		}
		resources = append(resources, &authv1.PolicyRoleResource{
			Id:     resourceItem.ID,
			Type:   resourceItem.Type,
			Value:  resourceItem.Value,
			Method: resourceItem.Method,
		})
	}
	return &authv1.PolicyRole{
		Id:        item.ID,
		Name:      item.Name,
		Value:     item.Value,
		Status:    item.Status,
		Remark:    item.Desc,
		MenuIds:   append([]int32(nil), item.Menus...),
		Resources: resources,
	}
}

func apiToPolicyAPI(item *ent.ApiResources) *authv1.PolicyApi {
	if item == nil {
		return nil
	}
	return &authv1.PolicyApi{
		Id:                item.ID,
		Path:              item.Path,
		Method:            item.Method,
		Description:       item.Description,
		Module:            item.Module,
		ModuleDescription: item.ModuleDescription,
		ResourcesGroup:    item.ResourcesGroup,
	}
}

func userRoleBindingToPolicyBinding(item *ent.UserRoleBinding, roleValue string) *authv1.PolicyUserRoleBinding {
	if item == nil {
		return nil
	}
	return &authv1.PolicyUserRoleBinding{
		Id:         strconv.FormatInt(item.ID, 10),
		UserId:     item.UserID,
		RoleId:     item.RoleID,
		RoleValue:  roleValue,
		CreateTime: item.CreateTime.Format(time.RFC3339),
		UpdateTime: item.UpdateTime.Format(time.RFC3339),
	}
}

func (r *adminRepo) CreateMenu(ctx context.Context, item *ent.Menu) (*ent.Menu, error) {
	defer r.data.db.Menu.Query().All(entcache.Evict(ctx))
	return r.data.db.Menu.Create().CreateAll(item).Save(ctx)
}

func (r *adminRepo) ListAuthWalkRoutes(ctx context.Context) ([]tools.WalkRouteItem, error) {
	ctx = authx.ForwardAuthorizationContext(ctx)
	res, err := r.data.authClient.GetWalkRoute(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	items := make([]tools.WalkRouteItem, 0, len(res.Items))
	for _, item := range res.Items {
		items = append(items, tools.WalkRouteItem{URL: item.Url, Method: item.Method})
	}
	return items, nil
}

func (r *adminRepo) ListUserWalkRoutes(ctx context.Context) ([]tools.WalkRouteItem, error) {
	ctx = authx.ForwardAuthorizationContext(ctx)
	res, err := r.data.userClient.GetWalkRoute(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	items := make([]tools.WalkRouteItem, 0, len(res.Items))
	for _, item := range res.Items {
		items = append(items, tools.WalkRouteItem{URL: item.Url, Method: item.Method})
	}
	return items, nil
}

func (r *adminRepo) ListCommonWalkRoutes(ctx context.Context) ([]tools.WalkRouteItem, error) {
	ctx = authx.ForwardAuthorizationContext(ctx)
	res, err := r.data.commonClient.GetWalkRoute(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	items := make([]tools.WalkRouteItem, 0, len(res.Items))
	for _, item := range res.Items {
		items = append(items, tools.WalkRouteItem{URL: item.Url, Method: item.Method})
	}
	return items, nil
}

func (r *adminRepo) UpdateMenu(ctx context.Context, id int64, item *ent.Menu) (*ent.Menu, error) {
	defer r.data.db.Menu.Query().All(entcache.Evict(ctx))
	return r.data.db.Menu.UpdateOneID(id).UpdateAll(item).Save(ctx)
}

func (r *adminRepo) DeleteMenu(ctx context.Context, id int64) error {
	defer r.data.db.Menu.Query().All(entcache.NewContext(ctx))
	return r.data.db.Menu.DeleteOneID(id).Exec(entcache.Evict(ctx))
}

func (r *adminRepo) GetDeptList(ctx context.Context) ([]*ent.Dept, error) {
	return r.data.db.Dept.Query().Order(dept.ByPid(func(options *sql.OrderTermOptions) {
		options.NullsFirst = true
	})).All(ctx)
}

func (r *adminRepo) AddDept(ctx context.Context, req *v1.DeptListItem) (*ent.Dept, error) {
	cmd := r.data.db.Dept.Create().
		SetName(req.Name).
		SetSort(req.OrderNo).
		SetStatus(req.Status != 0).
		SetDesc(req.Remark).
		SetExtension("").
		SetDom(req.Dom)
	pid, err := strconv.ParseInt(req.Pid, 10, 32)
	if err == nil && pid > 0 {
		cmd = cmd.SetPid(pid)
	}
	return cmd.Save(ctx)
}

func (r *adminRepo) UpdateDept(ctx context.Context, deptID int64, req *v1.DeptListItem) (*ent.Dept, error) {
	cmd := r.data.db.Dept.UpdateOneID(deptID).
		SetName(req.Name).
		SetSort(req.OrderNo).
		SetStatus(req.Status != 0).
		SetDesc(req.Remark).
		SetExtension("")
	pid, err := strconv.ParseInt(req.Pid, 10, 32)
	if err == nil && pid > 0 {
		cmd = cmd.SetPid(pid)
	}
	return cmd.Save(ctx)
}

func (r *adminRepo) DelDept(ctx context.Context, id int64) error {
	return r.data.db.Dept.DeleteOneID(id).Exec(ctx)
}

func (r *adminRepo) GetDeptLeafsChildren(ctx context.Context, id int64) ([]*ent.Dept, error) {
	root, err := r.data.db.Dept.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return root.QueryChildren().Where(dept.Not(dept.HasChildren())).All(ctx)
}

func (r *adminRepo) GetDeptById(ctx context.Context, id int64) (*ent.Dept, error) {
	return r.data.db.Dept.Get(ctx, id)
}

func (r *adminRepo) CreateSysLog(ctx context.Context, item *ent.SysLogRecord) error {
	_, err := r.data.db.SysLogRecord.Create().
		SetUserID(item.UserID).
		SetUserName(item.UserName).
		SetIsLogin(item.IsLogin).
		SetSessionID(item.SessionID).
		SetMethod(item.Method).
		SetPath(item.Path).
		SetRequestTime(item.RequestTime).
		SetIPAddress(item.IPAddress).
		SetIPLocation(item.IPLocation).
		SetLatency(item.Latency).
		SetOs(item.Os).
		SetBrowser(item.Browser).
		SetUserAgent(item.UserAgent).
		SetHeader(item.Header).
		SetGetParams(item.GetParams).
		SetPostData(item.PostData).
		SetResCode(item.ResCode).
		SetReason(item.Reason).
		SetResStatus(item.ResStatus).
		SetStack(item.Stack).
		Save(ctx)
	return err
}

func adminTables() []*schema.Table {
	return []*schema.Table{
		migrate.SysAPIResourcesTable,
		migrate.SysResourcesTable,
		migrate.SysRoleTable,
		migrate.APIResourcesRolesTable,
		migrate.ResourceRolesTable,
		migrate.SysUserRoleBindingTable,
		migrate.SysMenuTable,
		migrate.SysDeptTable,
		migrate.SysLogTable,
	}
}

func getSysLogListQuery(params *v1.GetSysLogListParams, isPage bool) func(s *sql.Selector) {
	return func(s *sql.Selector) {
		if params.IsLogin {
			s.Where(sql.EQ(syslogrecord.FieldIsLogin, true))
		}
		if params.UserName != "" {
			s.Where(sql.EQ(syslogrecord.FieldUserName, params.UserName))
		}
		if params.IpAddress != "" {
			s.Where(sql.Like(syslogrecord.FieldIPAddress, params.IpAddress+"%"))
		}
		if params.SessionId != "" {
			s.Where(sql.EQ(syslogrecord.FieldSessionID, params.SessionId))
		}
		if params.Path != "" {
			if ok, path := replaceBracesIfExists(params.Path); ok {
				s.Where(sql.Like(syslogrecord.FieldPath, path))
			} else {
				s.Where(sql.EQ(syslogrecord.FieldPath, path))
			}
		}
		if params.RequestTimeStart != "" {
			s.Where(sql.GTE(syslogrecord.FieldRequestTime, params.RequestTimeStart))
		}
		if params.RequestTimeEnd != "" {
			s.Where(sql.LTE(syslogrecord.FieldRequestTime, params.RequestTimeEnd))
		}
		if params.Method != "" {
			s.Where(sql.EQ(syslogrecord.FieldMethod, params.Method))
		}
		if params.Latency != 0 {
			if params.Latency > 1000 {
				s.Where(sql.GTE(syslogrecord.FieldLatency, params.Latency))
			} else {
				s.Where(sql.LTE(syslogrecord.FieldLatency, params.Latency))
			}
		}
		if isPage {
			s.OrderBy(sql.Desc(syslogrecord.FieldCreateTime))
			if params.PageSize != 0 {
				s.Limit(int(params.PageSize))
			}
			if params.CurrentPage != 0 {
				s.Offset(int(tools.GetPageOffset(params.CurrentPage, params.PageSize)))
			}
		}
	}
}

func (r *adminRepo) GetSysLogList(ctx context.Context, req *v1.GetSysLogListParams) ([]*ent.SysLogRecord, int64, error) {
	res, err := r.data.db.SysLogRecord.Query().Modify(getSysLogListQuery(req, true)).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	count, err := r.data.db.SysLogRecord.Query().Modify(getSysLogListQuery(req, false)).Count(ctx)
	return res, int64(count), err
}

func (r *adminRepo) GetSysLogInfo(ctx context.Context, id string) (*ent.SysLogRecord, error) {
	return r.data.db.SysLogRecord.Query().Where(syslogrecord.IDEQ(id)).First(ctx)
}

func replaceBracesIfExists(str string) (bool, string) {
	hasBraces := strings.Contains(str, "{") || strings.Contains(str, "}")
	if !hasBraces {
		return false, str
	}
	re := regexp.MustCompile(`{[^}]*}`)
	return true, re.ReplaceAllString(str, "%")
}

func toEntDialect(driver string) string {
	switch strings.ToLower(driver) {
	case "mysql":
		return dialect.MySQL
	case "sqlite", "sqlite3":
		return dialect.SQLite
	default:
		return dialect.Postgres
	}
}
