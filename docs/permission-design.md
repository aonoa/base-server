# 权限设计说明

## 1. 当前结论

当前 `monorepo` 线已经移除业务域模型。权限设计收敛为：

```text
user_id -> role
role -> resource_group + method
path + method -> resource_group
gateway path -> service_registry.http_prefix -> service_code
gateway -> auth.CheckAuthorization(user_id, path, method, service_code, scope_id, resource_group)
```

核心目标是先保证平台管理、菜单、API、资源组、站内信等现有能力稳定可控，不再为了未来多业务域扩展引入额外表、字段和页面。

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
- `sys_menu`
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
2. 根据 `sys_service_registry.http_prefix` 或内置前缀推导 `service_code`。
3. 根据 `sys_api_resources` 获取 API 的资源组。
4. 调用 `auth.CheckAuthorization`。
5. 放行后转发到下游服务。

业务服务仍保留实例级数据校验职责，例如“只能操作自己的记录”这类规则不放在网关 Casbin 中实现。

## 3. 核心概念

### 3.1 `service_code`

服务编码，用于区分接口由哪个服务提供，也用于 Casbin key 的命名空间。

当前不再把 `service_code` 放在 `sys_api_resources` 里维护。网关按 `sys_service_registry.http_prefix` 解析请求路径得到服务编码；未命中时再回退到内置前缀。

当前固定服务：

- `auth`
- `user`
- `admin`
- `common`

### 3.2 `resource_group`

资源组是接口级授权粒度。API 不直接表达复杂业务动作，先按 `path + method` 映射到一个资源组。

示例：

- `default`
- `user`
- `role`
- `menu`
- `api`
- `data`
- `admin`
- `site_message_manage`

### 3.3 `role`

角色是平台级职责身份。

当前典型角色：

- `root`：超级管理员，Casbin 中拥有全局放行能力。
- `admin`：管理人员，通过资源组和菜单配置获得管理能力。
- `default`：普通用户，默认拥有基础接口和站内信收件箱能力。

### 3.4 `scope_id`

`scope_id` 当前保留为可选扩展字段，用于请求头 `x-scope-id` 透传。当前默认链路不依赖它。

它不是业务域，也不表示多平台。后续如果需要组织、部门、租户或实例范围授权，可以在不恢复业务域模型的前提下继续使用。

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
- `service`
- `scope_id`

不再使用业务域字段。

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

失败时发送端会尝试完整快照恢复。

## 7. 站内信权限

站内信沿用当前平台权限模型：

- 普通用户通过 `/messages` 菜单和 `/common-api/v1/site-messages/my*` 接口查看收件箱。
- 管理人员通过 `/system/site-message` 菜单和 `site_message_manage` 资源组管理站内信。

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
