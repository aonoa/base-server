package biz

import (
	pb "base-server/api/gen/go/base_api/v1"
	"base-server/internal/conf"
	"context"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/util"
	adapter "github.com/casbin/ent-adapter"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"regexp"
	"strings"
)

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

var textModel string = `
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
m2 = r.obj == "/basic-api/auth/refresh" || g(r.sub, p2.sub) && g2(r.obj, p2.obj) && regexMatch(r.act, p2.act) || g(r.sub, "role:root")
m3 = g2(r.sub, p3.sub) && g3(r.obj, p3.obj) && regexMatch(r.act, p3.act) || g(r.sub, "role:root")  # api对资源组的权限（公共资源）（目前未使用）
`

// AuthUsecase is an Auth usecase.
type AuthUsecase struct {
	repo BaseRepo
	e    *casbin.Enforcer
	log  *log.Helper
}

func NewAuthUsecase(repo BaseRepo, e *casbin.Enforcer, logger log.Logger) *AuthUsecase {
	res := &AuthUsecase{
		repo: repo,
		e:    e,
		log:  log.NewHelper(logger),
	}
	res.generateAuthPolicy()
	return res
}

func KeyMatch6(key1 string, key2 string) bool {
	//fmt.Printf("key1:%s  key2:%s\n", key1, key2)

	i := strings.Index(key1, "?")

	if i != -1 {
		key1 = key1[:i]
	}

	key2 = strings.Replace(key2, "/*", "/.*", -1)

	re := regexp.MustCompile(`\{[^/]+\}`)
	key2 = re.ReplaceAllString(key2, "$1[^/]+$2")
	//fmt.Printf("res:%v key1:%s  key2:%s\n", util.RegexMatch(key1, "^"+key2+"$"), key1, key2)
	return util.RegexMatch(key1, "^"+key2+"$")
}

func NewEnforcer(confData *conf.Data) *casbin.Enforcer {
	//a, err := adapter.NewAdapter(confData.Database.Driver, confData.Database.Source)
	m, _ := model.NewModelFromString(textModel)
	a, err := adapter.NewAdapter(confData.Database.Driver, confData.Database.Source)
	//a, err := NewAdapter("postgres", "user=postgres password=postgres host=127.0.0.1 port=5432 dbname=casbin")
	if err != nil {
		panic(err)
	}
	e, _ := casbin.NewEnforcer(m, a)
	e.EnableAutoSave(true)

	e.AddNamedMatchingFunc("g2", "KeyMatch6", KeyMatch6)

	return e
}

func (uc *AuthUsecase) ReLoadPolicy() error {
	// 从存储中重新加载策略规则，用于手动修改数据库中权限表的情况
	return uc.e.LoadPolicy()
}

// HasRoleForUser 确定用户是否具有角色
func (uc *AuthUsecase) HasRoleForUser(user, role string, domain string) bool {
	ok, err := uc.e.HasRoleForUser(user, role, domain)
	if err != nil {
		return false
	}
	return ok
}

// EnforcePolicy 纯权限（未使用）
// id、资源、操作
func (uc *AuthUsecase) EnforcePolicy(rvals ...interface{}) (bool, error) {
	return uc.e.Enforce(ApiToResourceEnforceContext, rvals[0], rvals[1], rvals[2])
}

// AddUserRoles 用户添加角色
func (uc *AuthUsecase) AddUserRoles(user string, roles []string) {
	for _, role := range roles {
		_, err := uc.e.AddNamedGroupingPolicy(UserToRole, user, "role:"+role)
		if err != nil {
			uc.log.Error(err)
		}
	}
}

// DelUserRoles 用户删除角色
func (uc *AuthUsecase) DelUserRoles(user string, roles []string) {
	for _, role := range roles {
		_, err := uc.e.RemoveNamedGroupingPolicy(UserToRole, user, "role:"+role)
		if err != nil {
			uc.log.Error(err)
		}
	}
}

// DelUser 删除用户（同时删除用户关联的所有角色）
func (uc *AuthUsecase) DelUser(user string) {
	// 获取用户的所有角色
	namedGroupingPolicy, _ := uc.e.GetFilteredNamedGroupingPolicy(UserToRole, 0, user)
	for _, policy := range namedGroupingPolicy {
		_, err := uc.e.RemoveNamedGroupingPolicy(UserToRole, policy[0], policy[1])
		if err != nil {
			uc.log.Error(err)
		}
	}
	//_, err := uc.e.RemoveNamedGroupingPolicy(UserToRole, user)
	//if err != nil {
	//	println(err)
	//}
}

