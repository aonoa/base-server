# 权限设计说明

## 1. 背景与目标

当前仓库已拆分为微服务架构，但现有权限模型仍然偏向“单后台、单组织、单身份”模式，主要问题包括：

- `auth` 服务同时承担鉴权执行和业务角色/资源管理职责，边界过重。
- `user` 服务存在单一 `role_id` 设计，不支持同一用户在不同业务、不同组织下拥有不同身份。
- 不同业务服务的组织结构差异很大，无法用一套统一业务模型抽象。

本次权限重构的目标是：

- 支持同一用户在不同业务中拥有不同身份。
- 支持不同业务拥有不同的组织层级与成员关系。
- 让业务服务自治维护自己的权限主数据。
- 让 `auth` 服务专注于认证、统一授权和策略执行。
- 避免不同业务中同名角色、同名资源产生权限串扰。

## 2. 总体设计原则

权限设计遵循以下原则：

1. `user` 服务只负责平台身份，不负责业务角色。
2. `admin` 服务只负责平台治理信息，不负责业务权限真相。
3. 各业务服务负责维护自己的组织结构、成员关系、角色、资源、权限及其绑定关系。
4. `auth` 服务不再作为业务权限主数据中心，只保存标准化后的授权投影，用于统一鉴权。
5. 接口级权限必须以 `domain_code` 作为业务隔离边界。
6. `service_code` 只表示技术服务归属与路由落点，不作为主要业务授权边界。
7. API 权限统一抽象为 `API -> domain_code + resource_group + action`。
8. 角色授权统一抽象为 `role -> domain_code + resource_group + action`。
9. 所有用户角色绑定都必须带权限生效范围 `scope_id`。
10. 网关和 `auth` 只负责接口级统一鉴权，实例级数据权限仍由业务服务自行校验。
11. 业务服务修改权限模型后，应将标准化授权数据同步到 `auth`。
12. 同一个 `path + method` 只能有一个主业务域 owner；跨域调用、跨域聚合和跨域协作不等于多 owner。

接口归属判定的完整规则见 [API 归属与业务域判定规范](./api-ownership.md)。

## 3. 服务职责边界

### 3.1 `user` 服务

`user` 服务只维护平台用户身份相关信息：

- 用户基础信息
- 登录凭证
- 账号状态
- 平台统一 `user_id`

`user` 服务不应承担：

- 业务角色定义
- 组织结构管理
- 用户在业务中的身份绑定

现有 `user.role_id` 属于单一全局角色模型，应逐步废弃。

### 3.2 `admin` 服务

`admin` 服务只负责平台治理面：

- 业务域注册
- 服务注册
- API 目录注册与归属治理
- 权限投影源状态治理
- 平台级菜单、日志、审计
- 平台级配置

`admin` 不应承担：

- 业务组织结构真相
- 业务角色真相
- 业务用户角色绑定真相
- 业务资源实例权限真相

这里需要特别区分三类信息：

- `admin` 可以统一管理 `sys_api_resources` 里的 API 目录，以及 `path/method -> domain_code/resource_group/action` 的接口级权限映射
- `admin` 可以管理 API 的 `service_code` 技术归属，用于网关路由、审计和故障定位
- 但某个业务服务内部“谁拥有什么角色、角色拥有哪些资源、用户在哪个 scope 下绑定什么角色”仍然由业务服务自己持有

另外，`admin.sys_service_registry` 不只是展示用元数据：

- `gateway` 会优先基于服务注册表里的 `http_prefix -> service_code` 映射推导请求所属业务服务
- 后续网关也可以基于 API 目录进一步推导 `domain_code/resource_group/action`
- 未命中注册表时，才回退到当前仓库内置的老前缀规则

### 3.3 业务服务

每个业务服务是该业务权限主数据的唯一事实来源，负责维护：

- 组织结构
- 成员关系
- 角色定义
- 资源定义
- 权限定义
- 角色与权限绑定
- 用户在本业务中的身份绑定

例如：

- 医院服务维护医院、科室、医生、病人、医保监管等结构与权限。
- 商店服务维护门店、老板、店员、订单、库存等结构与权限。
- 学校服务维护校区、院系、教师、学生、课程等结构与权限。

不同业务的结构不要求同构。

### 3.4 `auth` 服务

`auth` 服务只负责：

- 认证
- token 签发与校验
- 统一授权判定
- 授权数据汇总与策略执行
- Casbin policy/grouping policy 装载

`auth` 不负责：

- 业务组织树真相管理
- 业务成员关系真相管理
- 业务角色语义管理

## 4. 核心概念模型

建议保留以下 8 个核心概念：

### 4.1 `domain_code`

