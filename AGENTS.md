# base-server — Agent notes

## What this repo is
- Go (`go 1.24.0`, `toolchain go1.24.4`) Kratos multi-service repo.
- Active services:
  - auth: HTTP `:8020`, gRPC `:9020`
  - user: HTTP `:8010`, gRPC `:9010`
  - admin: HTTP `:8030`, gRPC `:9030`
  - common: HTTP `:8040`, gRPC `:9040`
- Data: Ent ORM over PostgreSQL via `database/sql` + `pgx`.
- Auth: JWT HS256 + Casbin.

## Where to change things
- `app/*/service/cmd/service/` — each service entrypoint + Wire bootstrap.
- `app/*/service/internal/service/` — transport handlers.
- `app/*/service/internal/biz/` — use cases.
- `app/*/service/internal/data/` — service repositories + outbound clients.
- `app/*/service/internal/conf/` — service-local config proto + generated config structs.
- `api/protos/*/service/v1/` — source protos.
- `api/gen/go/**` — generated protobuf/kratos bindings (**do not edit**).
- `pkg/data/schema/` — Ent schema sources.
- `pkg/data/ent/` — Ent generated code (**do not edit**).
- `pkg/tools/` — shared helper functions.
- `api/openapi/` — generated OpenAPI output.

## Generated code boundaries
- `api/gen/go/**` — `make api`
- `api/openapi/**` — `make api` or `make openapi`
- `app/*/service/cmd/service/wire_gen.go` — `make wire`
- `app/*/service/internal/conf/*.pb.go` — `make config`
- `pkg/data/ent/**` — `make ent`

## Common commands
- Tooling install: `make init`
- Generate API: `make api`
- Generate config types: `make config`
- Generate Wire: `make wire`
- Generate Ent: `make ent`
- Build all services: `make build`
- Quick compile check: `go test ./...`

## Runtime config
- Each service uses its own `-conf` directory under `app/<service>/service/configs`.
- Config schema source of truth lives in the corresponding `app/<service>/service/internal/conf/conf.proto`.

## Repo-specific gotchas
- Placeholder secrets still exist in service config YAMLs; do not commit real credentials.
- Tests are mostly compile checks; there are effectively no `*_test.go` files.
- The old monolith runtime and Helm chart have been removed.

## When adding/changing APIs
- Edit the owning proto under `api/protos/*/service/v1/*.proto`.
- Run `make api`.
- Update the owning handler under `app/*/service/internal/service/`.