// UpdateRole 修改角色（继承关系的角色名称需要修改，角色关联的资源权限需要修改）
func (uc *AuthUsecase) UpdateRole(oldRole, newRole string) {
	namedGroupingPolicy, err := uc.e.GetFilteredNamedGroupingPolicy(UserToRole, 1, "role:"+oldRole)
	rules := [][]string{}
	for _, policy := range namedGroupingPolicy {
		println(policy)
		rules = append(rules, []string{
			policy[0],
			"role:" + newRole,
		})
	}
	_, err = uc.e.UpdateNamedGroupingPolicies(UserToRole, namedGroupingPolicy, rules)
	if err != nil {
		return
	}
	policyList, err := uc.e.GetFilteredNamedPolicy(PolicyUserToData, 0, "role:"+oldRole)
	if err != nil {
		return
	}
	rules = [][]string{}
	for _, policy := range policyList {
		println(policy)
		rules = append(rules, []string{
			"role:" + newRole,
			policy[1],
			policy[2],
		})
	}
	_, err = uc.e.UpdateNamedPolicies(PolicyUserToData, policyList, rules)
	if err != nil {
		return
	}
	apiPolicyList, err := uc.e.GetFilteredNamedPolicy(PolicyUserToApi, 0, "role:"+oldRole)
	if err != nil {
		return
	}
	rules = [][]string{}
	for _, policy := range apiPolicyList {
		println(policy)
		rules = append(rules, []string{
			"role:" + newRole,
			policy[1],
			policy[2],
		})
	}
	_, err = uc.e.UpdateNamedPolicies(PolicyUserToApi, apiPolicyList, rules)
	if err != nil {
		return
	}
}

// DelRoleData 删除角色上的资源
func (uc *AuthUsecase) DelRoleData(role string) {
	// 角色资源权限
	_, err := uc.e.RemoveFilteredNamedPolicy(PolicyUserToData, 0, "role:"+role)
	// 角色api权限
	_, err = uc.e.RemoveFilteredNamedPolicy(PolicyUserToApi, 0, "role:"+role)
	if err != nil {
		return
	}
}

// DelRole 删除角色
func (uc *AuthUsecase) DelRole(role string) {
	// 角色继承关系
	_, err := uc.e.RemoveFilteredNamedGroupingPolicy(UserToRole, 0, "role:"+role)
	// 用户角色关联
	_, err = uc.e.RemoveFilteredNamedGroupingPolicy(UserToRole, 1, "role:"+role)
	// 角色资源权限
	uc.DelRoleData(role)
	if err != nil {
		return
	}
}

////////////////////////////////////////////////////////
// 添加资源到资源组
// 从资源组中删除资源

func (uc *AuthUsecase) AddDataToGroup(data, dataGroup string) {
	_, err := uc.e.AddNamedGroupingPolicy(ResourceToGroup, data, "data:"+dataGroup)
	if err != nil {
		uc.log.Error(err)
	}
}
func (uc *AuthUsecase) DelDataToGroup(data, dataGroup string) {
	_, err := uc.e.RemoveNamedGroupingPolicy(ResourceToGroup, data, "data:"+dataGroup)
	if err != nil {
		uc.log.Error(err)
	}
}

func (uc *AuthUsecase) AddApiToGroup(api, apiGroup string) {
	_, err := uc.e.AddNamedGroupingPolicy(ApiToGroup, api, "api:"+apiGroup)
	if err != nil {
		uc.log.Error(err)
	}
}
func (uc *AuthUsecase) UpdateApiToGroup(oldApi, newApi []string) {
	if len(newApi) != 2 || len(oldApi) != 2 {
		uc.log.Error(errors.New(40000, "auth api update length error", "auth api update length error"))
	}
	oldApi[1] = "api:" + oldApi[1]
	newApi[1] = "api:" + newApi[1]
	_, err := uc.e.UpdateNamedGroupingPolicy(ApiToGroup, oldApi, newApi)
	if err != nil {
		uc.log.Error(err)
	}
}
func (uc *AuthUsecase) DelApiToGroup(api, apiGroup string) {
	_, err := uc.e.RemoveNamedGroupingPolicy(ApiToGroup, api, "api:"+apiGroup)
	if err != nil {
		uc.log.Error(err)
	}
}

// 为用户添加角色
// 删除用户角色

