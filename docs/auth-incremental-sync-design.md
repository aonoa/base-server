# auth 增量同步接口设计（v1）

## 1. 背景

当前权限投影链路是：

- `auth` 启动时在 `NewAuthUsecase()` 里直接执行 `syncAuthPolicy()`。
- `syncAuthPolicy()` 会清空当前 Casbin 内存策略，再通过 `admin` 的内部只读 RPC 拉取角色、API 目录、用户角色绑定，然后重新 `SavePolicy()`。
- `admin` 侧的角色、API、资源、用户角色绑定写操作，当前统一通过 `auth.ReLoadPolicy()` 触发一次全量重建。

这条链路已经能工作，但有 4 个明显问题：

- `auth` 启动依赖 `admin` 可用，服务顺序和白名单配置都比较脆。
- 任何小变更都会触发一次全量重建，写放大明显。
- `auth` 和 `admin` 之间的耦合点太多，启动阶段还会碰到“没有用户 JWT”的内部调用问题。
- 当前的恢复策略只有“整库重建”，没有“按对象增量修复”能力。

## 2. v1 目标

这版先不做多业务权限中心，只解决当前 `admin -> auth` 这一条链路。

目标是：

- `auth` 启动时不再主动回拉 `admin`。
- `auth` 只负责加载自己库里已有的 `casbin_rules` 并提供鉴权。
- `admin` 作为权限主数据 owner，在服务启动后主动向 `auth` 注册一次完整快照。
- `admin` 的日常写操作改为向 `auth` 发送精确 delta。
- delta 失败时仍然可以回退到一次完整快照，保留可恢复性。

## 3. 职责边界

- `admin` 仍然是角色、资源、API 资源目录、用户角色绑定的唯一事实来源。
- `auth` 不再负责“发现真相”，只维护授权投影和 Casbin 执行状态。
- `auth` 库中的 `casbin_rules` 仍然是投影存储，不是主数据。

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
- 在迁移期内，`ReLoadPolicy()` 可以先保留，作为最后兜底。

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

语义约束：

- `source_service` 在 v1 固定为 `admin`，先预留字段，后面再扩展多业务。
- `revision` 由发送方单调递增，建议先用毫秒时间戳或数据库变更版本号。
- 这是“全量替换该 source 的授权快照”，不是 merge。
- 请求必须幂等；相同 `source_service + revision` 重放不能产生副作用。

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

语义约束：

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
- 如果 `path` 或 `resources_group` 变化，必须带完整 `before/after`，让 `auth` 能正确更新 `g2`。

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

语义约束：

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

## 7. 建议的服务内执行流程

### 7.1 auth 启动

建议把现在的：

- `NewAuthUsecase()` 里直接 `syncAuthPolicy()`

改成：

- `NewEnforcer()` 后只执行 `LoadPolicy()`
- 不在启动阶段访问 `admin`

这样 `auth` 单独重启时，不会再依赖 `admin` 的白名单和可用性。

### 7.2 admin 启动注册

`admin` 启动完成后，异步执行：

1. 从本地数据库读出完整角色、API 资源目录、用户角色绑定。
2. 组装 `RegisterPermissionSnapshotRequest`。
3. 调用 `auth.RegisterPermissionSnapshot(...)`。
4. 失败则按固定退避重试，直到成功。

这一步是“启动注册”，用来替代今天 `auth` 自己回拉 `admin` 的启动重建。

### 7.3 admin 日常写操作

建议把下面这些写路径逐步从 `ReLoadPolicy()` 切到 delta：

- `AddRole/UpdateRole/DelRole` -> `ApplyRoleDelta`
- `AddApi/UpdateApi/DelApi` -> `ApplyApiDelta`
- `UpsertUserRoleBinding/DeleteUserRoleBinding` -> `ApplyUserRoleBindingDelta`

资源变更建议这样处理：

- `AddResource/UpdateResource/DelResource` 不单独设计 `ResourceDelta`
- 由 `admin` 找出受影响 role
- 对每个受影响 role 发送一次 `ApplyRoleDelta`

原因很简单：

- `auth` 真正关心的是“角色最终拥有哪组资源策略”
- 资源本身只是角色快照的一部分
- 直接按 role 发 delta，接口更稳定，也更容易复用现有 `role -> policy` 代码

## 8. 失败回退策略

v1 先不要做复杂队列，先保守落地：

1. 优先发送 delta。
2. delta 失败时，立即尝试发送一次 `RegisterPermissionSnapshot`。
3. 如果快照也失败，记录 error log，并保留现有 `casbin_rules` 不动。
4. 在迁移期内继续保留 `ReLoadPolicy()`，作为最后兜底入口。

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

## 10. 迁移顺序

推荐按下面顺序切，风险最小：

1. 在 `auth.proto` 增加快照注册和 3 个 delta RPC。
2. `auth` 侧先实现接口，但暂时不改现有 `ReLoadPolicy()`。
3. `admin` 启动时补一次 `RegisterPermissionSnapshot`。
4. `auth` 启动改成仅 `LoadPolicy()`，移除启动阶段对 `admin` 的主动回拉。
5. 逐条把 `admin` 写接口从 `ReLoadPolicy()` 改成 delta，同步保留失败回退快照。
6. 跑稳定后，再考虑移除旧的启动重建依赖和多余白名单。

## 11. 本版刻意不做的事

这版先不处理下面几件事：

- 多业务服务同时向 `auth` 写入权限主数据
- 跨服务命名空间隔离
- 事件总线或异步投递
- `admin.GetWalkRoute()` 这一类路由聚合接口的注册制改造

如果后面要把“路由发现”也改成注册制，可以沿用同一套思路：

- `source_service`
- `revision`
- 启动后全量 replace
- 日常按对象 delta

但它和当前最急的 `admin -> auth` 权限投影问题不是一条关键路径，可以放到下一轮。
