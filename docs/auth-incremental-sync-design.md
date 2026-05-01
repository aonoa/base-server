# auth 权限投影同步实现说明（当前代码）

## 1. 背景

当前权限投影链路是：

- `auth` 启动时在 `NewAuthUsecase()` 里只执行 `LoadPolicy()`，直接从自己的 `casbin_rules` 恢复策略。
- 权限主数据服务启动后，通过 `RegisterPermissionSnapshot(...)` 向 `auth` 注册一次完整快照，失败时按退避重试。
- 权限主数据服务日常写操作按对象发送 delta；delta 失败时回退到一次完整快照注册。

当前代码里的真相边界是：

- 当前仓库运行中的权限主数据仍主要来自 `admin`。
- `auth` 库里的 `casbin_rules` 是授权投影，不是主数据。
- `walk-routes` 只用于路由发现/对账，不会自动回写或替代 `sys_api_resources`。
- 修改 proto 或 HTTP 路由，不会自动推导出对应的 Casbin API 权限变更；只有修改权限主数据服务里的 API 目录，才会触发 `ApplyApiDelta`。
- 对 API 权限目录来说，同一个 `path + method` 只能有一个主业务域 `domain_code`；多业务消费或跨域编排不等于多业务域并列 owner。

## 2. v1 目标

当前实现已经具备“源服务向 `auth` 注入投影”的基础能力，但仓库内真正跑通的仍主要是 `admin -> auth` 这一条链路。

当前已落地的目标：

- `auth` 启动时不再主动回拉 `admin`。
- 权限主数据服务启动成功后主动注册完整快照。
- 角色、API、用户角色绑定已经支持按对象发送 delta。
- 资源增删改当前仍直接走完整快照回放。
- delta 失败时会回退到完整快照，保留恢复能力。
- 发送端已抽成通用投影 sender，后续业务服务可以直接复用。
- 投影发送端已支持把最近一次快照 / delta 的同步状态自动回写到 `admin.sys_projection_source_status`。

## 3. 职责边界

- 权限主数据服务负责输出标准化投影。
- `auth` 不负责“发现真相”，只维护授权投影和 Casbin 执行状态。
- `auth` 库中的 `casbin_rules` 仍然是投影存储，不是主数据。
- 路由发现结果不是 Casbin API 权限目录，它只用于发现和对账。

## 4. 设计原则

### 4.1 启动注册替代启动回拉

- 不再让 `auth` 在启动阶段主动调用 `admin`。
- 改成权限主数据服务启动完成后主动向 `auth` 注册完整权限快照。
- 这样 `auth` 单独重启时也能直接从自己的库里 `LoadPolicy()` 恢复服务能力。

### 4.2 日常同步只发足量 delta

- `auth` 不应在处理增量时再次回查主数据服务。
- 发送方发给 `auth` 的 delta 必须带足够的 `before/after` 信息。
- 这样 `auth` 才能在本地完成 rename、remove、upsert，而不重新走全量拉取。

### 4.3 保留完整快照回退路径

- 增量同步失败时，不要直接让链路失效。
- 发送方应该能回退到一次完整快照注册。
- 当前代码已不再依赖 `ReLoadPolicy()` 作为主链路。

## 5. 推荐 RPC 契约

这版继续沿用 `api/protos/auth/service/v1/auth.proto`，由权限主数据服务调用 `auth`。

### 5.1 完整快照注册

```proto
rpc RegisterPermissionSnapshot(RegisterPermissionSnapshotRequest) returns (google.protobuf.Empty);
```

```proto
message RegisterPermissionSnapshotRequest {
  string source_service = 1;
  uint64 revision = 2;
  repeated PolicyRole roles = 3;
  repeated PolicyApi apis = 4;
  repeated PolicyUserRoleBinding bindings = 5;
}
```

当前实现语义：

- `source_service` 表示投影来源服务；当前仓库实际主要用的是 `admin`，但字段和发送端都已经按通用来源服务保留。
- `revision` 由发送方单调递增，建议先用毫秒时间戳或数据库变更版本号。
- 当前实现分两种模式：
  - 遗留快照：如果消息里没有显式 `service/scope` 命名空间，按旧模式 `ClearPolicy()` 后整体重建
  - 新命名空间快照：如果消息里显式带了 `service`，则先按 `source_service` 移除该源旧投影，再把新快照合并进现有 Casbin 状态
