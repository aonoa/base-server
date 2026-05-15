# 组织域权限与 API 授权范围设计

状态：已实现核心链路，剩余业务数据过滤按服务逐步接入

适用版本线：`base-server/monorepo` + `vben-admin/monorepo`

## 1. 背景

当前系统已经有组织、部门、角色、用户角色绑定、API 目录、资源组和菜单能力，但权限模型仍存在几个需要收敛的问题：

- API 目录是否按组织复制不清晰。
- 组织管理员能否看到和分配全局 API 权限不清晰。
- 旧模型曾把服务/范围概念混进权限 key，组织域不是一等概念。
- 角色已经有 `organization_id`，但角色权限保存时还缺少“组织已开通范围”和“管理员可分配范围”的后端越权校验。
- 部门应只影响数据范围，不应进入网关接口鉴权域。

本设计目标是把权限模型收敛成：

```text
全局 API 目录 + 全局资源能力 + 组织可用范围 + 组织内角色授权
```

## 2. 设计结论

### 2.1 API 目录全局维护

`sys_api_resources` 继续作为全局 API 目录，不按组织复制。

```text
path + method -> resources_group
```

示例：

```text
/admin-api/v1/roles + GET -> role
/admin-api/v1/organizations + POST -> organization
/user-api/v1/users/{user_id} + GET -> user
```

组织不直接管理 API 记录，也不编辑 API path、method、module、resource_group。

### 2.2 组织只管理可用资源范围

平台管理员维护全局资源能力，并给组织开通可用范围。

组织管理员只能在当前组织已开通范围内给角色分配权限，不能越权分配全局能力。

```text
全局层：
sys_api_resources
sys_resources
sys_menu

组织层：
sys_organization_permission_scope
sys_role
sys_user_role_binding
sys_user_organization
sys_user_dept_membership
```

### 2.3 角色和用户角色绑定按组织隔离

角色属于组织：

```text
sys_role.organization_id
```

用户在组织下拥有角色：

```text
sys_user_role_binding.user_id
sys_user_role_binding.organization_id
sys_user_role_binding.role_id
```

同一个用户在不同组织可以拥有不同角色。网关鉴权时只使用当前组织域下的角色，不合并其他组织角色。

### 2.4 默认组织作为无组织头兜底

默认组织固定 ID：

```text
9f740c1b-0210-4e3a-858d-d128edea924d
```

当请求没有 `x-organization-id` 时，`auth` 使用默认组织作为权限域。

当请求显式带了 `x-organization-id` 时，只使用该组织域。目标模型不从当前组织回退到默认组织角色，避免跨组织权限泄漏。

### 2.5 部门只做数据范围

部门不进入网关 Casbin 模型。

部门用于组织内数据过滤，例如：

```text
用户属于组织 A 的部门 D1
业务查询时只看 D1 或 D1 子部门数据
```

接口能不能访问仍由组织角色决定。

数据权限使用角色上的 `data_scope` 表达：

```text
all                 全组织
self_dept           本部门
self_dept_and_child 本部门及下级
self                仅本人
custom_depts        指定部门集合
```

因此权限判断分两步：

```text
1. gateway/auth:
   user_id + organization_id + API -> allow/deny

2. service/biz:
   user_id + organization_id + dept_id + role.data_scope -> 数据过滤条件
```

组织决定用户当前使用哪套角色权限；部门决定接口放行后，数据能看到哪一部分。

## 3. 角色边界

### 3.1 平台管理员

平台管理员管理全局权限主数据：

- 维护 API 目录。
- 维护资源组和资源能力。
- 维护全局菜单。
- 给组织开通或回收可用权限范围。
- 创建组织和组织管理员。
- 查看全局投影状态和服务注册。

平台管理员可以跨组织操作，但这类能力必须由 `root` 或明确的平台级角色控制。

### 3.2 组织管理员

组织管理员只在当前 `x-organization-id` 组织内生效：

- 查看本组织成员。
- 管理本组织角色。
- 给本组织成员绑定本组织角色。
- 给本组织角色分配本组织已开通、且自己可分配的菜单和资源能力。
- 给本组织角色配置不超过自己可授权范围的数据范围。

