# API 归属与业务域判定规范

## 1. 目的

本文用于统一接口目录、权限映射、网关路由和服务注册里的归属口径，避免同一个 `path + method` 被多个业务域并列认领，导致授权、审计和故障定位语义不一致。

## 2. 核心结论

一个网关可见 API 由 `path + method` 唯一定位，并且在 API 目录中只能有一个主业务域 owner。

也就是说：

- `path + method` 只能映射到一个 `domain_code`
- `path + method` 只能映射到一个 `resource_group + action`
- `path + method` 可以被多个业务域消费或依赖，但不能被多个业务域并列拥有
- `service_code` 表示技术实现和路由落点，不表示业务授权边界

目标态接口级授权关系固定为：

```text
path + method -> domain_code + resource_group + action
user + domain_code + scope_id -> role
role + domain_code + resource_group + action -> allow
```

因此，`domain_code` 是权限判断的业务隔离边界，`service_code` 只是路由、审计、投影来源和故障定位元数据。

## 3. 字段语义

### 3.1 `domain_code`

`domain_code` 表示 API 的主业务权限域。

判定规则：

- 谁定义该 API 的业务规则，API 就归谁
- 谁决定该 API 的授权语义，API 就归谁
- 谁维护该 API 对应的角色、资源组、操作和用户绑定真相，API 就归谁

`domain_code` 不表示：

- 哪个服务实现了接口
- 哪些业务会调用接口
- 哪些页面会展示接口返回的数据

### 3.2 `service_code`

`service_code` 表示 API 的技术服务归属。

它用于：

- 网关路由落点
- 日志审计来源
- 故障定位
- 权限投影来源服务

它不用于：

- 主要业务授权边界
- 判断用户在某个业务域下是否有角色
- 代替 `domain_code + resource_group + action`

### 3.3 `sys_business_domain.owner_service`

`owner_service` 表示某个业务域的主责治理服务或默认承载服务。

它不表示该业务域所有 API 都必须由同一个服务实现。一个业务域可以由多个技术服务共同承载，例如 `hospital-core` 和 `hospital-billing` 都可以承载 `hospital` 域的 API。

### 3.4 `sys_service_registry.domain_code`

当前 `sys_service_registry.domain_code` 表示服务的默认或主业务域，用于平台治理和网关归属推导。

如果某个技术服务承载多个业务域的 API，不能用服务注册表里的单个 `domain_code` 反推所有接口的业务域。此时应以 `sys_api_resources.domain_code` 作为接口级业务域来源。

## 4. 为什么不允许多业务域并列 owner

同一个 `path + method` 同时属于多个业务域会带来以下问题：

- 授权不可判定：同一个请求需要按哪个 `domain_code + scope_id` 查角色不明确
- 审计不可归责：同一个接口的业务规则 owner 不明确
- 投影不可合并：多个来源同时写同一 API 映射时容易互相覆盖
- 演进不可控：任一业务域改规则都可能破坏其他业务域

所以平台规范是：允许跨域调用、跨域协作、跨域聚合，但 API 目录必须保留唯一主业务域 owner。

## 5. 常见场景判定

### 5.1 被多个业务使用

如果 API 只是被多个业务调用，它仍然只归业务规则 owner。

示例：

```text
GET /user-api/v1/users/{id}
```

医院、学校、商店都可能读取用户基础信息，但用户基础身份规则由 `user`/平台身份能力维护。它不因此同时属于 `hospital`、`school`、`store`。

### 5.2 聚合查询接口

聚合接口必须选择编排 owner。

判定顺序：

1. 如果聚合结果服务于某个明确业务流程，归该业务域。
2. 如果聚合结果是平台控制面视图，归 `platform`。
3. 如果只是技术聚合但没有稳定业务语义，应优先拆分为多个业务 API，由前端或 BFF 编排。

### 5.3 跨域写操作

跨域写操作不能因为同时改了多个域的数据就设置多个 owner。

正确做法是：

- 选择发起该业务用例的主业务域作为 `domain_code`
- 其他业务域通过内部 RPC、事件或事务编排参与
- 各业务服务仍负责自己的实例级数据校验和副作用保护

如果无法选出主业务域，通常说明接口职责过大，应拆成更小的业务命令或增加明确的流程域。

### 5.4 平台/基础设施接口

登录、刷新 token、系统菜单、平台审计、服务注册、业务域注册、投影状态等平台能力统一归 `platform` 或对应的平台子域。

这类接口可以服务所有业务，但不应归到某个具体业务域。

### 5.5 迁移期兼容接口

迁移期可能存在旧路径、新路径或旧权限模型兼容。

兼容规则：

- 仍然必须指定唯一 `domain_code`
- 可以保留旧 `service/scope` 运行时 key 作为兼容投影
- 文档和 seed 必须标注迁移状态，不能把兼容态解释为多 owner

## 6. 归属判定流程

新增或调整 API 时按以下顺序判断：

1. 确认 `path + method` 是否已存在；存在则不得新增第二条并列归属记录。
2. 找出该 API 的业务规则 owner，填写唯一 `domain_code`。
3. 找出实际提供 HTTP/gRPC 实现的服务，填写唯一 `service_code`。
4. 定义接口级授权资源，填写 `resource_group + action`。
5. 如果 API 涉及多个业务域，记录协作方，但不增加第二个 owner。
6. 如果无法选出唯一 owner，先拆 API 或补充一个明确的流程/平台域。

## 7. 与当前实现的关系

当前仓库处于迁移阶段，需要同时理解目标态和运行态：

- 目标态以 `domain_code + scope_id + role + resource_group + action` 作为授权模型
- 当前运行时仍保留 `service + path + method + scope_id` 兼容 key
- `sys_api_resources` 已包含 `service_code` 和 `domain_code`
- `auth.casbin_rules` 是授权投影，不是 API 主数据
- `walk-routes` 是路由发现和对账结果，不是 API 权限目录真相

当前判断接口业务域时，以 `sys_api_resources.domain_code` 为准；判断技术路由落点时，以 `sys_api_resources.service_code` 或 `sys_service_registry.http_prefix -> service_code` 为准。

## 8. 一句话规则

同一个 `path + method` 只能有一个主业务域 owner；其他业务域可以调用、依赖或参与编排，但不能与 owner 并列归属。
