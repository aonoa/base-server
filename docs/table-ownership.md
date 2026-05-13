# 表归属与服务使用说明

## 1. 目的

当前所有 Ent schema 都集中放在 `pkg/data/schema/` 下，便于统一生成 `pkg/data/ent/`，但这样会带来一个问题：

- 仅看 `pkg/data/schema/*.go`，看不出“这张表到底归哪个服务”

这份文档的目标就是把这个问题固定下来。

## 2. 先看结论

在当前代码里，**表归属不以 `pkg/data/schema/` 的目录位置为准，而以各服务自己的迁移列表为准**。

当前真正决定表归属的地方是：

- `app/user/service/internal/data/data.go` 里的 `userTables()`
- `app/admin/service/internal/data/data.go` 里的 `adminTables()`
- `auth` 服务自己的 Casbin adapter 存储

也就是说：

- `pkg/data/schema/` 只是“统一定义 Ent schema 的地方”
- “哪张表归哪个服务”要看具体服务启动时迁移了哪些表、直接读写了哪些表

## 3. 当前服务与数据库归属

### 3.1 `user` 服务

- 数据库：`user`
- 当前迁移表：
  - `sys_user`

### 3.2 `admin` 服务

- 数据库：`admin`
- 当前迁移表：
  - `sys_api_resources`
  - `sys_resources`
  - `sys_role`
  - `api_resources_roles`
  - `resource_roles`
  - `sys_user_role_binding`
  - `sys_menu`
  - `sys_dept`
  - `sys_log`
  - `sys_service_registry`
  - `sys_projection_source_status`

### 3.3 `auth` 服务

- 数据库：`auth`
- 当前不使用 `pkg/data/schema/*.go` 里的 Ent 表做主数据迁移
- 当前主要使用：
  - `casbin_rules`

`casbin_rules` 是授权投影存储，不是权限主数据。

### 3.4 `common` 服务

- 数据库：`common`
- 当前迁移表：
  - `sys_site_message`
  - `sys_site_message_receipt`

## 4. `pkg/data/schema` 到服务的映射

下面这张表以“当前代码实际迁移与直接读写关系”为准。

| Schema 文件 | 实际表名 | 所属数据库 | 主归属服务 | 说明 |
| --- | --- | --- | --- | --- |
| `pkg/data/schema/user.go` | `sys_user` | `user` | `user` | 平台用户基础信息、登录凭证、状态等 |
| `pkg/data/schema/role.go` | `sys_role` | `admin` | `admin` | 当前角色主数据已归 `admin` |
| `pkg/data/schema/resource.go` | `sys_resources` | `admin` | `admin` | 资源主数据已归 `admin` |
| `pkg/data/schema/api_resources.go` | `sys_api_resources` | `admin` | `admin` | API 资源目录主数据已归 `admin` |
| `pkg/data/schema/user_role_binding.go` | `sys_user_role_binding` | `admin` | `admin` | 当前用户角色绑定主数据已归 `admin` |
| `pkg/data/schema/menu.go` | `sys_menu` | `admin` | `admin` | 系统菜单 |
| `pkg/data/schema/dept.go` | `sys_dept` | `admin` | `admin` | 部门树 |
| `pkg/data/schema/log.go` | `sys_log` | `admin` | `admin` | 系统访问日志 / 操作日志 |
| `pkg/data/schema/service_registry.go` | `sys_service_registry` | `admin` | `admin` | 平台服务注册信息 |
| `pkg/data/schema/projection_source_status.go` | `sys_projection_source_status` | `admin` | `admin` | 权限投影源状态与同步观测信息 |
| `pkg/data/schema/site_message.go` | `sys_site_message` | `common` | `common` | 站内信发布记录、草稿、定时发布、撤回状态 |
| `pkg/data/schema/site_message_receipt.go` | `sys_site_message_receipt` | `common` | `common` | 站内信收件回执、已读/未读状态 |

## 5. 按服务看“直接拥有和直接写入”

### 5.1 `user` 服务直接拥有的表

#### `sys_user`

作用：

- 用户基础信息
- 用户名 / 密码
- 头像 / 状态 / 扩展信息

直接读写方：

- `user` 服务

间接依赖方：

- `auth` 服务通过 `user.ValidateUserAuth` / `user.GetUserAuthInfo` 读取用户认证信息
- `admin` 服务通过 `user.GetUserAuthInfo` 获取当前用户基础鉴权信息

备注：

- 当前 `sys_user` 的 Ent schema 里**没有** `role_id` 字段
- 用户角色绑定已经独立到 `sys_user_role_binding`

### 5.2 `admin` 服务直接拥有的表

#### `sys_role`

作用：

- 角色主数据

直接读写方：

- `admin`

间接依赖方：

- `auth` 不直接读这张表
- `admin` 会把角色快照 / delta 同步给 `auth`，最终投影到 `casbin_rules`

#### `sys_resources`

作用：

- 资源主数据

直接读写方：

- `admin`

备注：

- 当前资源增删改不会发细粒度 delta
- 仍然是触发一次完整 `RegisterPermissionSnapshot`

#### `sys_api_resources`

作用：

- API 权限目录主数据
- 记录 `path + method -> resources_group` 的接口目录映射

直接读写方：

- `admin`

备注：

- 前端“API 资源列表”改的是这张表
- `auth` 的 API 权限判断不直接读这张表，而是吃 `admin -> auth` 的投影同步
- `service_code` 不再存放在这张表；网关通过 `sys_service_registry.http_prefix` 推导服务归属和鉴权命名空间
- `resources_group` 是角色授权粒度