组织管理员不能：

- 修改全局 API 目录。
- 修改全局资源组。
- 修改全局菜单。
- 查看或分配组织未开通能力。
- 给别人分配自己没有的能力。
- 分配 `root` 或平台治理类能力，除非自己就是平台管理员。

### 3.3 普通组织成员

普通成员只能使用自己在当前组织角色拥有的菜单和 API 权限。

如果成员在某个非默认组织下没有角色，则在目标模型中不自动继承默认组织角色。需要组织管理员显式授予该组织下的角色。

## 4. 目标 Casbin 模型

目标模型把组织作为 Casbin domain：

```ini
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act
p2 = sub, dom, obj, act
p3 = sub, dom, obj, act

[role_definition]
g = _, _, _
g2 = _, _
g3 = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
# 角色对普通资源组权限
m = (g(r.sub, p.sub, r.dom) &&
     g3(r.obj, p.obj) &&
     regexMatch(r.act, p.act)) ||
    g(r.sub, "root", "global")

# 角色对 API 资源组权限
m2 = r.obj == "/auth-api/v1/refresh" ||
     (g(r.sub, p2.sub, r.dom) &&
      g2(r.obj, p2.obj) &&
      regexMatch(r.act, p2.act)) ||
     g(r.sub, "root", "global")

# API 对普通资源组权限，当前保留但不是主链路
m3 = (g2(r.sub, p3.sub) &&
      r.dom == p3.dom &&
      g3(r.obj, p3.obj) &&
      regexMatch(r.act, p3.act)) ||
     g(r.sub, "root", "global")
```

策略语义：

```text
g, user_id, role_code, organization_id
p, role_code, organization_id, data_group, method
p2, role_code, organization_id, resource_group, method
p3, api_path_or_group, organization_id, data_group, method
g2, api_path, resource_group
g3, data_object, data_group
```

示例：

```text
g, userA, admin, orgA
g, userA, default, orgB

g2, /admin-api/v1/roles, role
g2, /user-api/v1/users, user

p2, admin, orgA, role, (GET|POST|PUT|DELETE)
p2, default, orgB, default, GET
```

请求检查：

```text
sub = user_id
dom = x-organization-id 或 默认组织 ID
obj = gateway path
act = HTTP method
```

`service_code` 不参与 role key、subject key 或 domain。它只作为服务注册、平台治理展示和排查定位的元数据，不是权限隔离维度。

超级管理员使用全局 root：

```text
g, root_user_id, root, global
```

matcher 使用 `g(r.sub, "root", "global")`，不依赖当前请求的 `dom`。

## 5. 数据模型设计

### 5.1 继续保留的全局表

#### `sys_api_resources`

全局 API 目录。

关键字段：

```text
id
path
method
module
module_description
resources_group
```

唯一约束：

```text
unique(path, method)
```

#### `sys_resources`

全局资源能力。

当前主要表示 API 资源组权限：

```text
id
name
type
value
method
description
```

示例：

```text
type = api
value = role
method = (GET|POST|PUT|DELETE)
```

#### `sys_menu`

全局菜单和路由能力。

角色菜单仍通过 `sys_role.menus` 保存菜单 ID 列表。

### 5.2 新增组织可用范围表

建议新增：

```text
sys_organization_permission_scope
```

用途：记录组织被平台开通了哪些菜单和资源能力。

字段建议：

```text
id                 UUID/string
organization_id    string
permission_type    string  // menu 或 resource
permission_ref     string  // menu.id 或 sys_resources.id，统一存 string
created_by         string
created_time
updated_time
```

唯一约束：

```text
unique(organization_id, permission_type, permission_ref)
```

说明：

- `permission_type=resource` 时，`permission_ref` 指向 `sys_resources.id`。
- `permission_type=menu` 时，`permission_ref` 指向 `sys_menu.id` 的字符串形式。
- 这个表不直接存 API path。API 通过 `sys_resources.value -> sys_api_resources.resources_group` 间接生效。

### 5.3 角色表约束

`sys_role` 继续保留：

```text
organization_id
menus
resource edge
data_scope
data_scope_dept_ids
```

