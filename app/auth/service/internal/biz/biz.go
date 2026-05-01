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
	BootstrapRootUserID          = "f4f9e258-fa13-4467-95fb-c86019a377f9"
)

var textModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act
p2 = sub, obj, act
p3 = sub, obj, act
# p （用户->资源）
# p2 （用户->api）
# p3 （api->资源）

[role_definition]
g = _, _
g2 = _, _
g3 = _, _
# g  (用户->角色）
# g2 (api->api_group)
# g3 (date->date_group)

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && g3(r.obj, p.obj) && regexMatch(r.act, p.act) || g(r.sub, "role:root")   # 角色对普通资源组权限
# 角色对api的权限 匹配key1:/diagnoseClass/1/diagnoseRow/aa?2  key2:/diagnoseClass/{id}/diagnoseRow/*
# 支持{id},?参数和*通配符，当api为刷新token时，直接通过
m2 = r.obj == "/auth-api/v1/refresh" || g(r.sub, p2.sub) && g2(r.obj, p2.obj) && regexMatch(r.act, p2.act) || g(r.sub, "role:root")
m3 = g2(r.sub, p3.sub) && g3(r.obj, p3.obj) && regexMatch(r.act, p3.act) || g(r.sub, "role:root") # api对资源组的权限（公共资源）（目前未使用）
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
	authDomain := authorizationDomain(req.DomainCode, req.Service)
	action := req.Method
	if req.Action != "" {
		action = req.Action
	}
	subject := scopedUserKey(req.UserId, authDomain, req.ScopeId)
	object := qualifiedAPIPath(authDomain, req.Path)
	allowed, err := uc.e.Enforce(RoleToApiEnforceContext, subject, object, action)
	if err != nil {
		return nil, err
	}
	if !allowed && (authDomain != "" || req.ScopeId != "") {
		allowed, err = uc.e.Enforce(RoleToApiEnforceContext, req.UserId, req.Path, req.Method)
		if err != nil {
			return nil, err
		}
	}
	if !allowed && req.DomainCode != "" && req.Service != "" && req.DomainCode != req.Service {
		legacySubject := scopedUserKey(req.UserId, req.Service, req.ScopeId)
		legacyObject := qualifiedAPIPath(req.Service, req.Path)
		allowed, err = uc.e.Enforce(RoleToApiEnforceContext, legacySubject, legacyObject, req.Method)
		if err != nil {
			return nil, err
		}
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
	if !snapshotUsesServiceNamespace(req) {
		uc.e.ClearPolicy()
		uc.applyPermissionSnapshot(req)
		if err := uc.e.SavePolicy(); err != nil {
			return err
		}
		uc.log.Infof("permission snapshot applied source=%s revision=%d roles=%d apis=%d bindings=%d mode=legacy-clear-all", req.SourceService, req.Revision, len(req.Roles), len(req.Apis), len(req.Bindings))
		return nil
	}
	uc.removeSourceProjection(req.SourceService)
	uc.applyPermissionSnapshot(req)
	if err := uc.e.SavePolicy(); err != nil {
		return err
	}
	uc.log.Infof("permission snapshot applied source=%s revision=%d roles=%d apis=%d bindings=%d mode=source-merge", req.SourceService, req.Revision, len(req.Roles), len(req.Apis), len(req.Bindings))
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
		uc.removeRoleBindings(projectionDomain(before.DomainCode, before.Service), before.Value)
	default:
		if after.Value == "" {
			return kratoserrors.BadRequest("BAD_REQUEST", "role delta after.value is required")
		}
		beforeDomain := projectionDomain(before.DomainCode, before.Service)
		afterDomain := projectionDomain(after.DomainCode, after.Service)
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
		uc.AddServiceAPIToGroup(projectionDomain(after.DomainCode, after.Service), after.Path, after.ResourcesGroup)
	case after == nil:
		if before.Path == "" || before.ResourcesGroup == "" {
			return kratoserrors.BadRequest("BAD_REQUEST", "api delta before.path and before.resources_group are required")
		}
		uc.removeAPIGroup(projectionDomain(before.DomainCode, before.Service), before.Path, before.ResourcesGroup)
	default:
		if before.Path == "" || before.ResourcesGroup == "" {
			return kratoserrors.BadRequest("BAD_REQUEST", "api delta before.path and before.resources_group are required")
		}
		if after.Path == "" || after.ResourcesGroup == "" {
			return kratoserrors.BadRequest("BAD_REQUEST", "api delta after.path and after.resources_group are required")
		}
		beforeDomain := projectionDomain(before.DomainCode, before.Service)
		afterDomain := projectionDomain(after.DomainCode, after.Service)
		if before.Path != after.Path || before.ResourcesGroup != after.ResourcesGroup || beforeDomain != afterDomain {
			uc.updateAPIGroup(beforeDomain, before.Path, before.ResourcesGroup, afterDomain, after.Path, after.ResourcesGroup)
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
	switch {
	case req.After != nil:
		if req.After.UserId == "" {
			return kratoserrors.BadRequest("BAD_REQUEST", "user role binding after.user_id is required")
		}
		uc.replaceUserRoleBinding(req.After.UserId, projectionDomain(req.After.DomainCode, req.After.Service), req.After.ScopeId, defaultRoleValue(req.After.RoleValue))
	case req.Before != nil:
		if req.Before.UserId == "" {
			return kratoserrors.BadRequest("BAD_REQUEST", "user role binding before.user_id is required")
		}
		uc.replaceUserRoleBinding(req.Before.UserId, projectionDomain(req.Before.DomainCode, req.Before.Service), req.Before.ScopeId, "")
	}
	uc.log.Infof("user role binding delta applied source=%s revision=%d before=%s after=%s", req.SourceService, req.Revision, bindingDeltaValue(req.Before), bindingDeltaValue(req.After))
	return nil
}

func (uc *AuthUsecase) AddUserRoles(user string, roles []string) {
	uc.AddScopedUserRoles(user, "", "", roles)
}

func (uc *AuthUsecase) AddScopedUserRoles(user, service, scope string, roles []string) {
	subject := scopedUserKey(user, service, scope)
	for _, roleValue := range roles {
		_, err := uc.e.AddNamedGroupingPolicy(UserToRole, subject, qualifiedRoleKey(service, roleValue))
		if err != nil {
			uc.log.Error(err)
		}
	}
}

func (uc *AuthUsecase) AddApiToGroup(apiPath, apiGroup string) {
	uc.AddServiceAPIToGroup("", apiPath, apiGroup)
}

func (uc *AuthUsecase) AddServiceAPIToGroup(service, apiPath, apiGroup string) {
	_, err := uc.e.AddNamedGroupingPolicy(ApiToGroup, qualifiedAPIPath(service, apiPath), qualifiedAPIGroupKey(service, apiGroup))
	if err != nil {
		uc.log.Error(err)
	}
}

func (uc *AuthUsecase) AddPolicy(roleValue, typeStr, dataGroup, method string) {
	uc.AddServicePolicy("", roleValue, typeStr, dataGroup, method)
}

func (uc *AuthUsecase) AddServicePolicy(service, roleValue, typeStr, dataGroup, method string) {
	policyType := PolicyUserToData
	if typeStr == "api" {
		policyType = PolicyUserToApi
	}
	_, err := uc.e.AddNamedPolicy(policyType, qualifiedRoleKey(service, roleValue), qualifiedDataKey(service, typeStr, dataGroup), method)
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

func (uc *AuthUsecase) renameRoleBindings(oldService, oldRole, newService, newRole string) {
	oldKey := qualifiedRoleKey(oldService, oldRole)
	newKey := qualifiedRoleKey(newService, newRole)
	namedGroupingPolicy, err := uc.e.GetFilteredNamedGroupingPolicy(UserToRole, 1, oldKey)
	if err == nil && len(namedGroupingPolicy) > 0 {
		rules := make([][]string, 0, len(namedGroupingPolicy))
		for _, policy := range namedGroupingPolicy {
			rules = append(rules, []string{policy[0], newKey})
		}
		if _, err = uc.e.UpdateNamedGroupingPolicies(UserToRole, namedGroupingPolicy, rules); err != nil {
			uc.log.Error(err)
		}
	}

	for _, policyType := range []string{PolicyUserToData, PolicyUserToApi} {
		policyList, err := uc.e.GetFilteredNamedPolicy(policyType, 0, oldKey)
		if err != nil || len(policyList) == 0 {
			continue
		}
		rules := make([][]string, 0, len(policyList))
		for _, policy := range policyList {
			rules = append(rules, []string{newKey, policy[1], policy[2]})
		}
		if _, err = uc.e.UpdateNamedPolicies(policyType, policyList, rules); err != nil {
			uc.log.Error(err)
		}
	}
}

func (uc *AuthUsecase) removeRolePolicies(service, role string) {
	roleKey := qualifiedRoleKey(service, role)
	if _, err := uc.e.RemoveFilteredNamedPolicy(PolicyUserToData, 0, roleKey); err != nil {
		uc.log.Error(err)
	}
	if _, err := uc.e.RemoveFilteredNamedPolicy(PolicyUserToApi, 0, roleKey); err != nil {
		uc.log.Error(err)
	}
}

func (uc *AuthUsecase) removeRoleBindings(service, role string) {
	roleKey := qualifiedRoleKey(service, role)
	if _, err := uc.e.RemoveFilteredNamedGroupingPolicy(UserToRole, 1, roleKey); err != nil {
		uc.log.Error(err)
	}
	uc.removeRolePolicies(service, role)
}

func (uc *AuthUsecase) updateAPIGroup(oldService, oldPath, oldGroup, newService, newPath, newGroup string) {
	oldRule := []string{qualifiedAPIPath(oldService, oldPath), qualifiedAPIGroupKey(oldService, oldGroup)}
	newRule := []string{qualifiedAPIPath(newService, newPath), qualifiedAPIGroupKey(newService, newGroup)}
	updated, err := uc.e.UpdateNamedGroupingPolicy(ApiToGroup, oldRule, newRule)
	if err != nil {
		uc.log.Error(err)
		return
	}
	if updated {
		return
	}
	uc.removeAPIGroup(oldService, oldPath, oldGroup)
	uc.AddServiceAPIToGroup(newService, newPath, newGroup)
}

func (uc *AuthUsecase) removeAPIGroup(service, apiPath, apiGroup string) {
	if _, err := uc.e.RemoveNamedGroupingPolicy(ApiToGroup, qualifiedAPIPath(service, apiPath), qualifiedAPIGroupKey(service, apiGroup)); err != nil {
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

func authorizationDomain(domainCode, service string) string {
	return projectionDomain(domainCode, service)
}

func projectionDomain(domainCode, service string) string {
	domainCode = strings.TrimSpace(domainCode)
	if domainCode != "" {
		return domainCode
	}
	return strings.TrimSpace(service)
}

func roleDeltaValue(item *v1.PolicyRole) string {
	if item == nil {
		return "nil"
	}
	service := projectionDomain(item.DomainCode, item.Service)
	if service == "" {
		return item.Value
	}
	return service + ":" + item.Value
}

func apiDeltaValue(item *v1.PolicyApi) string {
	if item == nil {
		return "nil"
	}
	service := projectionDomain(item.DomainCode, item.Service)
	if service == "" {
		return item.Path + "->" + item.ResourcesGroup
	}
	return service + ":" + item.Path + "->" + item.ResourcesGroup
}

func bindingDeltaValue(item *v1.PolicyUserRoleBinding) string {
	if item == nil {
		return "nil"
	}
	return scopedUserKey(item.UserId, projectionDomain(item.DomainCode, item.Service), item.ScopeId) + "->" + item.RoleValue
}

func (uc *AuthUsecase) replaceUserRoleBinding(userID, service, scope, roleValue string) {
	subject := scopedUserKey(userID, service, scope)
	if _, err := uc.e.RemoveFilteredNamedGroupingPolicy(UserToRole, 0, subject); err != nil {
		uc.log.Error(err)
	}
	if roleValue != "" {
		uc.AddScopedUserRoles(userID, service, scope, []string{roleValue})
	}
}

func (uc *AuthUsecase) syncRolePolicies(role *v1.PolicyRole) {
	if role == nil || role.Value == "" {
		return
	}
	service := projectionDomain(role.DomainCode, role.Service)
	uc.removeRolePolicies(service, role.Value)
	if !role.Status {
		return
	}
	for _, resource := range role.Resources {
		if resource == nil || resource.Type == "" || resource.Value == "" {
			continue
		}
		uc.AddServicePolicy(service, role.Value, resource.Type, resource.Value, resource.Method)
	}
}

func (uc *AuthUsecase) applyPermissionSnapshot(req *v1.RegisterPermissionSnapshotRequest) {
	for _, api := range req.Apis {
		if api == nil || api.Path == "" || api.ResourcesGroup == "" {
			continue
		}
		uc.AddServiceAPIToGroup(projectionDomain(api.DomainCode, api.Service), api.Path, api.ResourcesGroup)
	}
	for _, binding := range req.Bindings {
		if binding == nil || binding.UserId == "" {
			continue
		}
		uc.AddScopedUserRoles(binding.UserId, projectionDomain(binding.DomainCode, binding.Service), binding.ScopeId, []string{defaultRoleValue(binding.RoleValue)})
	}
	uc.AddUserRoles(BootstrapRootUserID, []string{"root"})
	for _, role := range req.Roles {
		uc.syncRolePolicies(role)
	}
}

func snapshotUsesServiceNamespace(req *v1.RegisterPermissionSnapshotRequest) bool {
	if req == nil {
		return false
	}
	for _, role := range req.Roles {
		if role != nil && (strings.TrimSpace(role.Service) != "" || strings.TrimSpace(role.DomainCode) != "") {
			return true
		}
	}
	for _, api := range req.Apis {
		if api != nil && (strings.TrimSpace(api.Service) != "" || strings.TrimSpace(api.DomainCode) != "") {
			return true
		}
	}
	for _, binding := range req.Bindings {
		if binding != nil && (strings.TrimSpace(binding.Service) != "" || strings.TrimSpace(binding.DomainCode) != "" || strings.TrimSpace(binding.ScopeId) != "") {
			return true
		}
	}
	return false
}

func (uc *AuthUsecase) removeSourceProjection(source string) {
	source = strings.TrimSpace(source)
	if source == "" {
		return
	}
	rolePrefix := "role:" + source + ":"
	subjectPrefix := "subject:" + source + ":"
	apiPathPrefix := source + ":"
	apiGroupPrefix := "api:" + source + ":"

	uc.removeNamedPolicyByPrefix(PolicyUserToData, 0, rolePrefix)
	uc.removeNamedPolicyByPrefix(PolicyUserToApi, 0, rolePrefix)
	uc.removeNamedGroupingPolicyByPrefix(UserToRole, 0, subjectPrefix)
	uc.removeNamedGroupingPolicyByPrefix(UserToRole, 1, rolePrefix)
	uc.removeNamedGroupingPolicyByPrefix(ApiToGroup, 0, apiPathPrefix)
	uc.removeNamedGroupingPolicyByPrefix(ApiToGroup, 1, apiGroupPrefix)
}

func (uc *AuthUsecase) removeNamedPolicyByPrefix(policyType string, fieldIndex int, prefix string) {
	policies, err := uc.e.GetNamedPolicy(policyType)
	if err != nil {
		uc.log.Error(err)
		return
	}
	for _, policy := range policies {
		if fieldIndex >= len(policy) {
			continue
		}
		if strings.HasPrefix(policy[fieldIndex], prefix) {
			if _, err := uc.e.RemoveNamedPolicy(policyType, stringSliceToInterfaces(policy)...); err != nil {
				uc.log.Error(err)
			}
		}
	}
}

func (uc *AuthUsecase) removeNamedGroupingPolicyByPrefix(groupingType string, fieldIndex int, prefix string) {
	policies, err := uc.e.GetNamedGroupingPolicy(groupingType)
	if err != nil {
		uc.log.Error(err)
		return
	}
	for _, policy := range policies {
		if fieldIndex >= len(policy) {
			continue
		}
		if strings.HasPrefix(policy[fieldIndex], prefix) {
			if _, err := uc.e.RemoveNamedGroupingPolicy(groupingType, stringSliceToInterfaces(policy)...); err != nil {
				uc.log.Error(err)
			}
		}
	}
}

func stringSliceToInterfaces(items []string) []interface{} {
	out := make([]interface{}, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}

func scopedUserKey(userID, service, scope string) string {
	service = strings.TrimSpace(service)
	scope = strings.TrimSpace(scope)
	switch {
	case service == "" && scope == "":
		return userID
	case scope == "":
		return "subject:" + service + ":" + userID
	default:
		return "subject:" + service + ":" + scope + ":" + userID
	}
}

func qualifiedRoleKey(service, roleValue string) string {
	service = strings.TrimSpace(service)
	if service == "" {
		return "role:" + roleValue
	}
	return "role:" + service + ":" + roleValue
}

func qualifiedAPIGroupKey(service, group string) string {
	service = strings.TrimSpace(service)
	if service == "" {
		return "api:" + group
	}
	return "api:" + service + ":" + group
}

func qualifiedDataKey(service, typeStr, value string) string {
	service = strings.TrimSpace(service)
	if service == "" {
		return typeStr + ":" + value
	}
	return typeStr + ":" + service + ":" + value
}

func qualifiedAPIPath(service, path string) string {
	service = strings.TrimSpace(service)
	if service == "" {
		return path
	}
	return service + ":" + path
}
