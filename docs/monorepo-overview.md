# Monorepo Overview

## 1. 适用范围

本文只描述 `base-server/monorepo` 这条微服务版本线。

运行形态：

- gateway：统一 HTTP 入口
- auth：认证和授权执行
- user：用户主数据
- admin：平台治理主数据
- common：上传、SSE、LLM 能力

## 2. 目录总览

```text
base-server/
├── api/
│   ├── protos/<service>/service/v1/*.proto
│   ├── gen/go/**                 # 生成代码
│   └── openapi/openapi.yaml
├── app/
│   ├── gateway/service/
│   ├── auth/service/
│   ├── user/service/
│   ├── admin/service/
│   └── common/service/
├── pkg/
│   ├── authx/                    # 权限投影发送与共享认证逻辑
│   ├── data/                     # 共享 Ent schema / generated client
│   ├── logx/
│   ├── contextx/
│   ├── transportx/
│   └── tools/
├── deploy/
│   ├── configs/<service>/config.yaml
│   ├── scripts/
│   └── sql/
├── docker-compose-env.yml
├── docker-compose.yml
└── Makefile
```

## 3. 服务地图

| 服务 | 入口 | 本地配置 | 主数据库 | API 前缀 | 主要职责 |
| --- | --- | --- | --- | --- | --- |
| gateway | `app/gateway/service/cmd/service/main.go` | `app/gateway/service/configs` | 无独立业务库 | 外部统一入口 | 路由转发、JWT、Casbin、限流、请求日志 |
| auth | `app/auth/service/cmd/service/main.go` | `app/auth/service/configs` | `auth` | `/auth-api/v1/*` | 登录、刷新 token、授权投影执行、Casbin |
| user | `app/user/service/cmd/service/main.go` | `app/user/service/configs` | `user` | `/user-api/v1/*` | 用户 CRUD、用户认证信息、密码 |
| admin | `app/admin/service/cmd/service/main.go` | `app/admin/service/configs` | `admin` | `/admin-api/v1/*` | 菜单、角色、资源、API 目录、组织、部门、日志、平台治理、投影源状态 |
| common | `app/common/service/cmd/service/main.go` | `app/common/service/configs` | `common` | `/common-api/v1/*` | 文件上传、Copilot SSE、站内信、通用能力 |

## 4. 共享层边界

### 4.1 API 契约

- 源文件：`api/protos/<service>/service/v1/*.proto`
- 生成物：`api/gen/go/**`
- OpenAPI：`api/openapi/openapi.yaml`

改 proto 后需要运行：

```bash
make api
```

### 4.2 配置结构

- 源文件：`app/*/service/internal/conf/conf.proto`
- 生成物：`app/*/service/internal/conf/*.pb.go`

改配置 proto 后需要运行：

```bash
make config
```

### 4.3 数据层

- 源文件：`pkg/data/schema/*.go`
- 生成物：`pkg/data/ent/**`
- 服务内 repo：`app/*/service/internal/data/**`

改 Ent schema 后需要运行：

```bash
make ent
```

表和数据库归属不看 `pkg/data/schema` 所在目录，而看各服务自己的迁移列表。详细规则见 [table-ownership.md](./table-ownership.md)。

### 4.4 依赖注入

- Wire 入口：`app/*/service/cmd/service/wire.go`
- 生成物：`app/*/service/cmd/service/wire_gen.go`

改 provider 后需要运行：

```bash
make wire
```

## 5. 权限与治理主线

当前微服务线的权限真相边界是：

- `user`：平台身份真相
- `admin`：平台治理主数据真相
  - `sys_api_resources`
  - `sys_resources`
  - `sys_role`
  - `sys_user_role_binding`
  - `sys_organization`
  - `sys_user_organization`
  - `sys_menu`
  - `sys_dept`
  - `sys_log`
  - `sys_service_registry`
  - `sys_projection_source_status`
- `auth`：Casbin 投影执行，不是业务权限主数据中心

当前组织上下文已经进入权限链路：

- 前端在加载并校验当前用户组织后，通过 `x-organization-id` 发送当前组织 ID。
- gateway 将 `x-organization-id` 作为 `organization_id` 传给 `auth.CheckAuthorization`。
- `admin` 将 `sys_user_role_binding.organization_id` 投影为 auth 的组织 domain。
- 因此用户在不同组织下可以有不同角色、菜单和接口权限。
- 默认组织是全员组织；默认组织下无显式绑定的用户会获得 `default` 角色回退。

需要结合阅读：

- [permission-design.md](./permission-design.md)
- [api-ownership.md](./api-ownership.md)
- [auth-incremental-sync-design.md](./auth-incremental-sync-design.md)

## 6. 常见改动路径

### 6.1 新增或修改 API

1. 改对应 `api/protos/<service>/service/v1/*.proto`
2. `make api`
3. 改 `app/<service>/service/internal/service/**`
4. 必要时改 `biz` / `data`
5. 如果前端也会调用，继续跑：

```bash
make frontend-api
```

### 6.2 新增或修改数据表

1. 改 `pkg/data/schema/*.go`
2. `make ent`
3. 改服务内 `internal/data/**`
4. 检查该表归属哪个服务、落在哪个数据库

### 6.3 新增或修改权限接口

除了改 proto / handler 以外，还要同步检查：

- `admin` 中的 API 目录主数据
- `resource_group`
- `sys_service_registry.http_prefix -> service_code`
- `admin -> auth` 的快照或 delta 投影
- gateway 是否需要新路由、中间件或白名单

## 7. 当前版本线状态

截至 2026-05-09，本仓库可确认的状态：

- `go test ./...` 通过
- 测试以编译 / 冒烟检查为主，覆盖率不高
- gateway 的 HTTP 统一入口是 `:8000`
- OpenAPI 同步入口是 `make frontend-api`
- `common` 服务的数据层仍偏骨架化，目前没有与 `admin` 同量级的治理主数据
- `common` 服务当前已承载站内信收件箱 / 发布记录能力，并通过内部 gRPC 依赖 `user`、`admin` 服务补足收件人和角色校验

## 8. 配套文档

- [README.md](../README.md)
- [permission-design.md](./permission-design.md)
- [api-ownership.md](./api-ownership.md)
- [table-ownership.md](./table-ownership.md)
- [auth-incremental-sync-design.md](./auth-incremental-sync-design.md)
