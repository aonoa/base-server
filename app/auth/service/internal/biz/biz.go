package biz

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"time"

	v1 "base-server/api/gen/go/auth/service/v1"
	userv1 "base-server/api/gen/go/user/service/v1"
	"base-server/app/auth/service/internal/conf"
	"base-server/pkg/authx"
	"base-server/pkg/data/ent"
	"base-server/pkg/tools"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/util"
	adapter "github.com/casbin/ent-adapter"
	"github.com/go-kratos/kratos/v2/log"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/google/wire"
	"github.com/jinzhu/copier"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewEnforcer, NewAuthUsecase)

var (
	RoleToResourceEnforceContext = casbin.EnforceContext{RType: "r", PType: "p", EType: "e", MType: "m"}
	RoleToApiEnforceContext      = casbin.EnforceContext{RType: "r", PType: "p2", EType: "e", MType: "m2"}
	ApiToResourceEnforceContext  = casbin.EnforceContext{RType: "r", PType: "p3", EType: "e", MType: "m3"}
	UserToRole                   = "g"
	ApiToGroup                   = "g2"
	ResourceToGroup              = "g3"
	PolicyUserToData             = "p"
	PolicyUserToApi              = "p2"
	PolicyApiToData              = "p3"
)

var textModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act
p2 = sub, obj, act
p3 = sub, obj, act

