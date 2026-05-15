# 权限设计说明

## 1. 当前结论

当前 `monorepo` 线已经移除业务域模型。权限设计收敛为：

```text
user_id + organization_id -> role
role + organization_id -> resource_group + method
path + method -> resource_group
gateway -> auth.CheckAuthorization(user_id, path, method, organization_id)
```

其中 `organization_id` 是 Casbin domain。权限判断不再传递或解析 `service_code`、`scope_id`、`resource_group` 等旧字段；API 资源组只来自 `auth` 中已投影的全局 API 目录。

核心目标是先保证平台管理、菜单、API、资源组、组织、部门、站内信等现有能力稳定可控，并让同一用户在不同组织下拥有不同角色权限。

## 2. 职责边界

### 2.1 `user`

`user` 服务只维护平台用户身份：

- 用户基础信息
- 登录凭证
- 账号状态
- 密码维护

`user` 不维护角色、菜单和 API 权限。

### 2.2 `admin`

`admin` 是当前权限主数据中心，维护：

- `sys_role`
- `sys_resources`
- `sys_api_resources`
- `sys_user_role_binding`
- `sys_organization`
- `sys_user_organization`
- `sys_user_dept_membership`
- `sys_menu`
- `sys_dept`
- `sys_service_registry`
- `sys_projection_source_status`

当前角色同时影响两类能力：

- 菜单可见性：`sys_role.menus`
- 接口访问权：角色关联资源组 / API 目录后投影到 `auth`

### 2.3 `auth`

`auth` 不拥有权限主数据，只负责：

- 登录和 token 签发
- `casbin_rules` 授权投影加载
- `CheckAuthorization` 接口级授权判断
- 接收 `admin -> auth` 的快照和增量投影

`auth.casbin_rules` 是投影，不是主数据。

### 2.4 `gateway`

`gateway` 是统一入口：

1. 根据路由配置执行 JWT 校验。
2. 读取请求头 `x-organization-id` 作为当前组织 ID。
3. 调用 `auth.CheckAuthorization(user_id, path, method, organization_id)`。
4. 放行后转发到下游服务。

业务服务仍保留实例级数据校验职责，例如“只能操作自己的记录”这类规则不放在网关 Casbin 中实现。

## 3. 核心概念

### 3.1 `resource_group`

资源组是接口级授权粒度。API 不直接表达复杂业务动作，先按 `path + method` 映射到一个资源组。API 分组是全局的：

```text
g2, /admin-api/v1/roles, role
g2, /user-api/v1/users/{user_id}, user
```

示例：

- `default`
- `user`
- `role`
- `menu`
- `api`
- `data`
- `admin`
- `organization`
- `site_message_manage`

### 3.2 `role`

角色是组织内职责身份。

当前典型角色：

- `root`：超级管理员，Casbin 中拥有全局 `global` 域放行能力。
- `admin`：管理人员，通过资源组和菜单配置获得管理能力。
- `default`：普通用户，默认拥有基础接口和站内信收件箱能力。

角色主数据带 `organization_id`，投影到 `auth` 时使用 Casbin domain 隔离：

```text
g, userA, admin, orgA
g, userA, default, orgB
p2, admin, orgA, role, (GET|POST|PUT|DELETE)
p2, default, orgB, default, GET
```

因此不同组织可以复用相同 `role.value`，不会在投影层合并。

角色还带数据范围字段：

```text
data_scope: all | self_dept | self_dept_and_child | self | custom_depts
data_scope_dept_ids: custom_depts 时使用的部门 ID 集合
```

`data_scope` 不参与网关接口放行，由业务服务在接口放行后转换为数据查询条件。

### 3.3 `organization_id`

`organization_id` 当前用于承载当前组织 ID，由前端通过请求头 `x-organization-id` 传给 gateway，再由 gateway 传入 `auth.CheckAuthorization`，最终作为 Casbin domain。

它不是业务域，也不表示多平台或租户。当前作用是让用户角色绑定按组织隔离生效：

```text
sys_user_role_binding.organization_id
  -> g, user_id, role_value, organization_id
```

如果请求没有 `x-organization-id`，`auth` 使用默认组织 ID `9f740c1b-0210-4e3a-858d-d128edea924d` 作为 domain。请求显式带了组织 ID 时，只检查该组织 domain，不回退到默认组织角色，避免跨组织权限泄漏。

## 4. 数据模型

### 4.1 API 目录

`sys_api_resources` 当前字段重点是：

- `path`
- `method`
- `module`
- `module_description`
- `resources_group`

已移除：

- `business_key`
- `service_key`
- `service_code`
- `domain_code`

### 4.2 服务注册

`sys_service_registry` 当前字段重点是：

- `service_code`
- `service_name`
- `http_prefix`
- `grpc_service`
- `status`
- `projection_enabled`
- `description`

已移除：

- `domain_code`

### 4.3 投影源状态

`sys_projection_source_status` 当前以 `source_service` 为幂等键，记录权限投影同步状态。

已移除：

- `domain_code`

### 4.4 已删除表

业务域模型已删除：

- `sys_business_domain`

### 4.5 组织与部门

组织管理归 `admin` 服务：