建议新增唯一约束：

```text
unique(organization_id, value)
```

原因：

- 同一组织内角色编码不能重复。
- 不同组织可以复用相同角色编码，因为 Casbin domain 已经隔离。

角色保存时必须校验：

```text
role.organization_id == 当前操作组织
role.menus ⊆ 组织已开通 menu
role.resources ⊆ 组织已开通 resource
role.menus ⊆ 当前管理员可分配 menu
role.resources ⊆ 当前管理员可分配 resource
role.data_scope 是允许值
role.data_scope_dept_ids 只在 data_scope=custom_depts 时有效
role.data_scope_dept_ids 全部属于当前组织
```

当前实现口径：

- 已校验 `role.menus` / `role.resources` 不超过组织已配置可用范围。
- 已校验 `role.menus` / `role.resources` 不超过操作者当前组织角色拥有的可分配范围。
- 已校验 `role.data_scope` 不超过操作者可授权的数据范围：`all` 只能由拥有 `all` 的操作者授权，`custom_depts` 只能授权操作者已有的指定部门集合子集。
- 只有 bootstrap root 用户可以创建、更新或分配 `root`。
- 组织权限范围为空时表示该组织没有可分配的对应权限；角色不能选择超出已显式配置范围的菜单或资源。

`data_scope_dept_ids` 可在 v1 用 JSON 数组落在 `sys_role` 上，便于和现有 `menus` 字段一致；如果后续需要审计每个部门授权变化，再拆成独立关联表。

### 5.4 用户角色绑定约束

保存 `sys_user_role_binding` 时必须校验：

```text
目标用户是当前组织成员
目标角色属于当前组织
目标角色不是 root，除非操作者是平台管理员
操作者不能通过绑定角色给别人赋予自己不可分配的能力
```

### 5.5 部门绑定约束

`sys_user_dept_membership` 继续保持组织隔离：

```text
user_id
organization_id
dept_id
```

校验：

```text
目标用户是当前组织成员
目标部门属于当前组织
```

部门绑定变化不影响 Casbin 投影，只影响业务数据范围。

### 5.6 数据范围字段

建议在 `sys_role` 增加：

```text
data_scope           string  // all/self_dept/self_dept_and_child/self/custom_depts
data_scope_dept_ids  json    // custom_depts 指定部门集合，默认 []
```

默认值建议：

```text
root/admin -> all
default    -> self
```

角色拥有多个时，业务层按“权限最大化”合并数据范围：

```text
all 优先级最高，出现 all 即全组织
self_dept_and_child 合并所有本人部门及下级
self_dept 合并所有本人部门
custom_depts 合并所有指定部门
self 保留本人条件
```

最终查询条件应是这些范围的并集，而不是只取第一个角色。

## 6. 后端接口设计

### 6.1 平台权限主数据接口

仅平台管理员可用：

```text
GET    /admin-api/v1/apis
POST   /admin-api/v1/apis
PUT    /admin-api/v1/apis/{id}
DELETE /admin-api/v1/apis/{id}

GET    /admin-api/v1/resources
POST   /admin-api/v1/resources
PUT    /admin-api/v1/resources/{id}
DELETE /admin-api/v1/resources/{id}

GET    /admin-api/v1/menus
POST   /admin-api/v1/menus
PUT    /admin-api/v1/menus/{id}
DELETE /admin-api/v1/menus/{id}
```

组织管理员默认不应看到这些全局维护页面。

### 6.2 组织可用范围接口

新增平台管理员接口：

```text
GET /admin-api/v1/organizations/{organization_id}/permission-scope
PUT /admin-api/v1/organizations/{organization_id}/permission-scope
```

返回结构建议：

```json
{
  "organization_id": "org_a",
  "menu_ids": [1, 2, 3],
  "resource_ids": ["resource-user", "resource-role"]
}
```

保存时校验：

```text
organization_id 存在且有效
menu_ids 都存在
resource_ids 都存在
操作者应具备平台级组织权限范围管理能力
```

回收权限时建议 v1 默认阻止破坏性回收：

```text
如果某个被回收 menu/resource 仍被组织内角色使用，则返回 409
```