业务权限域，用于表达业务边界和权限隔离边界。

示例：

- `platform`
- `hospital`
- `store`
- `school`

`domain_code` 是权限模型里的主命名空间。接口级鉴权、角色授权和用户角色绑定都应带 `domain_code`。

对 API 来说，`domain_code` 表示该 `path + method` 的主业务权限域，并且必须唯一。一个 API 可以被多个业务域消费或参与编排，但 API 目录中不能把多个业务域并列记录为 owner。

### 4.2 `service_code`

技术服务编码，用于表达 API 的服务归属、路由落点、审计来源和投影来源。

示例：

- `admin`
- `auth`
- `user`
- `hospital-core`
- `hospital-billing`

`service_code` 不应作为主要业务授权边界。一个业务域可以由多个服务承载，一个服务也可能承载多个业务域的 API。

如果某个服务承载多个业务域的 API，接口级业务域应以 `sys_api_resources.domain_code` 为准，不能用服务注册表里的单个默认 `domain_code` 反推所有接口归属。

### 4.3 `scope_id`

权限生效范围，用于表达组织、租户、门店、部门、全局等边界。

示例：

- `global`
- `hospital:A`
- `hospital:A:cardiology`
- `store:001`

`scope_id` 不要求所有业务都有相同层级深度：

- 简单业务可以只有一个固定全局 scope。
- 复杂业务可以按机构、部门、科室逐级细化。

### 4.4 `role`

职责身份，表示“在某个业务域、某个范围内，你是什么角色”。

示例：

- `doctor`
- `patient`
- `hospital_admin`
- `store_owner`
- `clerk`

### 4.5 `resource_group`

接口级资源组，表示一组 API 对应的可授权资源集合。

示例：

- `patient_record`
- `appointment`
- `inventory`
- `order`

`resource_group` 是第一层接口鉴权的核心粒度。API 不直接绑定角色，API 先映射到 `domain_code + resource_group + action`，角色再被授予对应资源组上的操作权限。

### 4.6 `action`

对资源组执行的操作。

示例：

- `read`
- `create`
- `update`
- `delete`
- `approve`
- `refund`
- `export`

当前实现可以先把 HTTP method 当作兼容 action，例如 `GET/POST/PUT/DELETE`。目标态应支持业务语义 action，因为多个业务操作可能都是 `POST`，但权限含义完全不同。

### 4.7 `api`

网关可见的接口入口，由 `path + method` 唯一定位。

接口级权限映射为：

- `path + method -> domain_code + resource_group + action`

`domain_code + resource_group + action` 用于鉴权。`service_code` 可以作为 API 目录的技术归属元数据保存，但不进入授权策略匹配。

约束：

- 同一个 `path + method` 只能有一个 `domain_code` owner
- 同一个 `path + method` 只能有一组主授权语义 `resource_group + action`
- 多业务使用该接口时，其他业务是消费者或协作者，不是并列 owner
- 聚合接口、平台接口和迁移期兼容接口也必须指定唯一 owner

### 4.8 `binding`

把用户、业务域、范围、角色绑定起来，表示“某用户在某业务域的某范围下拥有什么角色”。

授权本质可以归纳为三段关系：

- `api(path + method) -> domain_code + resource_group + action`
- `user + domain_code + scope_id -> role`
- `role + domain_code + resource_group + action -> allow`

## 5. 组织与角色的关系

组织关系不建议直接建模为角色。

推荐区分：

- `scope` 表示组织或权限范围
- `role` 表示职责身份

例如：

- `hospital:A` 是 `scope`
- `doctor` 是 `role`

正确表达方式为：

- 用户 `u1` 在 `hospital:A` 下拥有 `doctor` 角色

不建议将组织直接编码进角色，例如：

- `hospital_A_doctor`
- `cardiology_doctor`

原因：

- 角色数量会爆炸
- 组织变更成本高
- 难以表达同一角色在不同组织下的复用
- 不利于实例级数据范围控制

## 6. 业务域、scope 与 Casbin domain

需要明确区分三类概念：

- `domain_code`：业务权限域，例如 `hospital`、`store`、`platform`
- `scope_id`：权限生效范围，例如 `global`、`hospital:A`、`store:001`
- Casbin `domain`：Casbin 模型里的隔离参数，可由工程实现选择如何编码

推荐把运行时授权上下文理解为：

- `domain_code` 负责业务隔离
- `scope_id` 负责组织、租户、门店、部门等范围隔离
- `resource_group + action` 负责表达具体接口权限

如果使用 Casbin RBAC with domains，可以将 Casbin domain 编码为稳定的授权范围键，例如：