// AddPolicy 添加单条权限
func (uc *AuthUsecase) AddPolicy(role, typeStr, dataGroup, method string) {
	policyType := PolicyUserToData
	if typeStr == "api" {
		policyType = PolicyUserToApi
	}
	_, err := uc.e.AddNamedPolicy(policyType, "role:"+role, typeStr+":"+dataGroup, method)
	if err != nil {
		uc.log.Error(err)
	}
}

// UpdatePolicy 更新单条权限（含类型，旧[角色，资源，操作]，新[角色，资源，操作]）
func (uc *AuthUsecase) UpdatePolicy(typeStr string, oldPolicy, newPolicy []string) {
	if len(oldPolicy) != 3 || len(newPolicy) != 3 {
		uc.log.Error(errors.New(40000, "auth policy update length error", "auth policy update length error"))
	}
	policyType := PolicyUserToData
	if typeStr == "api" {
		policyType = PolicyUserToApi
	}
	oldPolicy[1] = "role:" + oldPolicy[1]
	oldPolicy[2] = typeStr + ":" + oldPolicy[2]
	newPolicy[1] = "role:" + newPolicy[1]
	newPolicy[2] = typeStr + ":" + newPolicy[2]
	_, err := uc.e.UpdateNamedPolicy(policyType, oldPolicy, newPolicy)
	if err != nil {
		uc.log.Error(err)
	}
}

// DelPolicy 删除单条权限
func (uc *AuthUsecase) DelPolicy(role, typeStr, dataGroup, method string) {
	policyType := PolicyUserToData
	if typeStr == "api" {
		policyType = PolicyUserToApi
	}
	_, err := uc.e.RemoveNamedPolicy(policyType, "role:"+role, typeStr+":"+dataGroup, method)
	if err != nil {
		uc.log.Error(err)
	}
}

// AddPolicies 添加多条权限
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

// UpdateDataPolicy 更新资源的关联的权限（类型，值，操作）
func (uc *AuthUsecase) UpdateDataPolicy(oldPolicy, newPolicy []string) {
	if len(oldPolicy) != 3 || len(newPolicy) != 3 {
		uc.log.Error(errors.New(40000, "auth policy update length error", "auth policy update length error"))
	}
	policyType := PolicyUserToData
	if oldPolicy[0] == "api" {
		policyType = PolicyUserToApi
	}

	policy, err := uc.e.GetFilteredNamedPolicy(policyType, 1, oldPolicy[0]+":"+oldPolicy[1], oldPolicy[2])
	if err != nil {
		return
	}

	if newPolicy[0] == "api" {
		policyType = PolicyUserToApi
	} else {
		policyType = PolicyUserToData
	}
	rules := [][]string{}
	for _, v := range policy {
		println(v)
		rules = append(rules, []string{
			v[0],
			newPolicy[0] + ":" + newPolicy[1],
			newPolicy[2],
		})
		//uc.e.AddNamedPolicy(policyType, v[0], newPolicy[0]+":"+newPolicy[1], newPolicy[2])
	}
	_, err = uc.e.AddNamedPolicies(policyType, rules)
	if err != nil {
		return
	}
}

// DelDataPolicy 删除资源关联的权限
func (uc *AuthUsecase) DelDataPolicy(typeStr, dataGroup, method string) {
	policyType := PolicyUserToData
	if typeStr == "api" {
		policyType = PolicyUserToApi
	}

	// 这里过滤直接删
	_, err := uc.e.RemoveFilteredNamedPolicy(policyType, 1, typeStr+":"+dataGroup, method)
	if err != nil {
		return
	}
}

// 更新角色的权限变化

// 生成权限
// 将api关系生成api组
// 将其他资源生成资源组
// 将用户关系生成角色
// 将角色和资源绑定成权限