当前实现采用保守策略：若回收的菜单或资源仍被组织内角色使用，保存会失败；不做自动剪裁。

后续可以扩展 `force=true`，由后端自动从组织内角色中剪掉被回收权限并同步投影。

### 6.3 当前组织权限目录接口

新增接口供组织角色表单使用：

```text
GET /admin-api/v1/permission-catalog/current
```

返回当前组织可用且当前管理员可分配的权限：

```json
{
  "organization_id": "org_a",
  "menus": [],
  "resources": []
}
```

平台管理员查看某组织时可以带组织 ID：

```text
GET /admin-api/v1/organizations/{organization_id}/permission-catalog
```

组织管理员拿到的是过滤后的目录，不是全局 API 列表。

前端组织权限范围抽屉使用全局菜单和全局资源列表作为可开通来源；角色表单使用当前组织权限目录作为可分配来源。

### 6.4 角色接口

现有角色接口继续使用，但语义收敛为当前组织：

```text
GET    /admin-api/v1/roles
POST   /admin-api/v1/roles
PUT    /admin-api/v1/roles/{id}
DELETE /admin-api/v1/roles/{id}
```

规则：

- 未传 `organization_id` 时使用当前 `x-organization-id`。
- 非平台管理员不能跨组织传 `organization_id`。
- 创建和更新角色时，菜单和资源必须通过组织可用范围与操作者可分配范围校验。

### 6.5 用户角色绑定接口

现有接口继续使用：

```text
GET    /admin-api/v1/user-role-bindings/{user_id}
PUT    /admin-api/v1/user-role-bindings/{user_id}
DELETE /admin-api/v1/user-role-bindings/{user_id}
```

规则：

- 未传 `organization_id` 时使用当前 `x-organization-id`。
- 只能绑定当前组织内角色。
- 只能给当前组织成员绑定角色。
- 非平台管理员不能绑定 `root` 或平台治理角色。

## 7. 后端越权校验设计

组织管理员越权风险主要不在网关，而在 `admin` 服务的权限管理写接口。

### 7.1 操作者权限上下文

后端需要从请求上下文解析：

```text
actor_user_id
current_organization_id
is_platform_admin
actor_roles_in_current_org
actor_grantable_menus
actor_grantable_resources
actor_grantable_data_scope
```

`actor_grantable_*` 的计算规则：

```text
当前组织已开通范围 ∩ 操作者当前组织角色拥有范围
```

平台管理员可以绕过 `actor_grantable_*`，但仍应校验目标资源和目标部门存在。

### 7.2 写角色时校验

新增角色：

```text
target_organization_id == current_organization_id
selected_menus ⊆ organization_scope.menus
selected_resources ⊆ organization_scope.resources
selected_menus ⊆ actor_grantable_menus
selected_resources ⊆ actor_grantable_resources
```

更新角色：

```text
role.organization_id == current_organization_id
不能把普通角色升级成 root
不能给角色添加自己不可分配的 menu/resource
不能设置超出自己可管理部门范围的 data_scope/custom_depts
```

删除角色：

```text
role.organization_id == current_organization_id
如果角色仍被用户绑定，返回 409 或先解绑后删
不能删除当前组织最后一个管理员角色，除非平台管理员强制处理
```

### 7.3 绑定用户角色时校验

```text
target_user ∈ current_organization.members
target_roles 全部属于 current_organization
target_roles 的权限集合 ⊆ actor_grantable_*
target_roles 不包含 root，除非 actor 是平台管理员
target_roles 的 data_scope 不能让目标用户获得操作者不可授权的数据范围
```

### 7.4 管理组织可用范围时校验

只有平台管理员可以修改组织可用范围。

回收组织可用范围时：

```text
如果组织内角色还在使用该权限，默认拒绝并返回冲突信息
```

这样可以避免平台管理员误回收后导致组织菜单或接口权限突然丢失。

## 8. 权限投影设计

### 8.1 投影数据来源

`admin` 仍是权限主数据中心。

投影到 `auth` 时：

```text
sys_role -> p2
sys_user_role_binding -> g
sys_api_resources -> g2
```