[role_definition]
g = _, _
g2 = _, _
g3 = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && g3(r.obj, p.obj) && regexMatch(r.act, p.act) || g(r.sub, "role:root")
m2 = r.obj == "/auth-api/v1/refresh" || g(r.sub, p2.sub) && g2(r.obj, p2.obj) && regexMatch(r.act, p2.act) || g(r.sub, "role:root")
m3 = g2(r.sub, p3.sub) && g3(r.obj, p3.obj) && regexMatch(r.act, p3.act) || g(r.sub, "role:root")
`

type AuthRepo interface {
	Login(context.Context, *v1.LoginRequest) (string, error)
	GetUserAuthInfo(context.Context, string) (*userv1.GetUserAuthInfoReply, error)
	ListUserAuthBindings(context.Context) ([]*userv1.UserAuthBinding, error)
	GetCurrentUserMenuAuthority(context.Context, string) (*v1.GetCurrentUserMenuAuthorityReply, error)
	ListRoles(context.Context) ([]*ent.Role, error)
	ListAPIResources(context.Context) ([]*ent.ApiResources, error)
	ResolveRoleValues(context.Context, []int64) (map[int64]string, error)
	GetAllRoleList(context.Context, *v1.RolePageParams) ([]*ent.Role, error)
	GetRole(context.Context, int64) (*ent.Role, error)
	AddRole(context.Context, *v1.RoleListItem) (*ent.Role, error)
	UpdateRole(context.Context, int64, *v1.RoleListItem) (*ent.Role, error)
	DelRole(context.Context, int64) error
	GetApiList(context.Context, *v1.GetApiPageParams) ([]*ent.ApiResources, int64, error)
	GetApi(context.Context, string) (*ent.ApiResources, error)
	AddApi(context.Context, *ent.ApiResources) (*ent.ApiResources, error)
	UpdateApi(context.Context, *ent.ApiResources) (*ent.ApiResources, error)
	DelApi(context.Context, string) error
	GetResourceList(context.Context, *v1.GetResourcePageParams) ([]*ent.Resource, int64, error)
	AddResource(context.Context, *ent.Resource) (*ent.Resource, error)
	GetResource(context.Context, string) (*ent.Resource, error)
	UpdateResource(context.Context, *ent.Resource) (*ent.Resource, error)
	DelResource(context.Context, string) error
}

type AuthUsecase struct {
	repo AuthRepo
	e    *casbin.Enforcer
	key  string
	log  *log.Helper
}

func NewAuthUsecase(repo AuthRepo, e *casbin.Enforcer, logger log.Logger, auth *conf.Auth) *AuthUsecase {
	uc := &AuthUsecase{repo: repo, e: e, key: auth.ApiKey, log: log.NewHelper(logger)}
	if err := uc.syncAuthPolicy(); err != nil {
		uc.log.Error(err)
	}
	return uc
}

func KeyMatch6(key1 string, key2 string) bool {
	if i := strings.Index(key1, "?"); i != -1 {
		key1 = key1[:i]
	}
	key2 = strings.ReplaceAll(key2, "/*", "/.*")
	key2 = regexpPathParam.ReplaceAllString(key2, "$1[^/]+$2")
	return util.RegexMatch(key1, "^"+key2+"$")
}

var regexpPathParam = mustCompile(`\{[^/]+\}`)

func mustCompile(expr string) *regexp.Regexp {
	re, err := regexp.Compile(expr)
	if err != nil {
		panic(err)
	}
	return re
}

func NewEnforcer(confData *conf.Data) *casbin.Enforcer {
	m, _ := model.NewModelFromString(textModel)
	a, err := adapter.NewAdapter(confData.Database.Driver, confData.Database.Source)
	if err != nil {
		panic(err)
	}
	e, _ := casbin.NewEnforcer(m, a)
	e.EnableAutoSave(true)
	e.AddNamedMatchingFunc("g2", "KeyMatch6", KeyMatch6)
	return e
}

func (uc *AuthUsecase) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	uid, err := uc.repo.Login(ctx, req)
	if err != nil || uid == uuid.Nil.String() {
		return nil, err
	}
	return uc.GenerateToken(uid)
}

func (uc *AuthUsecase) GenerateToken(uid string, session ...string) (*v1.LoginReply, error) {
	now := time.Now()
	sessionID := tools.MD5(uuid.NewString())
	if len(session) > 0 && session[0] != "" {
		sessionID = session[0]
	}

	accessClaims := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, jwtv5.MapClaims{
		authx.ClaimUserID:    uid,
		"sub":               uid,
		authx.ClaimAudience:  authx.AudienceLogin,
		"exp":               now.Add(30 * time.Minute).Unix(),
		"nbf":               now.Unix(),
		"iat":               now.Unix(),
		authx.ClaimSessionID: sessionID,
	})
	accessToken, err := accessClaims.SignedString([]byte(uc.key))
	if err != nil {
		return nil, err
	}

	refreshClaims := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, jwtv5.MapClaims{
		authx.ClaimUserID:    uid,
		"sub":               uid,
		authx.ClaimAudience:  authx.AudienceRefresh,
		"exp":               now.Add(60 * time.Minute).Unix(),
		"nbf":               now.Add(25 * time.Minute).Unix(),
		"iat":               now.Unix(),
		"jti":               uuid.New().String(),
		authx.ClaimSessionID: sessionID,
	})
	refreshToken, err := refreshClaims.SignedString([]byte(uc.key))
	if err != nil {
		return nil, err
	}

	return &v1.LoginReply{
		UserId:       uid,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionId:    sessionID,
	}, nil
}

func (uc *AuthUsecase) RefreshToken(ctx context.Context, userID, aud, sessionID string) (*v1.LoginReply, error) {
	if aud != authx.AudienceRefresh {
		return &v1.LoginReply{}, nil
	}
	return uc.GenerateToken(userID, sessionID)
}

func (uc *AuthUsecase) GetAccessCodes(ctx context.Context, userID string) (*v1.GetAccessCodesReply, error) {
	user, err := uc.repo.GetUserAuthInfo(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &v1.GetAccessCodesReply{AccessCodeList: user.AccessCodes}, nil
}

func (uc *AuthUsecase) GetCurrentUserMenuAuthority(ctx context.Context, userID string) (*v1.GetCurrentUserMenuAuthorityReply, error) {
	user, err := uc.repo.GetUserAuthInfo(ctx, userID)
	if err != nil {
		return nil, err
	}
	roleItem, err := uc.repo.GetRole(ctx, user.RoleId)
	if err != nil {
		return nil, err
	}
	return &v1.GetCurrentUserMenuAuthorityReply{
		IsRoot:  roleItem.Value == "root",
		MenuIds: append([]int32(nil), roleItem.Menus...),
	}, nil
}

func (uc *AuthUsecase) CheckAuthorization(ctx context.Context, req *v1.CheckAuthorizationRequest) (*v1.CheckAuthorizationReply, error) {
	allowed, err := uc.e.Enforce(RoleToApiEnforceContext, req.UserId, req.Path, req.Method)
	if err != nil {
		return nil, err
	}
	return &v1.CheckAuthorizationReply{Allowed: allowed}, nil
}

func (uc *AuthUsecase) ReLoadPolicy(ctx context.Context) error {
	return uc.syncAuthPolicy()
}

func (uc *AuthUsecase) syncAuthPolicy() error {
	uc.e.ClearPolicy()
	uc.generateAuthPolicy()
	return uc.e.SavePolicy()
}

func (uc *AuthUsecase) AddUserRoles(user string, roles []string) {
	for _, roleValue := range roles {
		_, err := uc.e.AddNamedGroupingPolicy(UserToRole, user, "role:"+roleValue)
		if err != nil {
			uc.log.Error(err)
		}
	}
}

func (uc *AuthUsecase) AddApiToGroup(apiPath, apiGroup string) {
	_, err := uc.e.AddNamedGroupingPolicy(ApiToGroup, apiPath, "api:"+apiGroup)
	if err != nil {
		uc.log.Error(err)
	}
}

func (uc *AuthUsecase) AddPolicy(roleValue, typeStr, dataGroup, method string) {
	policyType := PolicyUserToData
	if typeStr == "api" {
		policyType = PolicyUserToApi
	}
	_, err := uc.e.AddNamedPolicy(policyType, "role:"+roleValue, typeStr+":"+dataGroup, method)
	if err != nil {
		uc.log.Error(err)
	}
}

func (uc *AuthUsecase) AddPolicies(typeStr string, rules [][]string) {
	policyType := PolicyUserToData
	if typeStr == "api" {
		policyType = PolicyUserToApi
	}
	_, err := uc.e.AddNamedPolicies(policyType, rules)
	if err != nil {
		uc.log.Error(err)
	}
}

func (uc *AuthUsecase) renameRoleBindings(oldRole, newRole string) {
	namedGroupingPolicy, err := uc.e.GetFilteredNamedGroupingPolicy(UserToRole, 1, "role:"+oldRole)
	if err == nil && len(namedGroupingPolicy) > 0 {
		rules := make([][]string, 0, len(namedGroupingPolicy))
		for _, policy := range namedGroupingPolicy {
			rules = append(rules, []string{policy[0], "role:" + newRole})
		}
		if _, err = uc.e.UpdateNamedGroupingPolicies(UserToRole, namedGroupingPolicy, rules); err != nil {
			uc.log.Error(err)
		}
	}

	for _, policyType := range []string{PolicyUserToData, PolicyUserToApi} {
		policyList, err := uc.e.GetFilteredNamedPolicy(policyType, 0, "role:"+oldRole)
		if err != nil || len(policyList) == 0 {
			continue
		}
		rules := make([][]string, 0, len(policyList))
		for _, policy := range policyList {
			rules = append(rules, []string{"role:" + newRole, policy[1], policy[2]})
		}
		if _, err = uc.e.UpdateNamedPolicies(policyType, policyList, rules); err != nil {
			uc.log.Error(err)
		}
	}
}

func (uc *AuthUsecase) removeRolePolicies(role string) {
	if _, err := uc.e.RemoveFilteredNamedPolicy(PolicyUserToData, 0, "role:"+role); err != nil {
		uc.log.Error(err)
	}
	if _, err := uc.e.RemoveFilteredNamedPolicy(PolicyUserToApi, 0, "role:"+role); err != nil {
		uc.log.Error(err)
	}
}

func (uc *AuthUsecase) removeRoleBindings(role string) {
	if _, err := uc.e.RemoveFilteredNamedGroupingPolicy(UserToRole, 0, "role:"+role); err != nil {
		uc.log.Error(err)
	}
	if _, err := uc.e.RemoveFilteredNamedGroupingPolicy(UserToRole, 1, "role:"+role); err != nil {
		uc.log.Error(err)
	}
	uc.removeRolePolicies(role)
}

func (uc *AuthUsecase) updateApiGroup(oldPath, oldGroup, newPath, newGroup string) {
	oldRule := []string{oldPath, "api:" + oldGroup}
	newRule := []string{newPath, "api:" + newGroup}
	if _, err := uc.e.UpdateNamedGroupingPolicy(ApiToGroup, oldRule, newRule); err != nil {
		uc.log.Error(err)
	}
}

func (uc *AuthUsecase) removeApiGroup(apiPath, apiGroup string) {
	if _, err := uc.e.RemoveNamedGroupingPolicy(ApiToGroup, apiPath, "api:"+apiGroup); err != nil {
		uc.log.Error(err)
	}
}

func (uc *AuthUsecase) removeDataPolicy(typeStr, dataGroup, method string) {
	policyType := PolicyUserToData
	if typeStr == "api" {
		policyType = PolicyUserToApi
	}
	if _, err := uc.e.RemoveFilteredNamedPolicy(policyType, 1, typeStr+":"+dataGroup, method); err != nil {
		uc.log.Error(err)
	}
}

func (uc *AuthUsecase) generateAuthPolicy() {
	ctx := context.Background()

	apiList, err := uc.repo.ListAPIResources(ctx)
	if err != nil {
		uc.log.Error(err)
	} else {
		for _, api := range apiList {
			uc.AddApiToGroup(api.Path, api.ResourcesGroup)
		}
	}

	roleValues := make(map[int64]string)
	userList, err := uc.repo.ListUserAuthBindings(ctx)
	if err != nil {
		uc.log.Error(err)
	} else {
		roleIDs := make([]int64, 0)
		seenRoleIDs := make(map[int64]struct{})
		for _, user := range userList {
			if user.RoleId > 0 {
				if _, ok := seenRoleIDs[user.RoleId]; !ok {
					seenRoleIDs[user.RoleId] = struct{}{}
					roleIDs = append(roleIDs, user.RoleId)
				}
			}
		}
		if len(roleIDs) > 0 {
			roleValues, err = uc.repo.ResolveRoleValues(ctx, roleIDs)
			if err != nil {
				uc.log.Error(err)
				roleValues = map[int64]string{}
			}
		}
		for _, user := range userList {
			if roleValue, ok := roleValues[user.RoleId]; ok && roleValue != "" {
				uc.AddUserRoles(user.UserId, []string{roleValue})
			} else {
				uc.AddUserRoles(user.UserId, []string{"default"})
			}
		}
	}

	roleList, err := uc.repo.ListRoles(ctx)
	if err != nil {
		uc.log.Error(err)
		return
	}
	for _, roleItem := range roleList {
		for _, resourceItem := range roleItem.Edges.Resource {
			uc.AddPolicy(roleItem.Value, resourceItem.Type, resourceItem.Value, resourceItem.Method)
		}
	}
}

func (uc *AuthUsecase) GetRoleList(ctx context.Context, req *v1.RolePageParams) (*v1.GetRoleListByPageReply, error) {
	roleList, err := uc.repo.GetAllRoleList(ctx, req)
	if err != nil {
		return nil, err
	}
	res := &v1.GetRoleListByPageReply{Items: make([]*v1.RoleListItem, 0, len(roleList)), Total: int64(len(roleList))}
	for i, item := range roleList {
		res.Items = append(res.Items, roleToReply(item, i))
	}
	return res, nil
}

func (uc *AuthUsecase) AddRole(ctx context.Context, req *v1.RoleListItem) (*v1.RoleListItem, error) {
	roleItem, err := uc.repo.AddRole(ctx, req)
	if err != nil {
		return nil, err
	}
	resources, err := roleItem.QueryResource().All(ctx)
	if err != nil {
		return nil, err
	}
	roleItem.Edges.Resource = resources

	rulesMap := map[string][][]string{}
	for _, resourceItem := range resources {
		rulesMap[resourceItem.Type] = append(rulesMap[resourceItem.Type], []string{
			"role:" + roleItem.Value,
			resourceItem.Type + ":" + resourceItem.Value,
			resourceItem.Method,
		})
	}
	for key, value := range rulesMap {
		uc.AddPolicies(key, value)
	}
	return roleToReply(roleItem, 0), nil
}

func (uc *AuthUsecase) UpdateRole(ctx context.Context, req *v1.RoleListItem) (*v1.RoleListItem, error) {
	roleID, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return nil, err
	}
	oldRole, err := uc.repo.GetRole(ctx, roleID)
	if err != nil {
		return nil, err
	}
	newRole, err := uc.repo.UpdateRole(ctx, roleID, req)
	if err != nil {
		return nil, err
	}
	resources, err := newRole.QueryResource().All(ctx)
	if err != nil {
		return nil, err
	}
	newRole.Edges.Resource = resources

	uc.removeRolePolicies(oldRole.Value)
	if oldRole.Value != newRole.Value {
		uc.renameRoleBindings(oldRole.Value, newRole.Value)
	}
	for _, resourceItem := range resources {
		uc.AddPolicy(newRole.Value, resourceItem.Type, resourceItem.Value, resourceItem.Method)
	}
	return roleToReply(newRole, 0), nil
}

func (uc *AuthUsecase) DelRole(ctx context.Context, roleID string) error {
	id, err := strconv.ParseInt(roleID, 10, 64)
	if err != nil {
		return err
	}
	roleItem, err := uc.repo.GetRole(ctx, id)
	if err != nil {
		return err
	}
	if err := uc.repo.DelRole(ctx, id); err != nil {
		return err
	}
	uc.removeRoleBindings(roleItem.Value)
	return nil
}

func (uc *AuthUsecase) GetApiList(ctx context.Context, req *v1.GetApiPageParams) (*v1.GetApiListByPageReply, error) {
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

func (uc *AuthUsecase) AddApi(ctx context.Context, req *v1.ApiListItem) (*v1.ApiListItem, error) {
	apiItem := &ent.ApiResources{}
	copier.Copy(apiItem, req)
	created, err := uc.repo.AddApi(ctx, apiItem)
	if err != nil {
		return nil, err
	}
	uc.AddApiToGroup(created.Path, created.ResourcesGroup)
	return apiToReply(created), nil
}

func (uc *AuthUsecase) UpdateApi(ctx context.Context, req *v1.ApiListItem) (*v1.ApiListItem, error) {
	oldAPI, err := uc.repo.GetApi(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	apiItem := &ent.ApiResources{}
	copier.Copy(apiItem, req)
	updated, err := uc.repo.UpdateApi(ctx, apiItem)
	if err != nil {
		return nil, err
	}
	uc.updateApiGroup(oldAPI.Path, oldAPI.ResourcesGroup, updated.Path, updated.ResourcesGroup)
	return apiToReply(updated), nil
}

func (uc *AuthUsecase) DelApi(ctx context.Context, apiID string) error {
	apiItem, err := uc.repo.GetApi(ctx, apiID)
	if err != nil {
		return err
	}
	if err := uc.repo.DelApi(ctx, apiID); err != nil {
		return err
	}
	uc.removeApiGroup(apiItem.Path, apiItem.ResourcesGroup)
	return nil
}

func (uc *AuthUsecase) GetResourceList(ctx context.Context, req *v1.GetResourcePageParams) (*v1.GetResourceListByPageReply, error) {
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

func (uc *AuthUsecase) AddResource(ctx context.Context, req *v1.ResourceListItem) (*v1.ResourceListItem, error) {
	resourceItem := &ent.Resource{}
	copier.Copy(resourceItem, req)
	created, err := uc.repo.AddResource(ctx, resourceItem)
	if err != nil {
		return nil, err
	}
	return resourceToReply(created), nil
}

func (uc *AuthUsecase) UpdateResource(ctx context.Context, req *v1.ResourceListItem) (*v1.ResourceListItem, error) {
	resourceItem := &ent.Resource{}
	copier.Copy(resourceItem, req)
	updated, err := uc.repo.UpdateResource(ctx, resourceItem)
	if err != nil {
		return nil, err
	}
	return resourceToReply(updated), nil
}

func (uc *AuthUsecase) DelResource(ctx context.Context, resourceID string) error {
	resourceItem, err := uc.repo.GetResource(ctx, resourceID)
	if err != nil {
		return err
	}
	if err := uc.repo.DelResource(ctx, resourceID); err != nil {
		return err
	}
	uc.removeDataPolicy(resourceItem.Type, resourceItem.Value, resourceItem.Method)
	return nil
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
