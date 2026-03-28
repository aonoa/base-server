# base-server

Kratos + Go 多服务后端。当前仓库以 gateway + 四个领域服务运行，统一共享根目录 `go.mod`。

当前开发基线：

- Go：`1.25.8`
- PostgreSQL：`18`

## 服务布局

- `app/gateway/service`：统一公网入口，使用 `go-kratos/gateway` 原生 endpoints 配置转发到下游服务
- `app/auth/service`：认证、JWT、Casbin 策略、角色/API/资源权限
- `app/user/service`：用户、密码、用户资料
- `app/admin/service`：菜单、部门、系统日志
- `app/common/service`：上传、SSE、LLM 能力
- `pkg/data`：共享 Ent schema / generated client / templates
- `pkg/tools`：共享工具函数

## 常用命令

### 初始化与代码生成

```bash
make init
make api
make config
make ent
make wire
```

说明：

- `make api` 生成 `api/protos/**` 对应的 pb/http/grpc/errors，并输出 OpenAPI 到 `api/openapi/`
- `make config` 生成各服务 `app/*/service/internal/conf/*.pb.go` 以及 gateway 自定义 middleware 的 proto 配置结构
- `make ent` 基于 `pkg/data/schema/` 重新生成 `pkg/data/ent/`
- `make wire` 重新生成各服务 `wire_gen.go`

### 构建

```bash
make build
```

生成：

- `./bin/gateway-service`
- `./bin/auth-service`
- `./bin/user-service`
- `./bin/admin-service`
- `./bin/common-service`

### 运行服务

```bash
./bin/gateway-service -conf ./app/gateway/service/configs
./bin/auth-service -conf ./app/auth/service/configs
./bin/user-service -conf ./app/user/service/configs
./bin/admin-service -conf ./app/admin/service/configs
./bin/common-service -conf ./app/common/service/configs
```

默认端口：

- gateway：HTTP `8000`
- auth：HTTP `8020`，gRPC `9020`
- user：HTTP `8010`，gRPC `9010`
- admin：HTTP `8030`，gRPC `9030`
- common：HTTP `8040`，gRPC `9040`

### 本地断点调试

推荐把数据库、Redis、Jaeger、Consul 放在容器里，只在 IDE 本地启动要调试的服务。

先启动基础依赖：

```bash
make dev-env-up
```

停止基础依赖：

```bash
make dev-env-down
```

如果刚从旧版 PostgreSQL 主版本切到 PG18，直接删卷重建：

```bash
make dev-env-reset
```

直接在终端本地运行服务：

```bash
make run-gateway
make run-auth
make run-user
make run-admin
make run-common
```

IDE 里也可以直接运行对应入口，并传相同参数：

- gateway：`app/gateway/service/cmd/service/main.go` + `-conf ./app/gateway/service/configs`
- auth：`app/auth/service/cmd/service/main.go` + `-conf ./app/auth/service/configs`
- user：`app/user/service/cmd/service/main.go` + `-conf ./app/user/service/configs`
- admin：`app/admin/service/cmd/service/main.go` + `-conf ./app/admin/service/configs`
- common：`app/common/service/cmd/service/main.go` + `-conf ./app/common/service/configs`

调试组合建议：

- 调 `user` / `admin` / `common`：`make dev-env-up` 后单独启动对应服务即可
- 调 `auth`：至少同时启动 `user` + `auth`
- 调 `gateway` 登录链路：建议同时启动 `gateway` + `auth` + `user`
- 调 `gateway` 管理链路：建议同时启动 `gateway` + `admin`

`make dev-env-up` 会自动确保 `base_networks` 存在，避免 `docker-compose-env.yml` 因外部网络缺失启动失败。

## 网关与路由

gateway 现在直接使用原生 `gateway.endpoints` / `gateway.middlewares` 配置模型。每个 endpoint 都可以单独配置：

- `path`
- `method`
- `protocol`
- `timeout`
- `backends`
- `middlewares`
- `retry`
- `host`
- `stream`

这意味着：