组织可用范围表不直接投影到 `auth`。它只用于 `admin` 写接口校验和前端目录过滤。

`data_scope` 也不直接投影到 `auth`。它由业务服务在接口放行后读取并转换为数据查询条件。

### 8.2 `auth` proto 调整

`PolicyRole` 需要携带组织 ID：

```proto
message PolicyRole {
  int64 id = 1;
  string name = 2;
  string value = 3;
  bool status = 4;
  string remark = 5;
  repeated int32 menu_ids = 6;
  repeated PolicyRoleResource resources = 7;
  string organization_id = 9;  // 新增，作为 Casbin dom
  string data_scope = 10;      // 不用于 auth 鉴权，可作为同步元数据保留
  repeated int64 data_scope_dept_ids = 11;
}
```

`PolicyUserRoleBinding.organization_id` 作为组织域：

```text
organization_id = sys_user_role_binding.organization_id
```

### 8.3 投影后的 Casbin 规则

角色权限：

```text
p2, role_value, organization_id, resource_value, resource_method
```

用户到角色：

```text
g, user_id, role_value, organization_id
```

API 到资源组：

```text
g2, api_path, resources_group
```

超级管理员：

```text
g, root_user_id, root, global
```

root 判断不使用 `r.dom`。无论请求处于哪个组织域，只要用户拥有 `global` 域下的 `root` 分组，即拥有全部接口权限。

### 8.4 清理与重建

当前模型不使用 `service_code` 做 key 前缀，`auth` 的投影重建直接清空后按快照重建。

建议优先采用：

```text
RegisterPermissionSnapshot(admin) -> 清空现有投影 -> 重建 admin 投影
```

当前使用全量 `ClearPolicy + apply snapshot`，前提是当前只有 `admin` 一个权限主数据源。

如果后续确实有多个投影源，需要在 `auth.casbin_rules` 或旁表中增加 source metadata，而不是把 `service_code` 塞回权限 key。

## 9. 网关鉴权流程

目标流程：

```text
browser
  -> request.ts 附加 x-organization-id
  -> gateway JWT
  -> gateway 解析 path/method
  -> auth.CheckAuthorization
  -> Casbin enforce(user_id, organization_id, path, method)
```

`auth.CheckAuthorization` 处理：

```text
dom = req.organization_id
if dom == "":
  dom = defaultOrganizationID

allowed = Enforce(RoleToApi, req.user_id, dom, req.path, method)
```

不再执行：

```text
subject:admin:{scope}:{user_id}
role:admin:{role_value}
admin:{path}
api:admin:{resource_group}
```

也不再使用 service 命名空间或旧 subject。

## 10. 业务数据范围流程

网关只判断接口能不能访问。接口放行后，业务服务或 `admin` service 的 biz 层负责把角色数据范围转换成查询条件。

### 10.1 数据范围上下文

业务层需要解析：

```text
user_id
organization_id
user_dept_id
role.data_scope
role.data_scope_dept_ids
dept tree in organization
```

如果用户在当前组织没有部门：

```text
self_dept           -> 空结果
self_dept_and_child -> 空结果
custom_depts        -> 按配置部门集合
self                -> user_id = 当前用户
all                 -> 全组织
```

### 10.2 范围到查询条件

推荐统一抽象成：

```text
DataScopeFilter {
  organization_id
  allow_all
  user_ids
  dept_ids
}
```

转换规则：

```text
all:
  allow_all = true

self:
  user_ids = [current_user_id]

self_dept:
  dept_ids = [current_user_dept_id]

self_dept_and_child:
  dept_ids = [current_user_dept_id + descendant_dept_ids]

custom_depts:
  dept_ids = role.data_scope_dept_ids
```

最终 SQL/Ent 条件必须始终带组织：

```text
organization_id = current_organization_id
AND (
  allow_all
  OR owner_user_id IN user_ids
  OR dept_id IN dept_ids
)
```

不同业务表的字段名可以不同，但语义要统一。例如用户列表可以用 `sys_user_dept_membership` 过滤部门成员；文档、订单等业务表可以使用自己的 `organization_id`、`dept_id`、`owner_user_id` 字段。

