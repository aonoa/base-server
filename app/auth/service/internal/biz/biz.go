package biz

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"

	v1 "base-server/api/gen/go/auth/service/v1"
	userv1 "base-server/api/gen/go/user/service/v1"
	"base-server/app/auth/service/internal/conf"
	"base-server/pkg/authx"
	"base-server/pkg/tools"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/util"
	adapter "github.com/casbin/ent-adapter"
	kratoserrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/google/wire"
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
	DefaultOrganizationID        = "9f740c1b-0210-4e3a-858d-d128edea924d"
	GlobalRootDomain             = "global"
)

var textModel = `
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act
p2 = sub, dom, obj, act
p3 = sub, dom, obj, act
# p （用户->资源）
# p2 （用户->api）
# p3 （api->资源）

[role_definition]
g = _, _, _
g2 = _, _
g3 = _, _
# g  (用户->角色）
# g2 (api->api_group)
# g3 (date->date_group)

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = (g(r.sub, p.sub, r.dom) && g3(r.obj, p.obj) && regexMatch(r.act, p.act)) || g(r.sub, "root", "global")   # 角色对普通资源组权限
# 角色对api的权限 匹配key1:/diagnoseClass/1/diagnoseRow/aa?2  key2:/diagnoseClass/{id}/diagnoseRow/*
# 支持{id},?参数和*通配符，当api为刷新token时，直接通过
m2 = r.obj == "/auth-api/v1/refresh" || (g(r.sub, p2.sub, r.dom) && g2(r.obj, p2.obj) && regexMatch(r.act, p2.act)) || g(r.sub, "root", "global")
m3 = (g2(r.sub, p3.sub) && r.dom == p3.dom && g3(r.obj, p3.obj) && regexMatch(r.act, p3.act)) || g(r.sub, "root", "global") # api对资源组的权限（公共资源）（目前未使用）
`

type AuthRepo interface {
	Login(context.Context, *v1.LoginRequest) (string, error)
	GetUserAuthInfo(context.Context, string) (*userv1.GetUserAuthInfoReply, error)
}

type AuthUsecase struct {
	repo         AuthRepo
	e            *casbin.Enforcer
	key          string
	log          *log.Helper
	projectionMu sync.Mutex
}

