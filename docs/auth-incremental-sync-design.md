# auth 增量同步实现说明（当前代码）

## 1. 背景

当前权限投影链路是：

- `auth` 启动时在 `NewAuthUsecase()` 里只执行 `LoadPolicy()`，直接从自己的 `casbin_rules` 恢复策略。
- `admin` 启动后通过 `startup_sync` 向 `auth.RegisterPermissionSnapshot(...)` 注册一次完整快照，失败时按退避重试。
- `admin` 日常写操作按对象发送 delta；delta 失败时回退到一次完整快照注册。

当前代码里的真相边界是：

- `admin` 库里的角色、API 资源目录、用户角色绑定，是权限主数据。
- `auth` 库里的 `casbin_rules` 是授权投影，不是主数据。
- `walk-routes` 只用于路由发现/对账，不会自动回写或替代 `sys_api_resources`。
- 修改 proto 或 HTTP 路由，不会自动推导出对应的 Casbin API 权限变更；只有修改 `admin` 的 API 资源目录，才会触发 `ApplyApiDelta`。

## 2. v1 目标

当前实现只覆盖 `admin -> auth` 这一条权限投影链路。

当前已落地的目标：

- `auth` 启动时不再主动回拉 `admin`。
- `admin` 启动成功后主动注册完整快照。
- `AddRole/UpdateRole/DelRole`、`AddApi/UpdateApi/DelApi`、`UpsertUserRoleBinding/DeleteUserRoleBinding` 已改为 delta。
- `AddResource/UpdateResource/DelResource` 当前仍直接走完整快照回放。
- delta 失败时会回退到完整快照，保留恢复能力。

## 3. 职责边界

- `admin` 仍然是角色、资源、API 资源目录、用户角色绑定的唯一事实来源。
- `auth` 不再负责“发现真相”，只维护授权投影和 Casbin 执行状态。
- `auth` 库中的 `casbin_rules` 仍然是投影存储，不是主数据。
- `/admin-api/v1/walk-routes` 返回的是路由发现结果，不是 Casbin API 权限目录。

## 4. 设计原则

### 4.1 启动注册替代启动回拉

- 不再让 `auth` 在启动阶段主动调用 `admin`。
- 改成 `admin` 启动完成后主动向 `auth` 注册完整权限快照。
- 这样 `auth` 单独重启时也能直接从自己的库里 `LoadPolicy()` 恢复服务能力。

### 4.2 日常同步只发足量 delta

- `auth` 不应在处理增量时再次回查 `admin`。
- `admin` 发给 `auth` 的 delta 必须带足够的 `before/after` 信息。
- 这样 `auth` 才能在本地完成 rename、remove、upsert，而不重新走全量拉取。

### 4.3 保留完整快照回退路径

- 增量同步失败时，不要直接让链路失效。
- `admin` 应该能回退到一次完整快照注册。
- 当前代码已不再依赖 `ReLoadPolicy()` 作为主链路。

## 5. 推荐 RPC 契约

这版建议只扩展 `api/protos/auth/service/v1/auth.proto`，由 `admin` 调 `auth`。

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

- `source_service` 在 v1 固定为 `admin`，先预留字段，后面再扩展多业务。
- `revision` 由发送方单调递增，建议先用毫秒时间戳或数据库变更版本号。
- 当前实现会直接 `ClearPolicy()`，然后重新应用整份快照并 `SavePolicy()`。
- 当前实现没有按 `source_service` 做分源 merge，而是重建整份本地 Casbin 投影。

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
}

message PolicyUserRoleBinding {
  string id = 1;
  string user_id = 2;
  int64 role_id = 3;
  string role_value = 4;
  string create_time = 5;
  string update_time = 6;
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

因此，增量接口不需要让 `auth` 再次查 `admin`，只要 `admin` 把 `before/after` 发全，`auth` 就能在本地原子改 Casbin 投影。

补充说明：

- API 接口级权限判断走 `p2 + g2`。
- `g2` 表示 `apiPath -> api:<resources_group>` 的分组映射。
- 前端编辑“API 资源列表”里的 `path` 或 `resources_group`，本质上是在改 `sys_api_resources` 主数据；这会通过 `ApplyApiDelta` 更新 `g2`。
- 仅修改 proto/http route，不会自动改 `sys_api_resources`，因此也不会自动改 Casbin。

## 7. 建议的服务内执行流程

### 7.1 auth 启动

当前代码已实现为：

- `NewEnforcer()` 创建 Enforcer，并开启 `AutoSave`
- `NewAuthUsecase()` 中执行 `LoadPolicy()`
- 启动阶段不访问 `admin`

### 7.2 admin 启动注册

当前代码通过 `startup_sync` transport 在 `admin` 启动后执行：

1. 从本地数据库读出完整角色、API 资源目录、用户角色绑定。
2. 组装 `RegisterPermissionSnapshotRequest`。
3. 调用 `auth.RegisterPermissionSnapshot(...)`。
4. 失败则按固定退避重试，直到成功。

成功后会持续等待；失败则按指数退避重试。

### 7.3 admin 日常写操作

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

- 多业务服务同时向 `auth` 写入权限主数据
- 跨服务命名空间隔离
- 事件总线或异步投递
- 代码路由与 `sys_api_resources` 的自动双向同步
- 根据 proto/http route 自动推导 API 资源目录

## 11. walk-routes 当前行为

当前实现里：

- `auth/user/common` 服务各自通过 `WalkHTTPRoutes(RestServer)` 返回本服务 HTTP 路由。
- `admin.GetWalkRoute()` 会聚合 `auth/user/common` 的远端结果，并追加 `admin` 自身的 HTTP 路由后统一去重。
- `admin.GetSelfWalkRoute()` 只返回 `admin` 自身路由。
- 这些接口主要用于观察和排查，不参与 Casbin 主数据同步。