- 这意味着：
  - 当前 `admin` 仍可作为遗留单源继续工作
  - 后续新业务服务可以按 `source_service` 并行向 `auth` 注入，不会互相覆盖

### 5.2 角色增量

```proto
rpc ApplyRoleDelta(ApplyRoleDeltaRequest) returns (google.protobuf.Empty);
```

```proto
message ApplyRoleDeltaRequest {
  string source_service = 1;
  uint64 revision = 2;
  PolicyRole before = 3;
  PolicyRole after = 4;
}
```

当前实现语义：

- 新增角色：`before` 为空，`after` 非空。
- 更新角色：`before`、`after` 都非空。
- 删除角色：`before` 非空，`after` 为空。
- 如果角色 `value` 被改名，必须同时带 `before.value` 和 `after.value`，让 `auth` 在本地完成 rename。

### 5.3 API 资源目录增量

```proto
rpc ApplyApiDelta(ApplyApiDeltaRequest) returns (google.protobuf.Empty);
```

```proto
message ApplyApiDeltaRequest {
  string source_service = 1;
  uint64 revision = 2;
  PolicyApi before = 3;
  PolicyApi after = 4;
}
```

语义约束：

- 新增 API：`before` 为空，`after` 非空。
- 更新 API：`before`、`after` 都非空。
- 删除 API：`before` 非空，`after` 为空。
- 如果 `path` 或 `resources_group` 变化，必须带完整 `before/after`。
- 同一个 `path + method` 不应由多个投影源并列上报为不同业务域 owner；跨域协作应在主业务域 API 下编排。
- `auth.ApplyApiDelta()` 当前会优先用 `UpdateNamedGroupingPolicy(g2, oldRule, newRule)` 更新 API -> 资源组映射。
- 如果旧规则不存在，则回退为“删除旧规则 + 添加新规则”。
- 这里更新的是 Casbin 的 `g2` 映射，不会自动修改 `admin` 库里的 `sys_api_resources` 以外的任何路由定义。

### 5.4 用户角色绑定增量

```proto
rpc ApplyUserRoleBindingDelta(ApplyUserRoleBindingDeltaRequest) returns (google.protobuf.Empty);
```

```proto
message ApplyUserRoleBindingDeltaRequest {
  string source_service = 1;
  uint64 revision = 2;
  PolicyUserRoleBinding before = 3;
  PolicyUserRoleBinding after = 4;
}
```

当前实现语义：

- 绑定新增：`before` 为空，`after` 非空。
- 绑定更新：`before`、`after` 都非空。
- 绑定删除：`before` 非空，`after` 为空。
- `PolicyUserRoleBinding` 必须直接携带 `role_value`，不要只传 `role_id`，否则 `auth` 还得再去查 `admin`。

### 5.5 投影消息定义

```proto
message PolicyRole {
  int64 id = 1;
  string name = 2;
  string value = 3;
  bool status = 4;
  string remark = 5;
  repeated int32 menu_ids = 6;
  repeated PolicyRoleResource resources = 7;
  string service = 8;
}

message PolicyRoleResource {
  string id = 1;
  string type = 2;
  string value = 3;
  string method = 4;
}

message PolicyApi {
  string id = 1;
  string path = 2;
  string method = 3;
  string description = 4;
  string module = 5;
  string module_description = 6;
  string resources_group = 7;
  string service = 8;
}

message PolicyUserRoleBinding {
  string id = 1;
  string user_id = 2;
  int64 role_id = 3;
  string role_value = 4;
  string create_time = 5;
  string update_time = 6;
  string service = 7;
  string scope_id = 8;
}
```

## 6. 与当前实现的映射

`auth` 侧已经有一批可复用的本地 Casbin 变更原语，不需要推翻重写：

- 角色相关：
  - `renameRoleBindings`
  - `removeRolePolicies`
  - `removeRoleBindings`
  - `AddPolicy`