- `sys_organization` 记录组织主数据。
- `sys_user_organization` 记录用户与组织的多对多成员关系。
- `sys_user_role_binding.organization_id` 记录用户在某个组织下拥有的角色。
- `sys_dept.organization_id` 记录部门归属组织。
- `sys_user_dept_membership` 记录用户在某个组织下绑定到哪个部门。
- 组织 ID 使用 UUID 字符串，避免通过递增 ID 推断组织数量；默认组织 ID 固定为 `9f740c1b-0210-4e3a-858d-d128edea924d`。

当前组织不是业务域、多平台或租户，但当前组织 ID 会作为 Casbin domain 参与网关级接口授权。组织相关权限仍按角色和资源组判断，只是“用户拥有哪些角色”按组织区分：

- 组织管理菜单：`/system/organization`
- 组织管理接口资源组：`organization`
- 管理员角色拥有组织管理菜单和 `organization` 资源组。
- 普通用户没有组织管理菜单。
- 默认内置角色（默认角色、超级管理员、管理员）归属默认组织 `default`；角色接口未传组织时使用当前组织，无法从登录上下文解析时回退默认组织 UUID。

用户可以切换自己的当前组织。当前组织同时作为数据上下文和接口授权 domain：

- `GET /admin-api/v1/my/organizations` 返回当前登录用户所属组织和当前组织。
- `PUT /admin-api/v1/my/current-organization` 切换当前组织，目标组织必须是该用户的有效成员组织。
- 两个接口使用 `default` 资源组，普通登录用户可访问。
- `sys_user_organization.is_primary` 当前作为“当前组织”标记；同一用户应只有一条有效当前组织记录。
- 前端切换组织后会重新拉取菜单并重建动态路由，因为菜单由当前组织角色绑定决定。
- 用户从某组织移除时，会清理该组织下的角色绑定和部门绑定，并重新同步权限投影；如果移除的是用户当前组织，当前组织会回退到有效剩余组织。

默认组织是全员组织：

- 所有用户都会被确保加入默认组织。
- 默认组织成员不能通过成员管理移入或移出。
- 没有显式角色绑定的用户，在默认组织下会合成 `default` 角色。
- 如果当前组织没有显式角色绑定，当前用户菜单会回退使用默认组织角色绑定，避免组织切换后菜单为空。
- `auth` 运行时也有 default-role fallback：当当前组织 domain 下用户没有显式角色绑定时，会检查同一 domain 下 `default` 角色是否允许访问当前 API。

部门接口仍使用 `dept` 资源组；部门树数据由组织隔离，未显式传 `organization_id` 时使用当前组织，创建/更新部门时父部门必须属于同一组织。角色接口、用户角色绑定接口和用户部门绑定接口未显式传 `organization_id` 时同样使用当前组织。

## 5. 运行时鉴权

当前 Casbin 模型主要使用：

- `g`：用户或 subject 到角色。
- `g2`：API path 到 API 资源组。
- `p2`：角色到 API 资源组的访问权。

请求链路：

```text
browser
  -> gateway
  -> JWT middleware
  -> Casbin middleware
  -> auth.CheckAuthorization
  -> downstream service
```

`CheckAuthorization` 使用：

- `user_id`
- `path`
- `method`
- `organization_id`

不再使用业务域字段。

`auth` 中当前主要 key 形态：

```text
g, user_id, role_value, organization_id
g2, api_path, resource_group
p2, role_value, organization_id, resource_group, method
```

`root` 角色在 Casbin 中保留全局放行能力：

```text
g, user_id, root, global
```

matcher 使用 `g(r.sub, "root", "global")`，不依赖当前请求组织。

## 6. 权限投影

当前真正跑通的权限投影链路是：

```text
admin DB
  -> admin projection source
  -> auth RegisterPermissionSnapshot / Apply*Delta
  -> auth.casbin_rules
```

常规变更：

- 角色变更：发送 role delta。
- API 变更：发送 API delta。
- 用户角色绑定变更：发送 binding delta。
- 资源组变更：回退为完整快照。
- 组织成员移除并清理角色绑定后：重新注册完整快照，确保 auth 删除该组织下的旧绑定。

失败时发送端会尝试完整快照恢复。

## 7. 站内信权限

站内信沿用当前平台权限模型：

- 普通用户通过 `/messages` 菜单和 `/common-api/v1/site-messages/my*` 接口查看收件箱。
- 管理人员通过 `/system/site-message` 菜单和 `site_message_manage` 资源组管理站内信。
- 站内信按当前组织投递和查询。`common` 从 `x-organization-id` 解析当前组织，消息和回执都会保存 `organization_id`。
- 发布、草稿、定时任务保存时记录当前组织；定时任务到期后按记录里的组织 ID 解析收件人，不依赖触发发布时的请求组织。
- 默认组织是全员组织，所以默认组织下发布等价于全员发布；其他组织下发布只投递该组织内启用成员。

不再使用摘要、指定用户或业务域字段。

## 8. 新增接口准则

新增或调整接口时：

1. 先在对应服务 proto 中定义接口。
2. 运行 `make api`。
3. 在 `sys_api_resources` 中登记唯一的 `path + method`。
4. 填写 `resources_group`。
5. 给需要的角色绑定对应资源组或 API 权限。
6. 同步 OpenAPI 到前端并更新调用点。

不再在 API 资源表中填写或设计 `service_code`、`domain_code`。
