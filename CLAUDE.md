# CLAUDE.md

This file provides guidance to Claude Code when working in this repository.

## Common commands

### Setup and code generation
- `make init` — install protobuf/Kratos/Wire tooling used by this repo
- `make api` — regenerate protobuf, Kratos HTTP/gRPC bindings, error helpers, and OpenAPI output in `api/openapi/`
- `make config` — regenerate config structs for each split service from `app/*/service/internal/conf/conf.proto`
- `make wire` — regenerate dependency injection code (`go generate ./...`)
- `make ent` — regenerate Ent ORM code from `pkg/data/schema/`
- `make all` — run the main generation flow (`api`, `config`, `generate`)

Notes:
- `make init` installs plugins, but not `protoc` itself.
- `make ent` expects the `ent` CLI to already be on `PATH`.

### Build and run
- `make build` — build split service binaries into `./bin/`
- `./bin/gateway-service -conf ./app/gateway/service/configs`
- `./bin/auth-service -conf ./app/auth/service/configs`
- `./bin/user-service -conf ./app/user/service/configs`
- `./bin/admin-service -conf ./app/admin/service/configs`
- `./bin/common-service -conf ./app/common/service/configs`

### Tests
- `go test ./...` — repo-wide compile/smoke test
- `go test ./path/to/package -run '^TestName$'` — run a single test

Notes:
- There are currently no meaningful `*_test.go` files in this repository, so `go test ./...` is mainly a compile check.

### Docker
- Generic per-service image build:
  - `docker build --build-arg SERVICE=gateway -t base-server-gateway:v1.1.0 .`
  - `docker build --build-arg SERVICE=auth -t base-server-auth:v1.1.0 .`
  - `docker build --build-arg SERVICE=user -t base-server-user:v1.1.0 .`
  - `docker build --build-arg SERVICE=admin -t base-server-admin:v1.1.0 .`
  - `docker build --build-arg SERVICE=common -t base-server-common:v1.1.0 .`
- Run by mounting the target service config directory to `/app/configs`

## Architecture overview

This is a Kratos-based Go backend with protobuf-first APIs, Ent for persistence, JWT authentication, Casbin authorization, and split services under `app/*/service`.

### Active services
- `gateway` — external HTTP entrypoint, endpoint routing, edge middleware
- `auth` — login, refresh, Casbin policy execution, authorization projection loading
- `user` — user CRUD, profile, password, auth identity data
- `admin` — platform governance, API catalog, projection source status, menu, dept, syslog
- `common` — upload, SSE, LLM integration

### High-level flow
1. API contracts are defined in `api/protos/*/service/v1/`.
2. Generated transport bindings in `api/gen/go/` are implemented by handlers in `app/*/service/internal/service/`.
3. Handlers delegate business logic to use cases in `app/*/service/internal/biz/`.
4. Use cases depend on repository interfaces implemented in `app/*/service/internal/data/`.
5. Shared persistence code lives in `pkg/data/` and shared helpers in `pkg/tools/`.
6. Gateway reads endpoint config and forwards external HTTP traffic to the owning split service.

### Important directories
- `app/gateway/service/`
- `app/auth/service/`
- `app/user/service/`
- `app/admin/service/`
- `app/common/service/`
- `api/protos/` — source protobuf definitions
- `api/gen/go/` — generated protobuf/Kratos code
- `api/openapi/` — generated OpenAPI output
- `pkg/data/schema/` — Ent schema sources
- `pkg/data/ent/` — Ent generated code
- `pkg/data/template/` — Ent templates
- `pkg/tools/` — shared helper code
- `third_party/` — protobuf dependencies used during code generation

## Repo-specific architecture details

### Bootstrap and runtime
- Each service has its own entrypoint under `app/<service>/service/cmd/service/main.go`.
- Each split service starts both HTTP and gRPC servers.
- Gateway is a separate HTTP edge service and should be included when changes affect routing, auth, or externally exposed paths.
- Each service loads config from its own `app/<service>/service/configs` directory.

### API layer
- Active API development should use:
  - `api/protos/auth/service/v1/`
  - `api/protos/user/service/v1/`
  - `api/protos/admin/service/v1/`
  - `api/protos/common/service/v1/`
- Avoid adding ad-hoc HTTP routes; add endpoints in proto files first, then regenerate.

### Data layer
- Shared Ent schema source lives in `pkg/data/schema/`.
- Shared generated Ent client lives in `pkg/data/ent/`.
- Service repositories live in `app/*/service/internal/data/`.
- Services use PostgreSQL via `database/sql` + `pgx`.

### Config model
- Each service has its own config schema under `app/*/service/internal/conf/conf.proto`.
- Keep each service’s `configs/config.yaml` aligned with its proto.

## Generated code boundaries

Do not hand-edit generated files. Edit the source, then regenerate.

- `api/gen/go/**` — generated from `api/protos/**` via `make api`
- `api/openapi/**` — generated via `make api` or `make openapi`
- `app/*/service/cmd/service/wire_gen.go` — generated via `make wire`
- `app/*/service/internal/conf/*.pb.go` — generated via `make config`
- `pkg/data/ent/**` — generated via `make ent`

Edit these source locations instead:
- `api/protos/**`
- `app/*/service/internal/conf/conf.proto`
- `pkg/data/schema/**`
- `app/*/service/cmd/service/wire.go`

## Common change workflows

### Adding or changing an API
1. Choose the split service proto under `api/protos/*/service/v1/` that technically serves the API.
2. Edit that proto file.
3. Run `make api`.
4. Update the matching handler under `app/*/service/internal/service/`.
5. Keep API catalog ownership consistent: one `path + method` has exactly one API catalog record with `resources_group`; service ownership is resolved from `sys_service_registry.http_prefix`. See `docs/api-ownership.md`.

### Changing config
1. Edit the target service’s `app/<service>/service/internal/conf/conf.proto`.
2. Run `make config`.
3. Update that service’s `app/<service>/service/configs/config.yaml`.

### Changing database schema
1. Edit `pkg/data/schema/`.
2. Run `make ent`.
3. Update affected service repository logic under `app/*/service/internal/data/`.

### Changing dependency injection
1. Edit `app/<service>/service/cmd/service/wire.go`.
2. Run `make wire`.

## Repo-specific gotchas
- Placeholder credentials remain in service config files; do not commit real secrets.
- Gateway config may expose behavior that is not obvious from service proto files alone, especially for auth, SSE, upload, and timeout behavior.
- There is no checked-in Helm chart anymore; deploy the four domain services plus gateway separately.
- CI Docker packaging should build per-service images, not a monolith.
- `go.mod` targets Go `1.25.8`.
