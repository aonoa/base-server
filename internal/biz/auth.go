package biz

import (
	"base-server/internal/conf"
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/util"
	adapter "github.com/casbin/ent-adapter"
	"github.com/go-kratos/kratos/v2/log"
	"regexp"
	"strings"
)

// AuthUsecase is an Auth usecase.
type AuthUsecase struct {
	e   *casbin.Enforcer
	log *log.Helper
}

func NewAuthUsecase(confAuth *conf.Auth, confData *conf.Data, logger log.Logger) *AuthUsecase {
	return &AuthUsecase{
		e:   NewEnforcer(confAuth, confData),
		log: log.NewHelper(logger),
	}
}

func KeyMatch6(key1 string, key2 string) bool {
	fmt.Printf("key1:%s  key2:%s\n", key1, key2)

	i := strings.Index(key1, "?")

	if i != -1 {
		key1 = key1[:i]
	}

	key2 = strings.Replace(key2, "/*", "/.*", -1)

	re := regexp.MustCompile(`\{[^/]+\}`)
	key2 = re.ReplaceAllString(key2, "$1[^/]+$2")
	fmt.Printf("res:%v key1:%s  key2:%s\n", util.RegexMatch(key1, "^"+key2+"$"), key1, key2)
	return util.RegexMatch(key1, "^"+key2+"$")
}

func NewEnforcer(confAuth *conf.Auth, confData *conf.Data) *casbin.Enforcer {
	a, err := adapter.NewAdapter(confData.Database.Driver, confData.Database.Source)
	//a, err := NewAdapter("postgres", "user=postgres password=postgres host=127.0.0.1 port=5432 dbname=casbin")
	if err != nil {
		panic(err)
	}
	//a.AddPolicy("", "p", []string{"d1f7b7c1-c0b6-4707-aa17-5055b09b3ae8", "/basic-api/getUserInfo", "GET"})
	//a.AddPolicy("", "p", []string{"d1f7b7c1-c0b6-4707-aa17-5055b09b3ae8", "/basic-api/getMenuList", "GET"})
	//a.AddPolicy("", "p", []string{"d1f7b7c1-c0b6-4707-aa17-5055b09b3ae8", "/basic-api/getPermCode", "GET"})
	e, _ := casbin.NewEnforcer(confAuth.ModelPath, a)
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

	//e, _ := casbin.NewEnforcer(confAuth.ModelPath, confAuth.PolicyPath)
	e.AddNamedMatchingFunc("g2", "KeyMatch6", KeyMatch6)
	//e.AddNamedMatchingFunc("g2", "KeyMatch5", util.KeyMatch5)
	//e.AddNamedDomainMatchingFunc("g2", "KeyMatch6", KeyMatch6)

	// ok, err := e.Enforce("d1f7b7c1-c0b6-4707-aa17-5055b09b3ae8", "/basic-api/getUserInfo/1", "GET")
	enforceContext := casbin.EnforceContext{RType: "r", PType: "p2", EType: "e", MType: "m2"}
	ok, err := e.Enforce(enforceContext, "d1f7b7c1-c0b6-4707-aa17-5055b09b3ae8", "/diagnoseClass/1/diagnoseRow/aa?2", "GET")
	fmt.Println(ok)
	if err != nil {
		fmt.Println(err)
	}
	return e
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