- `hospital:global`
- `hospital:A`
- `hospital:A:cardiology`
- `store:001`

但文档中的 `domain_code` 不等同于 Casbin domain。`domain_code` 是业务域，Casbin domain 更接近“业务域下的 scope”。

## 7. scope 的层级与继承

如果 scope 之间完全孤立，则会带来大量重复授权工作。

例如：

- 一个账号需要管理所有医院
- 一个医保账号需要查看某医院所有科室数据

这类场景不应要求把用户逐个绑定到所有下级 scope。

正确做法是支持 scope 层级关系：

- `hospital:global` 包含 `hospital:A`
- `hospital:global` 包含 `hospital:B`
- `hospital:A` 包含 `hospital:A:cardiology`
- `hospital:A` 包含 `hospital:A:surgery`

这样可以实现：

- 平台管理员绑定到 `hospital:global`
- 医院管理员绑定到 `hospital:A`
- 科室主任绑定到 `hospital:A:cardiology`

然后通过上级 scope 覆盖下级 scope。

需要注意：

- Casbin 支持 `RBAC with domains`
- Casbin 支持角色继承
- 但 scope 自身的父子继承关系通常仍需要业务侧或 `auth` 侧自行建模

因此建议：

- scope 树的真相源由业务服务维护
- 业务服务向 `auth` 投影标准化 scope 关系或当前用户的有效 scope 祖先集合
- `auth` 不直接维护业务组织真相，只使用投影后的 scope 关系参与鉴权
- 鉴权前先求当前 scope 的有效祖先集合，再结合 Casbin 做角色与权限判定

## 8. 权限命名空间设计

为了避免不同业务中同名角色、同名资源串权限，必须引入命名空间。

推荐规则：

- 所有接口级授权数据必须带 `domain_code`
- API 目录可以带 `service_code`，用于技术归属和路由，但它不参与授权策略匹配
- 所有用户角色绑定必须带 `domain_code + scope_id`
- `resource_group` 和 `permission` 建议显式带业务域前缀或由结构化字段保证隔离

示例：

- 角色：`hospital.doctor`
- 资源组：`hospital.patient_record`
- 权限：`hospital.patient_record.read`

在工程实现中，可以使用结构化字段保存：

- `domain_code`
- `role_code`
- `resource_group`
- `action`

Casbin 中再编码成稳定字符串，不建议只依赖字符串拼接作为唯一建模手段。

## 9. 标准化授权投影

虽然权限主数据由业务服务维护，但同步到 `auth` 时，应压缩成统一格式。

推荐同步分成三类动作：

- 启动或恢复时发送完整快照 `snapshot`
- 日常变更发送增量 `delta`
- 当 delta 失败或投影脏化时，再回退到完整快照

建议同步以下标准化对象：

### 9.1 角色定义

- `domain_code`
- `role_code`
- `role_name`

### 9.2 资源组与操作定义

- `domain_code`
- `resource_group`
- `action`

### 9.3 角色权限绑定

- `domain_code`
- `role_code`
- `resource_group`
- `action`

### 9.4 用户角色绑定

- `user_id`
- `domain_code`
- `scope_id`
- `role_code`

### 9.5 API 权限映射

- `domain_code`
- `path`
- `method`
- `resource_group`
- `action`

其中最关键的是：

- `domain_code` 用于业务隔离
- `scope_id` 用于授权范围隔离
- `resource_group + action` 用于表达接口级权限
- `service_code` 仅作为 API 技术归属元数据，不进入授权策略匹配

缺少 `domain_code` 或 `scope_id` 会退化为全局静态角色模型。缺少 `resource_group + action` 会退化为只按业务域放行，粒度过粗。

### 9.6 接口级鉴权关系

目标态接口级鉴权关系为：

- `API(path + method) -> domain_code + resource_group + action`
- `user + domain_code + scope_id -> role`
- `role + domain_code + resource_group + action -> allow`

网关只负责把请求上下文交给 `auth`，不直接判断业务实例细节。

## 10. 全局权限设计

允许存在“全局权限”，但应仅限于平台级通用能力，不应侵入具体业务资源。

适合平台公共化的权限包括：

- `platform.user.read`
- `platform.user.manage`
- `platform.auth.manage`
- `platform.audit.read`
- `platform.system.config`

不适合抽成公共权限的内容包括：

- `hospital.patient_record.read`
- `store.order.refund`
- `school.grade.submit`

原则如下：

- 平台公共权限归 `platform` 业务域
- 业务权限归各自的 `domain_code` 命名空间
- 平台管理权与业务数据操作权应明确区分
- 超级全局权限必须极少、可审计

## 11. 鉴权分层