### 10.3 多角色合并

同一用户在当前组织有多个角色时，数据范围取并集。

合并顺序：

```text
任一角色是 all -> allow_all
self_dept_and_child -> 合并本人部门及所有下级部门
self_dept -> 合并本人部门
custom_depts -> 合并指定部门集合
self -> 合并当前用户本人
```

不能只按角色优先级取第一个角色，否则会丢失其他角色授予的数据范围。

### 10.4 管理接口与业务接口

`admin` 服务自己的用户、部门、角色、组织成员查询也应使用同一套数据范围规则，但平台管理员和 `root` 可以绕过数据范围。

普通业务服务不应该直接依赖 Casbin 查询 `data_scope`。推荐由 `admin` 提供当前用户组织角色和数据范围的内部接口，或在后续抽出共享的权限上下文包。

## 11. 前端设计

### 11.1 平台管理员视图

平台管理员可见：

- API 资源列表。
- 资源组管理。
- 菜单管理。
- 服务注册。
- 投影源状态。
- 组织详情里的“可用权限范围”。

组织可用范围页面推荐放在组织详情或侧边抽屉：

```text
组织列表 -> 组织详情 -> 可用权限
```

可用权限按菜单树和资源组展示，不直接让用户逐条勾底层 API。

### 11.2 组织管理员视图

组织管理员可见：

- 当前组织成员。
- 当前组织部门。
- 当前组织角色。
- 当前组织角色权限表单。
- 当前组织角色数据范围配置。

角色权限表单的数据来自：

```text
GET /admin-api/v1/permission-catalog/current
```

组织管理员看到的是过滤后的菜单和资源能力。

不显示全局 API 维护页。如果需要解释某个资源组包含哪些 API，可以在角色表单中用只读抽屉展示：

```text
资源组 user
  GET /user-api/v1/users
  POST /user-api/v1/users
```

这个只读明细也必须按当前组织可用资源过滤。

角色数据范围表单：

```text
data_scope 下拉选择
custom_depts 时显示部门树多选
```

组织管理员只能选择后端返回的可授权数据范围；前端限制只是体验，后端仍必须做越权校验。

### 11.3 请求头

前端继续使用当前组织 store 维护 `x-organization-id`。

规则：

- 当前组织已加载且有效时发送 `x-organization-id`。
- 未加载或未选择时不发送，后端使用默认组织兜底。
- 切换组织后重新拉菜单、权限目录和数据范围上下文。

## 12. 迁移方案

### 12.1 数据库迁移

新增表：

```text
sys_organization_permission_scope
```

新增索引：

```text
sys_role unique(organization_id, value)
```

新增角色字段：

```text
sys_role.data_scope
sys_role.data_scope_dept_ids
```

### 12.2 当前数据初始化

项目尚未投入实际使用，不再设计旧数据格式迁移和权限回填策略。当前初始化口径是：

- 默认组织显式写入现有全部菜单和资源组权限范围。
- `root/admin/default` 内置角色直接使用当前 `organization_id + data_scope` 字段。
- 新建非默认组织默认没有可分配权限范围，必须由平台 root 显式配置。
- 组织权限范围为空不代表不限权，而是代表没有可分配的对应权限。

### 12.3 Casbin 投影迁移

上线步骤：

1. 部署支持 domain 模型的 `auth`。
2. 部署携带 `organization_id` 投影的 `admin`。
3. 清空或重建 `auth.casbin_rules`。
4. 执行 `make sync-admin-projection`。
5. 重启或热加载 gateway/auth/admin。

### 12.4 前端迁移

前端需要同步：

- 角色表单权限目录改为当前组织权限目录。
- 角色表单增加数据范围配置。
- API 目录和资源管理页面只给平台管理员显示。
- 组织详情增加可用权限配置入口。
- 组织切换后重新加载菜单、当前组织权限目录和数据范围上下文。

## 13. 验收标准

### 13.1 Auth

必须验证：

