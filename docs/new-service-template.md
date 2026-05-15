# 新增服务模板与步骤

## 1. 适用范围

本文只适用于 `base-server/monorepo` 微服务版本线，用于新增一个独立后端服务，例如：

```text
<service_code>         billing
<service_name>         计费服务
<http_prefix>          /billing-api
<grpc_service>         api.billing.service.v1.BillingService
<http_port>            8050
<grpc_port>            9050
<database_name>        billing
<resource_group>       billing
```

如果只是给已有服务增加业务模块，不需要新增 `app/<service>/service`，应优先在已有服务内按 proto、service、biz、data 分层扩展。

## 2. 新服务最小模板

新增服务至少需要这些路径：

```text
api/protos/<service_code>/service/v1/<service_code>.proto

app/<service_code>/service/
├── cmd/service/main.go
├── cmd/service/wire.go
├── cmd/service/wire_gen.go
├── configs/config.yaml
├── internal/biz/biz.go
├── internal/conf/conf.proto
├── internal/conf/conf.pb.go
├── internal/data/data.go
├── internal/server/grpc.go
├── internal/server/http.go
├── internal/server/server.go
└── internal/service/service.go

deploy/configs/<service_code>/config.yaml
```

通常从现有服务复制骨架：

- 简单 CRUD / 独立业务服务：参考 `app/user/service`。
- 需要数据库、对外 HTTP、内部 gRPC、跨服务 gRPC 客户端：参考 `app/common/service`。
- 需要平台治理、权限投影、启动同步：参考 `app/admin/service`，但不要把 admin 的权限主数据职责复制到普通服务。

复制后必须全量替换包路径、服务名、默认服务名、端口、数据库名、proto package、Go 类型名和生成代码注册函数。

## 3. 实施步骤

### 3.1 定义服务边界

先确认并记录：

- 服务职责：这个服务拥有什么业务主数据，哪些能力只是调用别的服务。
- API 前缀：使用 `/<service_code>-api/v1/*`，不要复用其他服务前缀。
- 数据库：是否需要独立数据库；如果需要，数据库名默认等于 `service_code`。
- 表归属：新增 Ent schema 放在 `pkg/data/schema/`，但真正归属由服务自己的迁移列表决定。
- 权限资源组：新增 API 要归属哪个 `resources_group`。
- 前端影响：前端是否需要调用新 API；需要则必须同步 OpenAPI 和生成客户端。

### 3.2 新增 API 契约

创建：

```text
api/protos/<service_code>/service/v1/<service_code>.proto
```

基础模板：

```proto
syntax = "proto3";

package api.<service_code>.service.v1;

import "google/api/annotations.proto";
import "google/protobuf/empty.proto";

option go_package = "base-server/api/gen/go/<service_code>/service/v1;v1";
option java_multiple_files = true;
option java_package = "api.<service_code>.service.v1";

service <ServiceName>Service {
  rpc GetWalkRoute(google.protobuf.Empty) returns (GetWalkRouteReply);

  rpc GetExampleList(GetExampleListRequest) returns (GetExampleListReply) {
    option (google.api.http) = {
      get: "/<service_code>-api/v1/examples"
    };
  }
}

message WalkRouteItem {
  string url = 1;
  string method = 2;
}

message GetWalkRouteReply {
  repeated WalkRouteItem items = 1;
}

message GetExampleListRequest {
  int64 current_page = 1;
  int64 page_size = 2;
}

message GetExampleListReply {
  repeated string items = 1;
  int64 total = 2;
}
```

要求：

- HTTP path 必须是 gateway 可见路径，例如 `/<service_code>-api/v1/examples`。
- 内部 gRPC 方法可以不配置 `google.api.http`。
- 每个服务保留 `GetWalkRoute`，便于 admin 聚合路由目录。

生成：

```bash
make api
```

生成物会落到：

```text
api/gen/go/<service_code>/service/v1/**
api/openapi/openapi.yaml
```

不要手工编辑 `api/gen/go/**`。

### 3.3 新增服务目录

从参考服务复制：

```bash
cp -R app/common/service app/<service_code>/service
```

然后逐项替换：

