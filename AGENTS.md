# base-server — Agent notes

## What this repo is
- Go (go.mod: `go 1.24.0`, `toolchain go1.24.4`), Kratos v2 service.
- Transports: HTTP `:8000`; gRPC `:9000` exists but is currently not started in `cmd/base-server/main.go` (`gs` commented).
- Data: Ent ORM (generated), Postgres via `database/sql` + `pgx` driver string; schema auto-create on startup.
- AuthZ/AuthN: JWT HS256 + Casbin. Casbin is backed by DB via `casbin/ent-adapter`.

## Where to change things (map)
- `cmd/base-server/` — entrypoint + Wire DI + embedded OpenAPI assets.
- `api/protos/` — **source** protos (edit here).
- `api/gen/go/` — **generated** protobuf/kratos bindings (**do not edit**).
- `internal/conf/` — config schema proto (`conf.proto`) + generated Go.
- `internal/data/schema/` — Ent schema sources (edit here).
- `internal/data/ent/` — Ent generated code (**do not edit**).
- `internal/biz/` — usecases (auth, menus, users, etc.).
- `internal/service/` — transport handlers (HTTP/gRPC) implementing generated interfaces.
- `internal/server/` — HTTP/gRPC server wiring + middleware.
- `deploy/helm/base-server/` — Helm chart (ConfigMap embeds app + casbin config).

## Generated code boundaries (changes get overwritten)
- `api/gen/go/**` (from `api/protos/**`) — `make api`
- `cmd/base-server/assets/openapi.yaml` — `make api` or `make openapi`
- `cmd/base-server/wire_gen.go` — `make wire` (runs `go generate ./...`)
- `internal/conf/conf.pb.go` — `make config`
- `internal/data/ent/**` — `make ent`

## Common commands
- Tooling install: `make init` (protoc plugins, wire, kratos)
- Generate API: `make api`
- Generate config types: `make config`
- Generate Wire: `make wire`
- Generate Ent: `make ent`
- Build: `make build` (writes `./bin/`)
- Quick compile check: `go test ./...`

## Runtime entrypoint / config
- Binary: `cmd/base-server/main.go`
- Config path flag: `-conf` (default `./configs`)
- Docker runs with `-conf /data/conf` and expects `config.yaml` in that directory.
- Config schema source of truth: `internal/conf/conf.proto`.

## Repo-specific gotchas
- Secrets are currently committed in plain YAML:
  - `configs/config.yaml` includes an LLM API key-like value.
  - `deploy/helm/base-server/templates/configmap.yaml` embeds DB source + auth keys.
  Treat these as placeholders; don’t copy real credentials into git.
- Tests: no `*_test.go` found; CI only builds docker on tag push.
- Casbin + manual DB imports: read `README.md` (sequence reset for `casbin_rules_id_seq`).

## When adding/changing APIs
- Edit protos in `api/protos/base_api/v1/*.proto`.
- Regenerate: `make api`.
- Implement/adjust handlers in `internal/service/` and wiring in `internal/server/`.