- 不挂 middleware 的 endpoint 就是纯直通代理
- 挂了 middleware 的 endpoint 会先经过网关处理，再转发到下游
- upload、SSE 这类特殊路由也通过配置表达，而不是在代码里写死分支

默认建议客户端只访问 gateway，由 gateway 按前缀转发到四个下游 HTTP 服务：

- `/auth-api/v1/*` -> auth `:8020`
- `/user-api/v1/*` -> user `:8010`
- `/admin-api/v1/*` -> admin `:8030`
- `/common-api/v1/*` -> common `:8040`

### 内置 middleware

当前 `go-kratos/gateway` 模块版本可直接使用这些内置 middleware：

- `cors`
- `logging`
- `rewrite`
- `tracing`
- `circuitbreaker`
- `bbr`
- `transcoder`
- `streamrecorder`

它们通过 `gateway.middlewares` 或 `gateway.endpoints[].middlewares` 直接声明。

### 自定义 middleware

仓库内额外注册了这些 edge middleware：

- `jwt`：校验 Bearer Token，支持按 `path + method + host` 白名单绕过
- `whitelist`：只允许命中的 HTTP 路由通过，未命中直接拒绝
- `ratelimit`：传统限流，支持全局或按 IP 限流
- `casbin`：基于角色/API 资源关系做接口授权
- `httplog`：按接口请求记录访问日志并上报 admin

对应配置 type URL：

- `type.googleapis.com/base_server.gateway.middleware.jwt.v1.JWT`
- `type.googleapis.com/base_server.gateway.middleware.whitelist.v1.Whitelist`
- `type.googleapis.com/base_server.gateway.middleware.ratelimit.v1.RateLimit`
- `type.googleapis.com/base_server.gateway.middleware.casbin.v1.Casbin`
- `type.googleapis.com/base_server.gateway.middleware.httplog.v1.HttpLog`

建议对安全相关 middleware 打开 `required: true`，这样拼错名字或配置解析失败时会直接报错，而不是静默跳过。

### JWT 边界

当前阶段 gateway 的 JWT 负责入口控制，下游服务仍然保留原有 JWT 校验：

- gateway 继续透传 `Authorization`
- auth/user/admin/common 仍使用各自服务内的认证中间件
- 当前是双层校验，不是把信任边界完全迁到 gateway

### Stream 与特殊路由

- `/common-api/v1/file/upload` 建议单独设置更长 `timeout`
- `/common-api/v1/copilot/sse` 需要 `stream: true`
- 流式 endpoint 上不要随意叠加不支持流式的 middleware
- `circuitbreaker` 依赖其自身初始化能力，使用前应确认运行环境和配置完整

### 示例路由行为表

下表对应 `app/gateway/service/configs/config.yaml` 当前示例配置：