- `cmd/service/main.go` 中的 `defaultServiceName`。
- 所有 import 路径中的原服务名。
- `internal/service/service.go` 中的服务结构体和构造函数。
- `internal/server/http.go` 和 `internal/server/grpc.go` 中的注册函数。
- `internal/biz`、`internal/data` 的 ProviderSet 和 repo 命名。
- 日志文件名、数据库名、端口号。

`internal/service/service.go` 中应保留路由导出能力：

```go
func (s *<ServiceName>Service) GetWalkRoute(ctx context.Context, req *emptypb.Empty) (*v1.GetWalkRouteReply, error) {
	items, err := tools.WalkHTTPRoutes(s.RestServer)
	if err != nil {
		return nil, err
	}
	res := &v1.GetWalkRouteReply{Items: make([]*v1.WalkRouteItem, 0, len(items))}
	for _, item := range tools.SortAndUniqueWalkRoutes(items) {
		res.Items = append(res.Items, &v1.WalkRouteItem{Url: item.URL, Method: item.Method})
	}
	return res, nil
}
```

HTTP server 注册后要保存 `RestServer`：

```go
srv := http.NewServer(opts...)
v1.Register<ServiceName>ServiceHTTPServer(srv, svc)
svc.RestServer = srv
return srv
```

### 3.4 新增配置结构和本地配置

编辑：

```text
app/<service_code>/service/internal/conf/conf.proto
app/<service_code>/service/configs/config.yaml
deploy/configs/<service_code>/config.yaml
```

配置至少包含：

```yaml
server:
  http:
    addr: 0.0.0.0:<http_port>
    timeout: 1s
  grpc:
    addr: 0.0.0.0:<grpc_port>
    timeout: 1s
data:
  database:
    driver: pgx
    source: postgresql://postgres:postgres.666@127.0.0.1:25432/<database_name>?sslmode=disable
  redis:
    addr: 127.0.0.1:26379
    read_timeout: 0.2s
    write_timeout: 0.2s
auth:
  api_key: auth-service-secret
  whitelist: []
logger:
  level: info
  filename: logs/<service_code>-service.log
  max_size: 100
  max_backups: 10
  max_age: 30
  compress: false
```

如果服务需要调用其他服务，在 `Services` 配置里显式声明对应 gRPC endpoint，不要在代码里写死端口。

注意本地配置和容器配置的地址不同：

- `app/<service_code>/service/configs/config.yaml` 面向本机运行，数据库和 Redis 使用 `127.0.0.1:25432`、`127.0.0.1:26379`，跨服务 gRPC 使用 `127.0.0.1:<grpc_port>`。
- `deploy/configs/<service_code>/config.yaml` 面向 Docker Compose，数据库和 Redis 使用 `postgres:5432`、`redis:6379`，跨服务 gRPC 使用 `<service_code>:<grpc_port>`。

生成配置代码：

```bash
make config
```

### 3.5 接入 Wire

确认：

```text
app/<service_code>/service/cmd/service/wire.go
app/<service_code>/service/internal/server/server.go
app/<service_code>/service/internal/service/service.go
app/<service_code>/service/internal/biz/biz.go
app/<service_code>/service/internal/data/data.go
```

每层都要暴露 `ProviderSet`，`wire.go` 组合顺序参考：

```go
func wireApp(*conf.Server, *conf.Data, *conf.Auth, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
```

如果配置里有 `Services`、`Llm` 等额外节点，`wireApp` 参数要和 `main.go` 中调用保持一致。

生成：

```bash
make wire
```

### 3.6 接入数据表

如果服务拥有自己的表：

1. 在 `pkg/data/schema/` 新增 Ent schema。
2. 在 `app/<service_code>/service/internal/data/data.go` 中定义 `<service_code>Tables()`。
3. `migrate.Create(..., <service_code>Tables(), schema.WithForeignKeys(false))` 只迁移本服务拥有的表。
4. 更新 `docs/table-ownership.md`，写清 schema、实际表名、数据库和主归属服务。
5. 如果是全新数据库，更新 `deploy/sql/init/00-create-databases.sql`。

生成：

