package types

// ResourceType 资源类型
type ResourceType string

const (
	ResourceType_User ResourceType = "resource_addUser" // 添加用户权限
)

// PolicyType 权限类型
type PolicyType string

const (
	PolicyType_Page PolicyType = "policy_page" // 页面权限		act:read

	PolicyType_User   PolicyType = "policy_user"   // 用户权限	act:add/del
	PolicyType_Role   PolicyType = "policy_role"   // 角色权限	act:add/del
	PolicyType_Dom    PolicyType = "policy_dom"    // 域权限		act:add/del
	PolicyType_Passwd PolicyType = "policy_passwd" // 密码权限	act:add/del
	PolicyType_Policy PolicyType = "policy_policy" // 权限权限	act:add/del
)

// AuthPolicy 权限结构
type AuthPolicy struct {
	// 类型
	PolicyType string
	// 对象
	PolicyObj string
	// 行为
	PolicyAct string
}

type MenuAuthPolicy struct {
	policy []*AuthPolicy
}