// generateAuthPolicy 重新生成 casbin 权限
func (uc *AuthUsecase) generateAuthPolicy() {
	// 获取api资源数据
	ctx := context.Background()
	apiList, err := uc.repo.GetApiList(ctx, &pb.GetApiPageParams{})
	if err != nil {
		uc.log.Error(err)
	}
	for _, api := range apiList {
		uc.AddApiToGroup(api.Path, api.ResourcesGroup)
	}

	// 获取所有的资源生成资源组（暂无）

	// 为所有用户生成角色
	// 获取所有用户及其用户拥有的角色
	userList, err := uc.repo.GetUserList(ctx, 0, &pb.GetUserParams{
		Page:     0,
		PageSize: 1000000,
	})
	if err != nil {
		uc.log.Error(err)
	}
	for _, user := range userList {
		if user.Edges.Roles != nil {
			if len(user.Edges.Roles) > 0 {
				role := user.Edges.Roles[0].Value
				uc.AddUserRoles(user.ID.String(), []string{role})
			}
		} else {
			uc.AddUserRoles(user.ID.String(), []string{"default"})
		}
	}

	// 将角色和资源绑定成权限
	// 获取所有的角色以及角色绑的资源信息
	roleList, err := uc.repo.GetAllRoleList(ctx, &pb.RolePageParams{
		Page:     0,
		PageSize: 1000000,
		Status:   1,
	})
	if err != nil {
		uc.log.Error(err)
	}
	for _, role := range roleList {
		if role.Edges.Resource != nil {
			for _, resource := range role.Edges.Resource {
				uc.AddPolicy(role.Value, resource.Type, resource.Value, resource.Method)
			}
		}
	}
}

////////////////////////////////////////////////////////

// 权限：
// 1.新建（删除）用户
// 2.新建（删除）角色（增删权限）
// 3.新建（删除）部门
// 4.修改用户密码

// 0.将用户加入到其他域
// 1.新建（删除）其他域用户
// 2.新建（删除）其他域角色（增删权限）
// 3.新建（删除）其他域部门
// 4.修改其他域用户密码
// 5.赋予用户多个域的管理权限

// 1.查看菜单
// 2.查看部门
// 3.查看用户
// 4.查看角色

// 1.访问资源
// 2.添加资源
// 3.删除资源
// 4.修改资源

// 域内
// 新建的用户（默认域，所属域，默认角色，角色包含的权限）
// 1.有默认角色
// 2.有所属域
// 3.有默认域
// 4.默认角色有部门权限

// 域
// 新建域（有默认角色）
// 1.生成域操作权限
// 2.将管理员赋予域操作权限
// 删除域
// 1.删除域操作权限

// 菜单是角色，接口权限属于菜单
// 角色继承菜单角色的权限
// 用户拥有角色

/////////////////////////////////////
// 默认有一个default域 0

// 每个域都有一个default角色和一个admin角色
// default角色有基础权限
// admin角色有域的全部权限

// default域的admin角色有其他域的全部权限
// 系统有域一个初始账号属于default域的admin角色
// 其他域下的用户只能有本域的权限

///////////////////////////////////////////
// 权限包括
// 1.新建（删除）用户	（本域或其他域）		sub, dom, obj, act
// 2.新建（删除）角色	（本域或其他域）
// 3.新建（删除）部门	（本域或其他域）
// 4.修改用户密码		（本域或其他域）
// 5.赋予用户权限		（本域或其他域）

// 6.api权限			（api权限绑定于页面）（后台配置）（先不区分域）
// 7.页面可见权限		（每个菜单是一个角色，其他角色能继承菜单角色的权限）	sub, dom, obj, act

// 8.其他常规资源的访问、操作权限			sub, dom, obj, act

// 9.部分资源应该是共享的（对所有域都生效，如页面可见权限）

////////////////////////////////////////////
// 默认有一个default域	（数据库中写0）

// 每个域都有一个default角色和一个admin角色
// default角色有基础权限
// admin角色有域的全部权限

// default域的admin角色有其他域的全部权限
// 系统有域一个初始账号属于default域的admin角色
// 其他域下的用户只能有本域的权限

//// 新建一个域
// 1.新建default角色和admin角色
// 2.将部分权限赋予default角色	（可配置）
// 3.将域的全部权限赋予admin角色
// 4.default域的admin角色继承当前域的admin角色

//// 删除一个域
// 删除域相关的所有权限

//// 新建用户
// 1.给用户当前域的default角色
// 2.给用户default域的default角色

// 删除用户
// 1.删除用户与当前域的关系
// 2.删除用户在当前域的相关权限

///////////////
// 用户、角色、权限（用户拥有角色，角色绑定权限）
/////// 权限（抽象为对资源的操作，资源类型、操作方法）
// 菜单权限	可见、不可见	（菜单绑定部分api，权限组合，所以可以抽象为菜单角色）
// api权限   get、post、delete、put
// 资源权限	增删改查
/////// 用户
// admin	（只拥有超级管理员角色、包含所有权限）
// 普通用户	（初始角色，拥有少量权限）
/////// 角色（角色会出现职能重复、所以角色应该是可以组合和继承的）
// admin	（超级管理员角色、包含所有权限）
// 普通		（初始角色，拥有少量权限）
/////////