| 请求 | 命中路由 | 网关处理 | 预期行为 |
| --- | --- | --- | --- |
| `POST /auth-api/v1/login`，无 token | `/auth-api/v1/login` | `ratelimit(SCOPE_IP, 5 rps, burst 10)` + 全局 `cors` | 允许转发到 auth `:8020`；不要求 JWT；同一 IP 高频请求会被 `429` |
| `POST /auth-api/v1/refresh`，无 token | `/auth-api/v1/*` | `jwt`（此 path 在 jwt whitelist） + 全局 `cors` | 允许转发到 auth `:8020`；不要求 JWT |
| `GET /auth-api/v1/profile`，无 token | `/auth-api/v1/*` | `jwt` + 全局 `cors` | 网关直接返回 `401`，不会转发到 auth |
| `GET /auth-api/v1/profile`，带有效 Bearer token | `/auth-api/v1/*` | `jwt` + 全局 `cors` | 允许转发到 auth `:8020` |
| `GET /user-api/v1/profile`，无 token | `/user-api/v1/*` | `jwt` + 全局 `cors` | 网关直接返回 `401` |
| `GET /user-api/v1/profile`，带有效 Bearer token | `/user-api/v1/*` | `jwt` + 全局 `cors` | 允许转发到 user `:8010` |
| `GET /admin-api/v1/users`，带有效 Bearer token | `/admin-api/v1/*` | `jwt` + `ratelimit(SCOPE_IP, 20 rps, burst 40)` + 全局 `cors` | 允许转发到 admin `:8030` |
| `GET /admin-api/v1/users`，无 token | `/admin-api/v1/*` | `jwt` + `ratelimit` + 全局 `cors` | 网关直接返回 `401` |
| `POST /common-api/v1/file/upload`，带有效 Bearer token | `/common-api/v1/file/upload` | `jwt` + 全局 `cors` | 允许转发到 common `:8040`；超时时间 `10m` |
| `POST /common-api/v1/file/upload`，无 token | `/common-api/v1/file/upload` | `jwt` + 全局 `cors` | 网关直接返回 `401` |
| `POST /common-api/v1/copilot/sse`，带有效 Bearer token | `/common-api/v1/copilot/sse` | `jwt` + 全局 `cors` | 允许转发到 common `:8040`；按 `stream: true` 走流式代理；超时时间 `24h` |
| `POST /common-api/v1/copilot/sse`，无 token | `/common-api/v1/copilot/sse` | `jwt` + 全局 `cors` | 网关直接返回 `401` |
| 任意跨域预检 `OPTIONS` 请求，`Origin` 命中 allowOrigins | 全局 `cors` | `cors` | 由网关直接返回 CORS 预检响应，不再继续走下游业务处理 |

注意：

- README 中的“带有效 Bearer token”只表示 gateway JWT 校验通过；下游服务仍会继续做自己的 JWT 校验。
- 示例配置里 `cors.allowOrigins` 只包含 `localhost` 和 `127.0.0.1`，其他来源的跨域请求会被 CORS 拒绝。
- `/auth-api/v1/login` 因为有更精确的独立 endpoint，会优先命中登录路由，而不是落到 `/auth-api/v1/*`。

## 测试

```bash
go test ./...
```

当前仓库几乎没有 `*_test.go`，这里主要用于编译/冒烟检查。

## 数据与依赖

本地联调默认依赖 PostgreSQL + Redis。

- 只起基础依赖：`docker-compose-env.yml`
- 一次启动完整环境：根目录 `docker-compose.yml`

默认数据库分库：

- auth DB：`auth`
- user DB：`user`
- admin DB：`admin`
- common DB：`common`

### PostgreSQL 18 升级说明

仓库现在统一使用 PostgreSQL 18，并采用“跨主版本直接删卷重建”的策略，不提供旧卷原地兼容。

原因：

- PostgreSQL 18 官方镜像调整了数据目录布局
- 旧的 PG16 数据卷不能直接复用到 PG18 容器
- 本仓库已经具备 init SQL + startup migrate + seed 流程，重建本地环境成本更低也更可靠

如果你从旧主版本升级：

```bash
docker-compose -f ./docker-compose-env.yml down -v
docker compose -f ./docker-compose.yml down -v
```

然后按你的场景重建。

### 一键启动完整环境

```bash
./deploy/scripts/up.sh
```

脚本会构建并启动：

- `postgres`
- `redis`
- `jaeger`
- `consul`
- `gateway`
- `auth`
- `user`
- `admin`
- `common`

默认对外端口：

- gateway：`8000`
- postgres：`25432`
- redis：`26379`
- jaeger UI：`16686`
- consul UI：`8500`

说明：

- auth/user/admin/common 默认只加入 compose 内部网络，不额外占用宿主机 `8010/8020/8030/8040/9010/9020/9030/9040`
- gateway 会在容器内通过 `auth:8020`、`user:8010`、`admin:8030`、`common:8040` 转发到下游服务
- 如果你刚从旧 Postgres 主版本切到 PG18，先执行 `docker compose -f ./docker-compose.yml down -v`

### 容器内配置目录

`docker-compose.yml` 不直接复用 `app/*/service/configs/config.yaml`，而是挂载部署专用配置：

- `deploy/configs/gateway/config.yaml`
- `deploy/configs/auth/config.yaml`
- `deploy/configs/user/config.yaml`
- `deploy/configs/admin/config.yaml`
- `deploy/configs/common/config.yaml`