```bash
make ent
```

不要只因为 Ent schema 在 `pkg/data/schema/` 下就让多个服务同时迁移同一张表。

### 3.7 接入 Makefile

在根目录 `Makefile` 新增：

```makefile
.PHONY: build-<service_code>
# build <service_code> service
build-<service_code>:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/<service_code>-service ./app/<service_code>/service/cmd/service

.PHONY: run-<service_code>
# run <service_code> service with local config
run-<service_code>:
	go run ./app/<service_code>/service/cmd/service -conf ./app/<service_code>/service/configs
```

同时把 `build-<service_code>` 加入 `build` 目标。

### 3.8 接入 Docker Compose 和部署配置

更新 `docker-compose.yml`：

```yaml
  <service_code>:
    build:
      context: .
      args:
        SERVICE: <service_code>
    image: base-server-<service_code>:local
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - ./deploy/configs/<service_code>:/app/configs:ro
```

如果 gateway 要转发到该服务，也要把 gateway 的 `depends_on` 加上新服务。

更新 `deploy/scripts/up.sh` 中的容器列表，方便旧版 `docker-compose` recreate workaround 清理新服务容器。

### 3.9 接入 Gateway

本地配置和部署配置都要更新：

```text
app/gateway/service/configs/config.yaml
deploy/configs/gateway/config.yaml
```

新增 endpoint：

```yaml
- path: /<service_code>-api/v1/*
  method: "*"
  protocol: HTTP
  timeout: 30s
  backends:
    - target: 127.0.0.1:<http_port>
  middlewares:
    - name: jwt
      required: true
      options:
        "@type": type.googleapis.com/base_server.gateway.middleware.jwt.v1.JWT
        signingKey: auth-service-secret
    - name: casbin
      required: true
      options:
        "@type": type.googleapis.com/base_server.gateway.middleware.casbin.v1.Casbin
        signingKey: auth-service-secret
    - name: httplog
      required: true
      options:
        "@type": type.googleapis.com/base_server.gateway.middleware.httplog.v1.HttpLog
        signingKey: auth-service-secret
```

上面示例里的 backend target 是本地运行配置。`deploy/configs/gateway/config.yaml` 中应使用 Compose 服务名：

```yaml
backends:
  - target: <service_code>:<http_port>
```

特殊接口按需单独建更精确 endpoint：

- 登录、公开回调、健康检查：评估是否加入 JWT/Casbin 白名单。
- 上传、大文件：设置更长 timeout。
- SSE/流式响应：设置 `stream: true`，不要叠加不支持流式的 middleware。

### 3.10 接入服务注册和权限 API 目录

服务注册主数据在 admin 库，默认 seed 文件：

```text
deploy/sql/seed/admin.sql
```

新增服务注册：

```sql
INSERT INTO sys_service_registry (
  id, create_time, update_time, service_code, service_name, http_prefix, grpc_service, status, projection_enabled, description
)
VALUES (
  'svc-<service_code>',
  '2026-05-15 00:00:00+08',
  '2026-05-15 00:00:00+08',
  '<service_code>',
  '<service_name>',
  '/<service_code>-api',
  'api.<service_code>.service.v1.<ServiceName>Service',
  true,
  false,
  '<service_description>'
)
ON CONFLICT (service_code) DO UPDATE SET
  update_time = EXCLUDED.update_time,
  service_name = EXCLUDED.service_name,
  http_prefix = EXCLUDED.http_prefix,
  grpc_service = EXCLUDED.grpc_service,
  status = EXCLUDED.status,
  projection_enabled = EXCLUDED.projection_enabled,
  description = EXCLUDED.description;
```

如果 `admin.sql` 里有清理服务列表，例如 `WHERE service_code NOT IN (...)`，必须把新服务加入白名单。

新增资源组：

```sql
INSERT INTO sys_resources (id, create_time, update_time, name, type, value, method, description)
VALUES (
  '<uuid>',
  '2026-05-15 00:00:00+08',
  '2026-05-15 00:00:00+08',
  '<service_name>接口权限',
  'api',
  '<resource_group>',
  '(GET|POST|PUT|DELETE)',
  '<description>'
)
ON CONFLICT (id) DO UPDATE SET
  update_time = EXCLUDED.update_time,
  name = EXCLUDED.name,
  type = EXCLUDED.type,
  value = EXCLUDED.value,
  method = EXCLUDED.method,
  description = EXCLUDED.description;
```