- API 相关：
  - `AddApiToGroup`
  - `updateApiGroup`
  - `removeApiGroup`
- 用户角色绑定相关：
  - `AddUserRoles`
  - `RemoveFilteredNamedGroupingPolicy` 可直接复用

因此，增量接口不需要让 `auth` 再次回查主数据，只要发送方把 `before/after` 发全，`auth` 就能在本地原子改 Casbin 投影。

补充说明：

- API 接口级权限判断走 `p2 + g2`。
- `g2` 表示 `apiPath -> api:<resources_group>` 的分组映射。
- 当前仓库里，前端编辑“API 资源列表”里的 `path`、`resources_group`、`service_code`、`domain_code`，本质上是在改 `sys_api_resources` 主数据；这会通过 `ApplyApiDelta` 更新 `g2`，并把 `service_code` 投影到 `PolicyApi.service`。
- `service_code` 在当前投影里用于运行时命名空间兼容和技术服务隔离；目标业务授权边界仍应收敛到 `domain_code + resource_group + action`。
- 仅修改 proto/http route，不会自动改 `sys_api_resources`，因此也不会自动改 Casbin。

## 6.1 通用发送端

当前代码已在 [`pkg/authx/projection.go`](../pkg/authx/projection.go) 增加通用投影发送层：

- 统一封装 `RegisterPermissionSnapshot`
- 统一封装 `ApplyRoleDelta`
- 统一封装 `ApplyApiDelta`
- 统一封装 `ApplyUserRoleBindingDelta`

这样当前 `admin` 仍然可以继续作为投影源服务，后续新增业务服务时也不需要再复制一套 `auth` proto 拼装逻辑。

当前代码还已在 [`pkg/authx/projection_source.go`](../pkg/authx/projection_source.go) 增加通用源服务适配接口：

- `ProjectionSource`
- `BuildSnapshotFromSource`
- `ProjectionClient.RegisterSourceSnapshot`

这意味着后续业务服务只要实现：

- `SourceService()`
- `ListProjectionRoles(ctx)`
- `ListProjectionAPIs(ctx)`
- `ListProjectionBindings(ctx)`

就能直接走标准快照注册链路。

当前 `admin` 已经先切成这一套样板实现，见：

- [`app/admin/service/internal/data/projection_source.go`](../app/admin/service/internal/data/projection_source.go)
- [`app/admin/service/internal/data/data.go`](../app/admin/service/internal/data/data.go)

另外，当前代码也已在 [`pkg/authx/projection_sync.go`](../pkg/authx/projection_sync.go) 增加两块通用同步辅助：

- `NewProjectionStartupSyncServer`：源服务启动后自动重试注册完整快照
- `SyncProjectionDelta`：delta 失败时统一回退到完整快照

当前 `admin` 已经切到这套通用 helper，上线新业务服务时可以直接复用同样的启动同步和失败回退模式。

当前代码还已在 [`pkg/authx/projection_status.go`](../pkg/authx/projection_status.go) 增加通用投影状态上报能力：

- `ProjectionStatus`
- `ProjectionStatusReporter`
- `AdminProjectionStatusReporter`

发送端在调用 `RegisterPermissionSnapshot` / `Apply*Delta` 后，会以 `source_service` 为幂等键，把最近一次同步模式、同步状态、最后快照版本、最后错误信息回写到 `admin` 控制面。

`auth` 侧当前也已经落地了源级替换逻辑：

- 对显式 `service` 命名空间的快照，按 `source_service` 删除旧投影后再合并新投影
- 对旧格式快照，保留整体清空后重建的兼容行为

## 6.2 当前已落地的 service / scope 扩展

当前代码已经落地以下兼容扩展：

- `CheckAuthorizationRequest` 新增 `service`、`scope_id`
- `PolicyRole` 新增 `service`
- `PolicyApi` 新增 `service`
- `PolicyUserRoleBinding` 新增 `service`、`scope_id`

当前行为：