这些配置把 `127.0.0.1` 改成了 compose service name，例如：

- Postgres：`postgres:5432`
- Redis：`redis:6379`
- auth -> user gRPC：`user:9010`
- gateway -> 各服务 HTTP：`auth:8020`、`user:8010`、`admin:8030`、`common:8040`

### 数据库初始化与 seed

Postgres 首次初始化时会自动执行：

- `deploy/sql/init/00-create-databases.sql`

它只负责创建四个数据库，不负责建表。表结构由服务启动后的 Ent 自动迁移完成：

- auth：创建 `sys_role`、`sys_api_resources`、`sys_resources`、`api_resources_roles`、`resource_roles`
- user：创建 `sys_user`
- admin：创建 `sys_menu`、`sys_dept`、`sys_log`

基础数据使用单独 seed 文件：

- `deploy/sql/seed/auth.sql`
- `deploy/sql/seed/admin.sql`
- `deploy/sql/seed/user.sql`

等待服务完成首次迁移后执行：

```bash
./deploy/scripts/seed.sh
```

seed 完成后默认账号：

- `vben / 123456`：root
- `jack / 123456`：admin

### seed 说明

- `deploy/scripts/seed.sh` 可以重复执行：固定 ID 的基础角色、菜单、用户、API 资源会按 seed 内容更新，关联关系会跳过已存在记录
- 默认账号 `jack` / `vben` 在重复 seed 时会被归一化到 seed 里的固定记录，避免同名默认账号重复累积
- 如果想回到完全空白的本地环境，建议清理卷后重建：

```bash
docker compose -f ./docker-compose.yml down -v
./deploy/scripts/up.sh
./deploy/scripts/seed.sh
```

- 当前 seed 已按拆分服务结构整理，不再直接整份导入旧的 `deploy/sql/pg_dump.sql`
- 旧 dump 中的 `user_roles` 已转换为当前 `sys_user.role_id`
- 旧 `basic-api` 路径已按当前拆分后的 `/auth-api/v1/*`、`/user-api/v1/*`、`/admin-api/v1/*`、`/common-api/v1/*` 重新整理

## Docker

根目录 `Dockerfile` 现在是通用的单服务镜像构建文件，通过 `SERVICE` 选择目标服务：

```bash
docker build --build-arg SERVICE=gateway -t base-server-gateway:v1.1.0 .
docker build --build-arg SERVICE=auth -t base-server-auth:v1.1.0 .
docker build --build-arg SERVICE=user -t base-server-user:v1.1.0 .
```

运行时挂载对应服务配置目录到 `/app/configs`：

```bash
docker run --rm -v "$PWD/app/auth/service/configs":/app/configs base-server-auth:v1.1.0
```

## Helm

仓库内旧版单体 Helm chart 已移除；如需部署，请按 gateway + 四个领域服务分别编排。

## Casbin

手动导入或迁移数据后，如 `casbin_rules` 的序列与现有数据冲突，可手动重置：

```sql
SELECT MAX(id) FROM casbin_rules;
SELECT last_value FROM pg_sequences WHERE schemaname = 'public' AND sequencename = 'casbin_rules_id_seq';
ALTER SEQUENCE casbin_rules_id_seq RESTART WITH {max_id + 1};
```

## 故障排查

### 本地启动 user 报 `127.0.0.1:25432 connection refused`

通常说明本地依赖里的 postgres 没起来，而不是 user 服务本身有问题。优先执行：

```bash
make dev-env-reset
```

如果是从旧版 PG 升到 PG18，这一步尤其必要，因为旧卷不会自动迁移到新的目录布局。

### init SQL 没生效

`deploy/sql/init/00-create-databases.sql` 只会在新数据卷首次初始化时执行。若你已经起过旧卷，再改 compose 或 init SQL，必须删卷后重建：

```bash
docker-compose -f ./docker-compose-env.yml down -v
docker compose -f ./docker-compose.yml down -v
```