新增 API 目录：

```sql
INSERT INTO sys_api_resources (
  id, create_time, update_time, description, path, method, module, module_description, resources_group
)
VALUES (
  'api-<service_code>-example-list',
  '2026-05-15 00:00:00+08',
  '2026-05-15 00:00:00+08',
  '获取示例列表',
  '/<service_code>-api/v1/examples',
  'GET',
  '<service_code>',
  '<service_name>',
  '<resource_group>'
)
ON CONFLICT (path, method) DO UPDATE SET
  update_time = EXCLUDED.update_time,
  description = EXCLUDED.description,
  module = EXCLUDED.module,
  module_description = EXCLUDED.module_description,
  resources_group = EXCLUDED.resources_group;
```

授权给默认角色或管理员时，按当前权限模型写入 `resource_roles` 和 `api_resources_roles`，再触发 admin 到 auth 的权限投影同步。

普通业务服务不应该绕过 admin 直接写 auth 的 Casbin 数据。

### 3.11 接入 admin 路由聚合

如果希望管理页“获取系统所有 API 接口”能看到新服务路由，需要扩展 admin 对新服务的 gRPC 客户端：

- `app/admin/service/internal/conf/conf.proto`：在 `Services` 中增加新服务 endpoint。
- `app/admin/service/configs/config.yaml` 和 `deploy/configs/admin/config.yaml`：增加新服务 gRPC 地址。
- `app/admin/service/internal/data/data.go`：建立新服务 gRPC client，补充 cleanup。
- `app/admin/service/internal/data/data.go`：新增 `List<ServiceName>WalkRoutes`。
- `app/admin/service/internal/biz/biz.go`：在 `GetWalkRoute` fetchers 中加入新服务。

完成后运行：

```bash
make config
make wire
```

如果暂时不接入 admin 聚合，也必须在文档或任务说明中记录原因；否则 API 目录人工维护时容易漏新服务路由。

### 3.12 同步前端 OpenAPI

只有当前端需要调用新接口时才执行：

```bash
make frontend-api
```

该命令会：

```text
base-server/api/openapi/openapi.yaml -> ../vben-admin/openapi.yaml
cd ../vben-admin && pnpm run generate:api
```

然后在 `vben-admin/monorepo` 下检查生成客户端、类型和调用点。不要把单体版 `/basic-api/*` 口径混进微服务版。

## 4. 验证清单

代码和生成验证：

```bash
make api
make config
make ent
make wire
make build
go test ./...
```

按实际改动裁剪：

- 只改 proto：至少 `make api && make build`。
- 改配置 proto：至少 `make config && make build`。
- 改 Ent schema：至少 `make ent && go test ./...`。
- 改 Wire provider：至少 `make wire && make build`。
- 改前端调用：后端 `make frontend-api` 后，前端运行相关 lint/test/build。

本地运行验证：

```bash
make dev-env-up
make run-<service_code>
make run-gateway
```

通过 gateway 访问：

```bash
curl -i http://127.0.0.1:8000/<service_code>-api/v1/examples
```

如果接口需要 JWT/Casbin，使用有效 Bearer token，并确认当前组织下角色已经拥有对应 `resources_group` 的 API 权限。

## 5. 提交前检查

提交前确认：

- `git status --short` 只包含本次新增服务相关文件。
- 生成代码已同步，且没有手工编辑 generated 文件。
- 本地配置和 `deploy/configs` 都已更新。
- gateway 本地配置和部署配置都已更新。
- `deploy/sql/init` 和 `deploy/sql/seed` 已覆盖数据库、服务注册、API 目录和权限资源组。
- `docs/monorepo-overview.md`、`docs/api-ownership.md`、`docs/table-ownership.md` 已按影响更新。
- 如果前端调用新 API，`vben-admin` 的 OpenAPI 和调用点已同步。