#### `api_resources_roles`

作用：

- API 资源与角色关系表

直接读写方：

- `admin`

#### `resource_roles`

作用：

- 资源与角色关系表

直接读写方：

- `admin`

#### `sys_user_role_binding`

作用：

- 用户到角色的绑定关系

直接读写方：

- `admin`

备注：

- 当前这是“谁拥有什么角色”的主数据表
- `auth` 不直接读这张表，而是由 `admin` 同步到 Casbin 分组策略

#### `sys_menu`

作用：

- 后台菜单定义

直接读写方：

- `admin`

备注：

- 当前菜单可见性仍然由角色里记录的 `menus` 字段参与裁剪

#### `sys_dept`

作用：

- 部门树

直接读写方：

- `admin`

#### `sys_log`

作用：

- 网关 / 后端请求日志、系统日志

直接读写方：

- `admin`

### 5.3 `common` 服务直接拥有的表

#### `sys_site_message`

作用：

- 站内信发布记录
- 草稿、定时发布、已发布、已撤回状态
- 发布人、发布时间、计划发布时间、接收人数

直接读写方：

- `common`

间接依赖方：

- `admin` 负责下发站内信管理菜单和管理接口权限
- `user` 提供全员发布时的有效用户列表

备注：

- 当前实现只支持全员发布，不支持按指定用户投递
- 定时发布是懒触发模式，由收件箱/管理接口访问时顺带提升到已发布

#### `sys_site_message_receipt`

作用：

- 记录每个用户是否已读
- 支持单条标记已读 / 未读、全部标记已读

直接读写方：

- `common`

备注：

- 已撤回的站内信会删除对应回执，不再出现在收件箱和未读数中

间接写入方：

- gateway 的 `httplog` 中间件最终会上报到 `admin`

#### `sys_service_registry`

作用：

- 平台服务注册表
- 描述服务编码、HTTP 前缀、gRPC 服务名、是否启用权限投影

直接读写方：

- `admin`

备注：

- 用于支撑服务注册、网关前缀解析和投影源观测
- 当前只记录服务信息，不再附加额外归属模型

#### `sys_projection_source_status`

作用：

- 权限投影源状态表
- 记录 source service 的同步模式、状态、最近快照版本、最近错误

直接读写方：

- `admin`

备注：

- 这是控制面观测表，不是 `auth` 的授权执行存储

## 6. 权限链路里的特殊表

下面这些表不在 `pkg/data/schema/` 里，但在权限链路里非常关键。

### `auth.casbin_rules`

作用：

- Casbin policy / grouping policy 存储

数据库：

- `auth`

主归属服务：

- `auth`

数据性质：

- 权限投影
- 不是权限主数据

来源：

- `auth` 启动时从本地 `casbin_rules` 执行 `LoadPolicy()`
- `admin` 启动后通过 `RegisterPermissionSnapshot`
- `admin` 日常写操作通过 `ApplyRoleDelta` / `ApplyApiDelta` / `ApplyUserRoleBindingDelta`

## 7. 当前最容易混淆的点

### 7.1 Schema 放一起，不代表表归属放一起

`pkg/data/schema/` 当前是“代码生成入口”，不是“服务边界目录”。

不要按下面这种方式理解：

- `schema 在共享目录` = `所有服务都直接持有这张表`

正确理解是：

- `schema 在共享目录` = `Ent 统一生成方便`
- `服务实际归属` = `哪个服务迁移它 + 哪个服务直接读写它`

### 7.2 历史 seed / README 不能替代当前迁移代码

仓库里仍然存在一些历史遗留信息，例如：

- 旧版 seed SQL
- 旧版 README 描述
- 迁移中的权限设计文档

这些内容可能还保留着“`auth` 管角色/资源/API 主数据”时期的痕迹。

因此当前判断表归属时，**优先级应该是**：

1. 服务启动时实际迁移了哪些表
2. 服务 repo/usecase 实际直接读写哪些表
3. 最后才参考 README / seed / 设计文档

### 7.3 `auth` 不再是角色/资源/API 主数据服务

就当前代码而言：

- `auth` 负责认证、Token、Casbin 执行、权限投影
- `admin` 负责角色、资源、API 资源目录、用户角色绑定主数据

所以当前权限领域里：

- 主数据在 `admin`
- 投影在 `auth`

### 7.4 `sys_user` 和 `sys_user_role_binding` 不在一个服务里

这是当前最容易误判的一点：

- `sys_user` 在 `user` 库，由 `user` 服务直接拥有
- `sys_user_role_binding` 在 `admin` 库，由 `admin` 服务直接拥有

也就是说：

- 平台用户身份归 `user`
- 平台后台角色绑定归 `admin`

## 8. 当前建议的排查路径

如果以后要查“某张表到底该去哪改”，建议按下面顺序找：

1. 先看 `pkg/data/schema/*.go`，确认表名
2. 再看哪个服务的 `internal/data/data.go` 把这张表放进了迁移列表
3. 再看哪个服务 repo/usecase 在直接增删改查这张表
4. 如果是权限相关，再额外确认它是不是只同步到了 `auth.casbin_rules`

## 9. 一句话版结论

当前代码里可以直接按下面记：

- `user` 服务：只直接拥有 `sys_user`
- `admin` 服务：直接拥有角色、资源、API 资源目录、用户角色绑定、菜单、部门、日志
- `auth` 服务：不拥有这些权限主数据表，只拥有 `casbin_rules` 这类授权投影
- `common` 服务：当前还没有落到 `pkg/data/schema` 的业务表