- 同一用户在组织 A 是 admin，在组织 B 是 default，请求带组织 B 时不能获得组织 A 的 admin 权限。
- 不传 `x-organization-id` 时使用默认组织权限。
- `root` 用户通过 `global` domain 放行。
- `g2` 仍支持 `/path/{id}` 模板匹配。
- - `root` 使用 `global` 域放行，不依赖当前请求的 `x-organization-id`。

建议测试：

```bash
go test ./app/auth/service/internal/biz
```

### 13.2 Admin

必须验证：

- 组织管理员不能给角色分配组织未开通资源。
- 组织管理员不能给角色分配自己没有的资源。
- 组织管理员不能修改其他组织角色。
- 组织管理员不能给其他组织成员绑定角色。
- 组织管理员不能绑定 `root`。
- 组织管理员不能设置超出自己可授权范围的数据范围。
- `custom_depts` 只能选择当前组织部门。
- 平台管理员可以给组织开通资源。
- 回收仍被角色使用的资源时返回冲突。

建议测试：

```bash
go test ./app/admin/service/internal/biz ./app/admin/service/internal/data
```

### 13.3 Gateway

必须验证：

- `x-organization-id=org_a` 和 `x-organization-id=org_b` 得到不同鉴权结果。
- 不传 `x-organization-id` 使用默认组织。
- 切换组织后菜单和接口权限一致，不出现“能看到菜单但接口 403”的错配。

### 13.4 Service/Biz Data Scope

必须验证：

- `all` 可以看到当前组织全部数据，但不能跨组织。
- `self_dept` 只能看到本人部门数据。
- `self_dept_and_child` 可以看到本人部门及下级部门数据。
- `self` 只能看到本人拥有或创建的数据。
- `custom_depts` 只能看到指定部门集合的数据。
- 多角色数据范围按并集生效。
- 用户无部门时，`self_dept` 和 `self_dept_and_child` 返回空结果。

### 13.5 Frontend

必须验证：

- 平台管理员能看到 API、资源、组织可用权限配置。
- 组织管理员看不到全局 API 维护页。
- 组织管理员角色表单只显示当前组织可用且自己可分配的菜单和资源。
- 组织管理员角色表单只能选择自己可授权的数据范围。
- `custom_depts` 模式下只能选择当前组织部门。
- 切换组织后重新加载菜单、权限目录和数据范围上下文。

建议测试：

```bash
pnpm -F @vben/web-antd run typecheck
```

如果项目既有类型错误阻断，需要记录具体已知错误，并补充针对权限页面的单测或组件测试。

## 14. 分阶段实施计划

### 阶段 1：文档和模型确认

- 已确认组织域 Casbin 模型。
- 已确认 `data_scope` 取值。
- 已确认默认组织作为无组织头兜底。

### 阶段 2：后端数据与接口

- 已新增 `sys_organization_permission_scope` Ent schema。
- 已给 `sys_role` 增加 `data_scope` 和 `data_scope_dept_ids`。
- 已增加组织可用范围查询和保存接口。
- 已增加当前组织权限目录接口。
- 已给角色保存、用户角色绑定、组织成员管理补齐越权校验。
- 业务数据过滤仍按后续具体业务接口逐步接入。

### 阶段 3：Casbin domain 重构

- 已修改 `auth` text model。
- 已修改投影 proto 和 sender。
- 已修改 admin 投影构造。
- 已修改 auth snapshot/delta 应用。
- 已修改 `CheckAuthorization` 的默认组织兜底逻辑。

### 阶段 4：前端页面

- 已在组织列表增加可用权限配置抽屉。
- 已让角色表单改用当前组织权限目录。
- 已增加角色数据范围配置。
- 平台 API/资源/菜单管理是否隐藏可按后续菜单授权继续收紧。

### 阶段 5：初始化与回归

- 已让默认组织显式初始化全量菜单和资源权限范围。
- 已重建当前权限投影生成链路。
- 需要在本地用默认组织、普通组织、平台管理员、组织管理员四类账号继续做浏览器验收。

## 15. 待评审问题

1. 组织管理员是否允许查看只读 API 明细抽屉，还是完全只展示资源组名称。
2. 数据范围上下文由 `admin` 提供内部接口，还是先在需要数据过滤的服务内各自实现。