func NewAuthUsecase(repo AuthRepo, e *casbin.Enforcer, logger log.Logger, auth *conf.Auth) *AuthUsecase {
	uc := &AuthUsecase{repo: repo, e: e, key: auth.ApiKey, log: log.NewHelper(logger)}
	if err := uc.e.LoadPolicy(); err != nil {
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
		"sub":                uid,
		authx.ClaimAudience:  authx.AudienceLogin,
		"exp":                now.Add(3 * time.Minute).Unix(),
		"nbf":                now.Unix(),
		"iat":                now.Unix(),
		authx.ClaimSessionID: sessionID,
	})
	accessToken, err := accessClaims.SignedString([]byte(uc.key))
	if err != nil {
		return nil, err
	}

	refreshClaims := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, jwtv5.MapClaims{
		authx.ClaimUserID:    uid,
		"sub":                uid,
		authx.ClaimAudience:  authx.AudienceRefresh,
		"exp":                now.Add(60 * time.Minute).Unix(),
		"nbf":                now.Add(2 * time.Minute).Unix(),
		"iat":                now.Unix(),
		"jti":                uuid.New().String(),
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
		return nil, kratoserrors.Unauthorized("UNAUTHORIZED", "invalid refresh token")
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

func (uc *AuthUsecase) CheckAuthorization(ctx context.Context, req *v1.CheckAuthorizationRequest) (*v1.CheckAuthorizationReply, error) {
	action := req.Method
	if req.Action != "" {
		action = req.Action
	}
	domain := organizationDomain(req.OrganizationId)
	allowed, err := uc.enforceWithDefaultRoleFallback(req.UserId, domain, req.Path, action)
	if err != nil {
		return nil, err
	}
	return &v1.CheckAuthorizationReply{Allowed: allowed}, nil
}

func (uc *AuthUsecase) RegisterPermissionSnapshot(ctx context.Context, req *v1.RegisterPermissionSnapshotRequest) error {
	_ = ctx
	if req == nil {
		return kratoserrors.BadRequest("BAD_REQUEST", "permission snapshot is required")
	}
	uc.projectionMu.Lock()
	defer uc.projectionMu.Unlock()
	uc.e.ClearPolicy()
	uc.applyPermissionSnapshot(req)
	if err := uc.e.SavePolicy(); err != nil {
		return err
	}
	uc.log.Infof("permission snapshot applied source=%s revision=%d roles=%d apis=%d bindings=%d mode=domain-clear-all", req.SourceService, req.Revision, len(req.Roles), len(req.Apis), len(req.Bindings))
	return nil
}

func (uc *AuthUsecase) ApplyRoleDelta(ctx context.Context, req *v1.ApplyRoleDeltaRequest) error {
	_ = ctx
	if req == nil || (req.Before == nil && req.After == nil) {
		return kratoserrors.BadRequest("BAD_REQUEST", "role delta requires before or after")
	}
	uc.projectionMu.Lock()
	defer uc.projectionMu.Unlock()
	before, after := req.Before, req.After
	switch {
	case before == nil:
		if after.Value == "" {
			return kratoserrors.BadRequest("BAD_REQUEST", "role delta after.value is required")
		}
		uc.syncRolePolicies(after)
	case after == nil:
		if before.Value == "" {
			return kratoserrors.BadRequest("BAD_REQUEST", "role delta before.value is required")
		}
		uc.removeRoleBindings(roleDomain(before), before.Value)
	default:
		if after.Value == "" {
			return kratoserrors.BadRequest("BAD_REQUEST", "role delta after.value is required")
		}
		beforeDomain := roleDomain(before)
		afterDomain := roleDomain(after)
		if before.Value != "" && (before.Value != after.Value || beforeDomain != afterDomain) {
			uc.renameRoleBindings(beforeDomain, before.Value, afterDomain, after.Value)
		}
		uc.syncRolePolicies(after)
	}
	uc.log.Infof("role delta applied source=%s revision=%d before=%s after=%s", req.SourceService, req.Revision, roleDeltaValue(before), roleDeltaValue(after))
	return nil
}

func (uc *AuthUsecase) ApplyApiDelta(ctx context.Context, req *v1.ApplyApiDeltaRequest) error {
	_ = ctx
	if req == nil || (req.Before == nil && req.After == nil) {
		return kratoserrors.BadRequest("BAD_REQUEST", "api delta requires before or after")
	}
	uc.projectionMu.Lock()
	defer uc.projectionMu.Unlock()
	before, after := req.Before, req.After
	switch {
	case before == nil:
		if after.Path == "" || after.ResourcesGroup == "" {
			return kratoserrors.BadRequest("BAD_REQUEST", "api delta after.path and after.resources_group are required")
		}
		uc.AddAPIToGroup(after.Path, after.ResourcesGroup)
	case after == nil:
		if before.Path == "" || before.ResourcesGroup == "" {
			return kratoserrors.BadRequest("BAD_REQUEST", "api delta before.path and before.resources_group are required")
		}
		uc.removeAPIGroup(before.Path, before.ResourcesGroup)
	default:
		if before.Path == "" || before.ResourcesGroup == "" {
			return kratoserrors.BadRequest("BAD_REQUEST", "api delta before.path and before.resources_group are required")
		}
		if after.Path == "" || after.ResourcesGroup == "" {
			return kratoserrors.BadRequest("BAD_REQUEST", "api delta after.path and after.resources_group are required")
		}
		if before.Path != after.Path || before.ResourcesGroup != after.ResourcesGroup {
			uc.updateAPIGroup(before.Path, before.ResourcesGroup, after.Path, after.ResourcesGroup)
		}
	}
	uc.log.Infof("api delta applied source=%s revision=%d before=%s after=%s", req.SourceService, req.Revision, apiDeltaValue(before), apiDeltaValue(after))
	return nil
}

func (uc *AuthUsecase) ApplyUserRoleBindingDelta(ctx context.Context, req *v1.ApplyUserRoleBindingDeltaRequest) error {
	_ = ctx
	if req == nil || (req.Before == nil && req.After == nil) {
		return kratoserrors.BadRequest("BAD_REQUEST", "user role binding delta requires before or after")
	}
	uc.projectionMu.Lock()
	defer uc.projectionMu.Unlock()
	if req.Before != nil {
		if req.Before.UserId == "" {
			return kratoserrors.BadRequest("BAD_REQUEST", "user role binding before.user_id is required")
		}
		uc.removeUserRoleBinding(req.Before.UserId, effectiveBindingDomain(req.Before), defaultRoleValue(req.Before.RoleValue))
	}
	if req.After != nil {
		if req.After.UserId == "" {
			return kratoserrors.BadRequest("BAD_REQUEST", "user role binding after.user_id is required")
		}
		uc.addUserRoleBinding(req.After.UserId, effectiveBindingDomain(req.After), defaultRoleValue(req.After.RoleValue))
	}
	uc.log.Infof("user role binding delta applied source=%s revision=%d before=%s after=%s", req.SourceService, req.Revision, bindingDeltaValue(req.Before), bindingDeltaValue(req.After))
	return nil
}

func (uc *AuthUsecase) AddUserRoles(user string, roles []string) {
	uc.AddDomainUserRoles(user, DefaultOrganizationID, roles)
}

func (uc *AuthUsecase) AddDomainUserRoles(user, domain string, roles []string) {
	domain = organizationDomain(domain)
	for _, roleValue := range roles {
		_, err := uc.e.AddNamedGroupingPolicy(UserToRole, user, roleValue, domain)
		if err != nil {
			uc.log.Error(err)
		}
	}
}

func (uc *AuthUsecase) AddApiToGroup(apiPath, apiGroup string) {
	uc.AddAPIToGroup(apiPath, apiGroup)
}

func (uc *AuthUsecase) AddAPIToGroup(apiPath, apiGroup string) {
	_, err := uc.e.AddNamedGroupingPolicy(ApiToGroup, apiPath, apiGroup)
	if err != nil {
		uc.log.Error(err)
	}
}

func (uc *AuthUsecase) AddPolicy(roleValue, typeStr, dataGroup, method string) {
	uc.AddDomainPolicy(DefaultOrganizationID, roleValue, typeStr, dataGroup, method)
}

func (uc *AuthUsecase) AddDomainPolicy(domain, roleValue, typeStr, dataGroup, method string) {
	policyType := PolicyUserToData
	if typeStr == "api" {
		policyType = PolicyUserToApi
	}
	_, err := uc.e.AddNamedPolicy(policyType, roleValue, organizationDomain(domain), dataGroup, method)
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

func (uc *AuthUsecase) enforceWithDefaultRoleFallback(subject, domain, object, action string) (bool, error) {
	domain = organizationDomain(domain)
	allowed, err := uc.e.Enforce(RoleToApiEnforceContext, subject, domain, object, action)
	if err != nil || allowed {
		return allowed, err
	}
	hasBinding, err := uc.hasExplicitRoleBinding(subject, domain)
	if err != nil || hasBinding {
		return allowed, err
	}
	return uc.isDefaultRoleAllowed(domain, object, action)
}

func (uc *AuthUsecase) hasExplicitRoleBinding(subject, domain string) (bool, error) {
	policies, err := uc.e.GetFilteredNamedGroupingPolicy(UserToRole, 0, subject)
	if err != nil {
		return false, err
	}
	for _, policy := range policies {
		if len(policy) >= 3 && policy[2] == organizationDomain(domain) {
			return true, nil
		}
	}
	return false, nil
}

func (uc *AuthUsecase) isDefaultRoleAllowed(domain, object, action string) (bool, error) {
	apiGroups, err := uc.matchingAPIGroups(object)
	if err != nil {
		return false, err
	}
	for _, apiGroup := range apiGroups {
		if len(apiGroup) < 2 {
			continue
		}
		resourceGroup := apiGroup[1]
		policies, err := uc.e.GetFilteredNamedPolicy(PolicyUserToApi, 0, "default", organizationDomain(domain), resourceGroup)
		if err != nil {
			return false, err
		}
		for _, policy := range policies {
			if len(policy) < 4 {
				continue
			}
			if util.RegexMatch(action, policy[3]) {
				return true, nil
			}
		}
	}
	return false, nil
}

func (uc *AuthUsecase) matchingAPIGroups(object string) ([][]string, error) {
	apiGroups, err := uc.e.GetNamedGroupingPolicy(ApiToGroup)
	if err != nil {
		return nil, err
	}
	items := make([][]string, 0)
	for _, apiGroup := range apiGroups {
		if len(apiGroup) < 2 {
			continue
		}
		if apiGroup[0] == object || KeyMatch6(object, apiGroup[0]) {
			items = append(items, apiGroup)
		}
	}
	return items, nil
}

func (uc *AuthUsecase) renameRoleBindings(oldDomain, oldRole, newDomain, newRole string) {
	oldDomain = organizationDomain(oldDomain)
	newDomain = organizationDomain(newDomain)
	namedGroupingPolicy, err := uc.e.GetFilteredNamedGroupingPolicy(UserToRole, 1, oldRole, oldDomain)
	if err == nil && len(namedGroupingPolicy) > 0 {
		rules := make([][]string, 0, len(namedGroupingPolicy))
		for _, policy := range namedGroupingPolicy {
			rules = append(rules, []string{policy[0], newRole, newDomain})
		}
		if _, err = uc.e.UpdateNamedGroupingPolicies(UserToRole, namedGroupingPolicy, rules); err != nil {
			uc.log.Error(err)
		}
	}

	for _, policyType := range []string{PolicyUserToData, PolicyUserToApi} {
		policyList, err := uc.e.GetFilteredNamedPolicy(policyType, 0, oldRole, oldDomain)
		if err != nil || len(policyList) == 0 {
			continue
		}
		rules := make([][]string, 0, len(policyList))
		for _, policy := range policyList {
			rules = append(rules, []string{newRole, newDomain, policy[2], policy[3]})
		}
		if _, err = uc.e.UpdateNamedPolicies(policyType, policyList, rules); err != nil {
			uc.log.Error(err)
		}
	}
}

func (uc *AuthUsecase) removeRolePolicies(domain, role string) {
	domain = organizationDomain(domain)
	if _, err := uc.e.RemoveFilteredNamedPolicy(PolicyUserToData, 0, role, domain); err != nil {
		uc.log.Error(err)
	}
	if _, err := uc.e.RemoveFilteredNamedPolicy(PolicyUserToApi, 0, role, domain); err != nil {
		uc.log.Error(err)
	}
}

func (uc *AuthUsecase) removeRoleBindings(domain, role string) {
	domain = organizationDomain(domain)
	if _, err := uc.e.RemoveFilteredNamedGroupingPolicy(UserToRole, 1, role, domain); err != nil {
		uc.log.Error(err)
	}
	uc.removeRolePolicies(domain, role)
}

func (uc *AuthUsecase) updateAPIGroup(oldPath, oldGroup, newPath, newGroup string) {
	oldRule := []string{oldPath, oldGroup}
	newRule := []string{newPath, newGroup}
	updated, err := uc.e.UpdateNamedGroupingPolicy(ApiToGroup, oldRule, newRule)
	if err != nil {
		uc.log.Error(err)
		return
	}
	if updated {
		return
	}
	uc.removeAPIGroup(oldPath, oldGroup)
	uc.AddAPIToGroup(newPath, newGroup)
}

func (uc *AuthUsecase) removeAPIGroup(apiPath, apiGroup string) {
	if _, err := uc.e.RemoveNamedGroupingPolicy(ApiToGroup, apiPath, apiGroup); err != nil {
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

func defaultRoleValue(roleValue string) string {
	if roleValue == "" {
		return "default"
	}
	return roleValue
}

func organizationDomain(scope string) string {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		return DefaultOrganizationID
	}
	return scope
}

func roleDomain(role *v1.PolicyRole) string {
	if role == nil {
		return DefaultOrganizationID
	}
	return organizationDomain(role.OrganizationId)
}

func roleDeltaValue(item *v1.PolicyRole) string {
	if item == nil {
		return "nil"
	}
	return roleDomain(item) + ":" + item.Value
}

func apiDeltaValue(item *v1.PolicyApi) string {
	if item == nil {
		return "nil"
	}
	return item.Path + "->" + item.ResourcesGroup
}

func bindingDeltaValue(item *v1.PolicyUserRoleBinding) string {
	if item == nil {
		return "nil"
	}
	return item.UserId + ":" + effectiveBindingDomain(item) + "->" + item.RoleValue
}

func effectiveBindingDomain(binding *v1.PolicyUserRoleBinding) string {
	if binding == nil {
		return DefaultOrganizationID
	}
	return roleBindingDomain(binding.OrganizationId, defaultRoleValue(binding.RoleValue))
}

func roleBindingDomain(scope, roleValue string) string {
	domain := organizationDomain(scope)
	if roleValue == "root" {
		return GlobalRootDomain
	}
	return domain
}

func (uc *AuthUsecase) removeUserRoleBinding(userID, domain, roleValue string) {
	domain = roleBindingDomain(domain, roleValue)
	if roleValue == "" {
		if _, err := uc.e.RemoveFilteredNamedGroupingPolicy(UserToRole, 0, userID, domain); err != nil {
			uc.log.Error(err)
		}
		return
	}
	if _, err := uc.e.RemoveNamedGroupingPolicy(UserToRole, userID, roleValue, domain); err != nil {
		uc.log.Error(err)
	}
}

func (uc *AuthUsecase) addUserRoleBinding(userID, domain, roleValue string) {
	domain = roleBindingDomain(domain, roleValue)
	if roleValue == "" {
		return
	}
	uc.AddDomainUserRoles(userID, domain, []string{roleValue})
}

func (uc *AuthUsecase) syncRolePolicies(role *v1.PolicyRole) {
	if role == nil || role.Value == "" {
		return
	}
	domain := roleDomain(role)
	uc.removeRolePolicies(domain, role.Value)
	if !role.Status {
		return
	}
	for _, resource := range role.Resources {
		if resource == nil || resource.Type == "" || resource.Value == "" {
			continue
		}
		uc.AddDomainPolicy(domain, role.Value, resource.Type, resource.Value, resource.Method)
	}
}

func (uc *AuthUsecase) applyPermissionSnapshot(req *v1.RegisterPermissionSnapshotRequest) {
	for _, api := range req.Apis {
		if api == nil || api.Path == "" || api.ResourcesGroup == "" {
			continue
		}
		uc.AddAPIToGroup(api.Path, api.ResourcesGroup)
	}
	for _, binding := range req.Bindings {
		if binding == nil || binding.UserId == "" {
			continue
		}
		roleValue := defaultRoleValue(binding.RoleValue)
		domain := roleBindingDomain(binding.OrganizationId, roleValue)
		uc.AddDomainUserRoles(binding.UserId, domain, []string{roleValue})
	}
	for _, role := range req.Roles {
		uc.syncRolePolicies(role)
	}
}