- `gateway` 会优先按 `admin.sys_service_registry.http_prefix` 推导 `service`
- 如果服务注册表未命中，再回退到仓库内置的 `/admin-api`、`/user-api` 等老前缀规则
- `gateway` 允许通过请求头 `x-scope-id` 透传 `scope_id`
- `auth` 在本地 Casbin key 上对 `service` 和 `scope_id` 做名字空间编码
- 如果请求带了 `service/scope` 但没有命中新投影，`auth` 仍会回退尝试旧的无命名空间 key，保证现有链路兼容

## 7. 建议的服务内执行流程

### 7.1 auth 启动

当前代码已实现为：

- `NewEnforcer()` 创建 Enforcer，并开启 `AutoSave`
- `NewAuthUsecase()` 中执行 `LoadPolicy()`
- 启动阶段不访问 `admin`

### 7.2 源服务启动注册

当前代码通过启动期同步逻辑在源服务启动后执行：

1. 从本地数据库读出完整角色、API 资源目录、用户角色绑定。
2. 组装 `RegisterPermissionSnapshotRequest`。
3. 调用 `auth.RegisterPermissionSnapshot(...)`。
4. 失败则按固定退避重试，直到成功。

成功后会持续等待；失败则按指数退避重试。

同时：

- 成功会把投影源状态更新为 `synced`
- 失败会把投影源状态更新为 `error`
- `last_snapshot_revision`、`last_sync_time`、`last_error` 会一起更新到 `sys_projection_source_status`

### 7.3 源服务日常写操作

当前代码已经是下面这组行为：

- `AddRole/UpdateRole/DelRole` -> `ApplyRoleDelta`
- `AddApi/UpdateApi/DelApi` -> `ApplyApiDelta`
- `UpsertUserRoleBinding/DeleteUserRoleBinding` -> `ApplyUserRoleBindingDelta`

资源变更当前仍然这样处理：

- `AddResource/UpdateResource/DelResource` 不发 `ResourceDelta`
- 直接触发一次 `RegisterPermissionSnapshot`

这意味着：

- 资源变更粒度还比较粗
- 但实现更简单，失败恢复路径也统一

## 8. 失败回退策略

当前代码的失败回退策略是：

1. 优先发送 delta。
2. delta 失败时，立即尝试发送一次 `RegisterPermissionSnapshot`。
3. 如果快照也失败，向上返回错误，并保留现有 `casbin_rules` 不动。

这样可以保证：

- 常态是增量更新。
- 异常时还有完整修复手段。
- 迁移过程中不需要一次性删除旧逻辑。

## 9. 白名单与鉴权建议

新增的 `auth` 内部 RPC 都会在无用户 JWT 场景下被 `admin` 调用，因此需要加入 `auth` 服务白名单：

- `/api.auth.service.v1.AuthService/RegisterPermissionSnapshot`
- `/api.auth.service.v1.AuthService/ApplyRoleDelta`
- `/api.auth.service.v1.AuthService/ApplyApiDelta`
- `/api.auth.service.v1.AuthService/ApplyUserRoleBindingDelta`

v1 可以继续沿用当前“内网 + whitelist + 共享 `api_key`”的方式，不额外引入服务身份体系。

如果后面要继续收敛风险，再单独补：

- `x-service-key`
- 服务级签名
- mTLS

## 10. 当前未覆盖的点

当前代码仍未覆盖或未自动化的点：

- 多业务服务同时向 `auth` 注入授权投影
- 跨服务、跨业务域投影命名空间隔离的完整目标态落地
- 事件总线或异步投递
- 代码路由与 `sys_api_resources` 的自动双向同步
- 根据 proto/http route 自动推导 API 资源目录

## 11. walk-routes 当前行为

当前实现里：

- `auth/user/common` 服务各自通过 `WalkHTTPRoutes(RestServer)` 返回本服务 HTTP 路由。
- `admin.GetWalkRoute()` 会聚合 `auth/user/common` 的远端结果，并追加 `admin` 自身的 HTTP 路由后统一去重。
- `admin.GetSelfWalkRoute()` 只返回 `admin` 自身路由。
- 这些接口主要用于观察和排查，不参与 Casbin 主数据同步。