鉴权建议分成两层：

### 11.1 接口级鉴权

由网关和 `auth` 负责，判断：

- 当前请求属于哪个 `domain_code`
- 当前请求命中哪个 `resource_group + action`
- 当前请求处于哪个 `scope_id`
- 当前用户在该 `scope_id` 下拥有哪些 `role`
- 当前角色是否拥有该 `domain_code + resource_group + action`

### 11.2 实例级数据鉴权

由业务服务负责，判断：

- 该用户能否操作某一条具体数据
- 该数据是否属于当前组织/科室/门店
- 该用户是否具备更细粒度的业务规则条件

不应尝试把所有实例级业务规则都塞进 `auth`。

## 12. 请求上下文要求

现有仅基于 `user_id + path + method` 的授权请求不够支撑多业务、多范围身份模型。

建议后续授权请求最少包含：

- `user_id`
- `scope_id`
- `path`
- `method`

`auth` 应基于 API 目录投影解析：

- `path + method -> domain_code + resource_group + action`

如果网关已经预解析出 API 元数据，也可以显式传入：

- `domain_code`
- `resource_group`
- `action`

只有这样，`auth` 才能判断：

- 同一用户在不同业务下是否拥有不同权限
- 同一用户在同一业务的不同 scope 下是否拥有不同角色
- 同一业务域内不同资源组、不同操作是否具备不同授权

## 13. 与当前仓库的重构方向

结合当前仓库，后续重构应朝以下方向推进：

1. 逐步废弃 `user.role_id` 单角色模型。
2. 将 `auth` 中现有 `role/resource/api` 从“主数据管理”收敛为“授权投影与策略执行”。
3. `admin` 收敛为平台治理面，不再承担业务权限真相中心。
4. 各业务服务拥有自己的角色、资源、权限与组织模型。
5. `auth` 接收业务服务同步的标准化授权数据。
6. `CheckAuthorization` 从当前 `service + path + method + scope_id` 逐步升级为 `domain_code + path + method + scope_id` 或 `domain_code + resource_group + action + scope_id`。
7. `sys_api_resources` 增加或规范化 `action` 字段，避免只用 HTTP method 表达业务操作。
8. 网关继续作为统一接口级鉴权入口。
9. 业务服务保留实例级权限判断权。

### 13.1 当前代码阶段

截至当前代码：

- 运行中的权限投影链路仍主要是 `admin -> auth`
- 但 `auth` 的投影 RPC 已经保留 `source_service`
- 代码中已新增通用投影发送层，后续业务服务可以直接复用同一套 sender，而不必继续复制 `admin` 的拼装逻辑

因此当前阶段属于：

- 架构目标已切换到“业务服务持有权限主数据”
- 仓库里的现有实现还处于“先把投影注入链路做通用化，再逐步接入更多业务服务”
- 当前运行时鉴权仍主要使用 `service + path + method + scope_id`，`domain_code` 还没有进入 `auth` 的运行时策略键

## 14. 本次设计确认结论

本次对话已确认以下结论：

1. 当前微服务架构下，角色和资源不应统一由 `auth` 作为业务主数据管理。
2. `admin` 只负责平台治理，不负责业务权限真相。
3. 业务服务负责维护自己的组织结构、成员关系、角色、资源、权限及绑定关系。
4. 权限变更后，业务服务将授权事实标准化后同步到 `auth`。
5. 标准化授权模型统一抽象为：`domain_code + scope_id + role + resource_group + action`。
6. `scope` 表示权限范围，既可以简单为全局，也可以有多层组织结构。
7. 不要求所有业务的 scope 结构一致。
8. 组织不是角色，组织关系应用 `scope` 表达。
9. Casbin 可用于 domain 隔离与角色权限判定，但 scope 继承关系通常需额外建模。
10. 平台公共权限可以存在，但只应覆盖真正的平台级能力。
11. 鉴权必须区分接口级授权和实例级数据授权。
12. API 先映射到 `domain_code + resource_group + action`，角色再被授予对应资源组操作权限。

## 15. 后续设计工作

基于本说明，后续建议继续输出并实现以下内容：

1. 权限领域对象与表结构设计
2. 业务服务向 `auth` 的同步接口或消息模型设计
3. `auth` 内部授权投影表设计
4. Casbin model 设计
5. API 目录 `action` 字段与 `path/method -> domain_code/resource_group/action` 映射设计
6. 网关 `CheckAuthorization` 入参与上下文改造
7. 从现有 `user.role_id` 模型迁移到 `user + domain_code + scope_id + role` 模型的步骤
8. 从当前 `admin -> auth` 单源投影过渡到多业务服务并行投影的步骤
