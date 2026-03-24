# base-server

Kratos + Go 多服务后端。当前仓库只保留拆分后的四个服务，统一共享根目录 `go.mod`。

## 服务布局

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
- `make config` 生成各服务 `app/*/service/internal/conf/*.pb.go`
- `make ent` 基于 `pkg/data/schema/` 重新生成 `pkg/data/ent/`
- `make wire` 重新生成各服务 `wire_gen.go`

### 构建

```bash
make build
```

生成：

- `./bin/auth-service`
- `./bin/user-service`
- `./bin/admin-service`
- `./bin/common-service`

### 运行服务

```bash
./bin/auth-service -conf ./app/auth/service/configs
./bin/user-service -conf ./app/user/service/configs
./bin/admin-service -conf ./app/admin/service/configs
./bin/common-service -conf ./app/common/service/configs
```

默认端口：

- auth：HTTP `8020`，gRPC `9020`
- user：HTTP `8010`，gRPC `9010`
- admin：HTTP `8030`，gRPC `9030`
- common：HTTP `8040`，gRPC `9040`

## 测试

```bash
go test ./...
```

当前仓库几乎没有 `*_test.go`，这里主要用于编译/冒烟检查。

## 数据与依赖

本地联调默认依赖 PostgreSQL + Redis，可结合仓库内的 `docker-compose-env.yml` 启动。

默认数据库分库：

- auth DB：`auth`
- user DB：`user`
- admin DB：`admin`
- common DB：`common`

## Docker

根目录 `Dockerfile` 现在是通用的单服务镜像构建文件，通过 `SERVICE` 选择目标服务：

```bash
docker build --build-arg SERVICE=auth -t base-server-auth:v1.1.0 .
docker build --build-arg SERVICE=user -t base-server-user:v1.1.0 .
```

运行时挂载对应服务配置目录到 `/app/configs`：

```bash
docker run --rm -v "$PWD/app/auth/service/configs":/app/configs base-server-auth:v1.1.0
```

## Helm

仓库内旧版单体 Helm chart 已移除；如需部署，请按四个服务分别编排。

## Casbin

手动导入或迁移数据后，如 `casbin_rules` 的序列与现有数据冲突，可手动重置：

```sql
SELECT MAX(id) FROM casbin_rules;
SELECT last_value FROM pg_sequences WHERE schemaname = 'public' AND sequencename = 'casbin_rules_id_seq';
ALTER SEQUENCE casbin_rules_id_seq RESTART WITH {max_id + 1};
```
