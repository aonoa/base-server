package biz

import (
	pb "base-server/api/gen/go/base_api/v1"
	"base-server/internal/conf"
	"context"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/util"
	adapter "github.com/casbin/ent-adapter"
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
m = g(r.sub, p.sub) && g3(r.obj, p.obj) && regexMatch(r.act, p.act) || r.sub == "root"   # 角色对普通资源组权限
# 角色对api的权限 匹配key1:/diagnoseClass/1/diagnoseRow/aa?2  key2:/diagnoseClass/{id}/diagnoseRow/*
# 支持{id},?参数和*通配符，当api为刷新token时，直接通过
m2 = r.obj == "/basic-api/auth/refresh" || g(r.sub, p2.sub) && g2(r.obj, p2.obj) && regexMatch(r.act, p2.act) || r.sub == "root"
m3 = g2(r.sub, p3.sub) && g3(r.obj, p3.obj) && regexMatch(r.act, p3.act) || r.sub == "root"  # api对资源组的权限（公共资源）
`

// AuthUsecase is an Auth usecase.
type AuthUsecase struct {
	repo BaseRepo
	e    *casbin.Enforcer
	log  *log.Helper
}

func NewAuthUsecase(repo BaseRepo, e *casbin.Enforcer, logger log.Logger) *AuthUsecase {
	return &AuthUsecase{
		repo: repo,
		e:    e,
		log:  log.NewHelper(logger),
	}
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
	//a.AddPolicy("", "p", []string{"d1f7b7c1-c0b6-4707-aa17-5055b09b3ae8", "/basic-api/getUserInfo", "GET"})
	//a.AddPolicy("", "p", []string{"d1f7b7c1-c0b6-4707-aa17-5055b09b3ae8", "/basic-api/getMenuList", "GET"})
	//a.AddPolicy("", "p", []string{"d1f7b7c1-c0b6-4707-aa17-5055b09b3ae8", "/basic-api/getPermCode", "GET"})
	e, _ := casbin.NewEnforcer(m, a)
	//
	e.EnableAutoSave(true)
	//// 从存储中重新加载策略规则
	//err := e.LoadPolicy()
	//if err != nil {
	//	return nil
	//}
	//e.AddNamedMatchingFunc("g", "KeyMatch6", KeyMatch6)

	////e.AddPolicies([][]string{
	////	{"user_group", "/basic-api/getUserInfo", "GET"},
	////	{"user_group", "/basic-api/getMenuList", "GET"},
	////	{"user_group", "/basic-api/getPermCode", "GET"},
	////})
	////
	////e.AddNamedPolicy("p2", "user_group", "data_group", "(read)|(write)")
	////e.AddNamedGroupingPolicy("g", "d1f7b7c1-c0b6-4707-aa17-5055b09b3ae8", "user_group")
	////
	////e.AddNamedGroupingPolicies("g2", [][]string{
	////	{"data1", "data_group"},
	////	{"data2", "data_group"},
	////})

	e.AddNamedMatchingFunc("g2", "KeyMatch6", KeyMatch6)
	//e.AddNamedMatchingFunc("g2", "KeyMatch5", util.KeyMatch5)
	//e.AddNamedDomainMatchingFunc("g2", "KeyMatch6", KeyMatch6)

	// 测试
	//// ok, err := e.Enforce("d1f7b7c1-c0b6-4707-aa17-5055b09b3ae8", "/basic-api/getUserInfo/1", "GET")
	//enforceContext := casbin.EnforceContext{RType: "r", PType: "p2", EType: "e", MType: "m2"}
	////ok, err := e.Enforce(enforceContext, "d1f7b7c1-c0b6-4707-aa17-5055b09b3ae8", "/diagnoseClass/1/diagnoseRow/aa?2", "GET")
	//ok, err := e.Enforce(enforceContext, "a0bb672a-a4b1-4ec9-807a-ba11e000d2a4", "/basic-api/user/info", "GET")
	//fmt.Println(ok)
	//if err != nil {
	//	fmt.Println(err)
	//}
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

// EnforcePolicy 纯权限
// id、资源、操作
func (uc *AuthUsecase) EnforcePolicy(rvals ...interface{}) (bool, error) {
	enforceContext := casbin.EnforceContext{RType: "r", PType: "p3", EType: "e", MType: "m2"}
	return uc.e.Enforce(enforceContext, rvals[0], rvals[1], rvals[2])
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

// DelUser 删除用户
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
func (uc *AuthUsecase) DelApiToGroup(api, apiGroup string) {
	_, err := uc.e.RemoveNamedGroupingPolicy(ApiToGroup, api, "api:"+apiGroup)
	if err != nil {
		uc.log.Error(err)
	}
}

// 为用户添加角色
// 删除用户角色

// 添加角色对api资源的权限
// 删除角色对api资源的权限

func (uc *AuthUsecase) AddApiPolicy(role, apiGroup, method string) {
	_, err := uc.e.AddNamedPolicy(PolicyUserToApi, "role:"+role, "api:"+apiGroup, method)
	if err != nil {
		uc.log.Error(err)
	}
}

func (uc *AuthUsecase) DelApiPolicy(role, apiGroup, method string) {
	_, err := uc.e.RemoveNamedPolicy(PolicyUserToApi, "role:"+role, "api:"+apiGroup, method)
	if err != nil {
		uc.log.Error(err)
	}
}

func (uc *AuthUsecase) AddDataPolicy(role, dataGroup, method string) {
	_, err := uc.e.AddNamedPolicy(PolicyUserToData, "role:"+role, "data:"+dataGroup, method)
	if err != nil {
		uc.log.Error(err)
	}
}

func (uc *AuthUsecase) DelDataPolicy(role, dataGroup, method string) {
	_, err := uc.e.RemoveNamedPolicy(PolicyUserToData, "role:"+role, "data:"+dataGroup, method)
	if err != nil {
		uc.log.Error(err)
	}
}

// 生成权限
// 将api关系生成api组
// 将其他资源生成资源组
// 将用户关系生成角色
// 将角色和资源绑定成权限
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
