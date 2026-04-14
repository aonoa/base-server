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
2. 各业务服务负责维护自己的组织结构、成员关系、角色、资源、权限及其绑定关系。
3. `auth` 服务不再作为业务权限主数据中心，只保存标准化后的授权投影，用于统一鉴权。
4. 所有授权数据都必须带业务命名空间 `service`。
5. 所有用户角色绑定都必须带权限生效范围 `scope`。
6. 网关和 `auth` 只负责接口级统一鉴权，实例级数据权限仍由业务服务自行校验。
7. 业务服务修改权限模型后，应将标准化授权数据同步到 `auth`。

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

### 3.2 业务服务

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

### 3.3 `auth` 服务

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

建议保留以下 6 个核心概念：

### 4.1 `service`

业务命名空间，用于区分不同业务。

示例：

- `hospital`
- `store`
- `school`
- `platform`

### 4.2 `scope`

权限生效范围，用于表达组织、租户、门店、部门、全局等边界。

示例：

- `hospital:global`
- `hospital:A`
- `hospital:A:cardiology`
- `store:global`

`scope` 不要求所有业务都有相同层级深度：

- 简单业务可以只有一个固定全局 scope。
- 复杂业务可以按机构、部门、科室逐级细化。

### 4.3 `role`

职责身份，表示“在某个范围内，你是什么角色”。

示例：

- `doctor`
- `patient`
- `hospital_admin`
- `store_owner`
- `clerk`

### 4.4 `resource`

被保护对象，表示需要被访问或操作的业务资源。

示例：

- `patient_record`
- `appointment`
- `inventory`
- `order`

### 4.5 `action`

对资源执行的操作。

示例：

- `read`
- `create`
- `update`
- `delete`
- `approve`

### 4.6 `binding`

把用户、范围、角色绑定起来，表示“某用户在某范围下拥有什么角色”。

授权本质可以归纳为两段关系：

- `user + scope + role`
- `role + resource + action`

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

## 6. 业务隔离与 Casbin domain

不同业务以及同一业务下的不同授权范围，都可以利用 Casbin 的 `domain` 概念进行隔离。

但这里的 `domain` 不应只表示“业务名”，更适合表示“权限作用域”。

推荐理解为：

- `service` 用于业务命名空间
- `domain/scope` 用于权限生效范围

推荐编码示例：

- `store:global`
- `hospital:global`
- `hospital:A`
- `hospital:A:cardiology`

因此：

- 单业务单实例场景，可以使用固定 scope，例如 `store:global`
- 多机构场景，可以使用细粒度 scope，例如 `hospital:A`

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

- scope 树由业务服务或 `auth` 维护标准化关系
- 鉴权前先求当前 scope 的有效祖先集合
- 再结合 Casbin 做角色与权限判定

## 8. 权限命名空间设计

为了避免不同业务中同名角色、同名资源串权限，必须引入命名空间。

推荐规则：

- 所有授权数据必须带 `service`
- `scope` 编码中也包含 `service`
- `resource` 和 `permission` 建议显式带业务前缀

示例：

- 角色：`hospital.doctor`
- 资源：`hospital.patient_record`
- 权限：`hospital.patient_record.read`

在工程实现中，可以使用结构化字段保存：

- `service`
- `role_code`
- `resource_code`
- `action`

Casbin 中再编码成稳定字符串，不建议只依赖字符串拼接作为唯一建模手段。

## 9. 标准化授权投影

虽然权限主数据由业务服务维护，但同步到 `auth` 时，应压缩成统一格式。

建议同步以下标准化对象：

### 9.1 角色定义

- `service`
- `role_code`
- `role_name`

### 9.2 权限定义

- `service`
- `resource`
- `action`

### 9.3 角色权限绑定

- `service`
- `role_code`
- `resource`
- `action`

### 9.4 用户角色绑定

- `user_id`
- `service`
- `scope_id`
- `role_code`

### 9.5 API 权限映射

- `service`
- `path`
- `method`
- `resource`
- `action`

其中最关键的是：

- `service` 用于业务隔离
- `scope_id` 用于授权范围隔离

缺少任一维度都会退化为全局静态角色模型。

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

- 平台公共权限归 `platform` 命名空间
- 业务权限归各业务服务命名空间
- 平台管理权与业务数据操作权应明确区分
- 超级全局权限必须极少、可审计

## 11. 鉴权分层

鉴权建议分成两层：

### 11.1 接口级鉴权

由网关和 `auth` 负责，判断：

- 当前请求属于哪个 `service`
- 当前请求处于哪个 `scope`
- 当前用户在该 `scope` 下拥有哪些 `role`
- 当前 API 需要什么 `resource + action`

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
- `service`
- `scope_id`
- `path`
- `method`

只有这样，`auth` 才能判断：

- 同一用户在不同业务下是否拥有不同权限
- 同一用户在同一业务的不同 scope 下是否拥有不同角色

## 13. 与当前仓库的重构方向

结合当前仓库，后续重构应朝以下方向推进：

1. 逐步废弃 `user.role_id` 单角色模型。
2. 将 `auth` 中现有 `role/resource/api` 从“主数据管理”收敛为“授权投影与策略执行”。
3. 各业务服务拥有自己的角色、资源、权限与组织模型。
4. `auth` 接收业务服务同步的标准化授权数据。
5. `CheckAuthorization` 增加 `service` 和 `scope_id`。
6. 网关继续作为统一接口级鉴权入口。
7. 业务服务保留实例级权限判断权。

## 14. 本次设计确认结论

本次对话已确认以下结论：

1. 当前微服务架构下，角色和资源不应统一由 `auth` 作为业务主数据管理。
2. 业务服务负责维护自己的组织结构、成员关系、角色、资源、权限及绑定关系。
3. 权限变更后，业务服务将授权事实标准化后同步到 `auth`。
4. 标准化授权模型统一抽象为：`service + scope + role + resource + action`。
5. `scope` 表示权限范围，既可以简单为全局，也可以有多层组织结构。
6. 不要求所有业务的 scope 结构一致。
7. 组织不是角色，组织关系应用 `scope` 表达。
8. Casbin 可用于 domain 隔离与角色权限判定，但 scope 继承关系通常需额外建模。
9. 平台公共权限可以存在，但只应覆盖真正的平台级能力。
10. 鉴权必须区分接口级授权和实例级数据授权。

## 15. 后续设计工作

基于本说明，后续建议继续输出并实现以下内容：

1. 权限领域对象与表结构设计
2. 业务服务向 `auth` 的同步接口或消息模型设计
3. `auth` 内部授权投影表设计
4. Casbin model 设计
5. 网关 `CheckAuthorization` 入参与上下文改造
6. 从现有 `user.role_id` 模型迁移到 `user + scope + role` 模型的步骤
