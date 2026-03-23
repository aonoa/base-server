# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common commands

### Setup and code generation
- `make init` — install protobuf/Kratos/Wire tooling used by this repo
- `make api` — regenerate protobuf, Kratos HTTP/gRPC bindings, error helpers, and `cmd/base-server/assets/openapi.yaml`
- `make config` — regenerate config structs from `internal/conf/conf.proto`
- `make wire` — regenerate dependency injection code (`go generate ./...`)
- `make ent` — regenerate Ent ORM code from `internal/data/schema/`
- `make all` — run the main generation flow (`api`, `config`, `generate`)

Notes:
- `make init` installs plugins, but not `protoc` itself.
- `make ent` expects the `ent` CLI to already be on `PATH`.

### Build and run
- `make build` — build binaries into `./bin/`
- `./bin/base-server -conf ./configs` — run the built server locally
- `go run ./cmd/base-server -conf ./configs` — run without building a binary first

### Tests
- `go test ./...` — repo-wide compile/smoke test
- `go test ./path/to/package -run '^TestName$'` — run a single test

Notes:
- There are currently no `*_test.go` files in this repository, so `go test ./...` is mainly a compile check.
- There is no repo-standard lint target or golangci-lint config checked in.

### Docker and deployment
- `docker build -t base-server:v1.1.0 .`
- `docker run --rm -p 8000:8000 -p 9000:9000 -v </path/to/configs>:/data/conf <image>`
- `helm template base-server ./deploy/helm/base-server`
- `helm install base-server ./deploy/helm/base-server`
- `helm upgrade base-server ./deploy/helm/base-server`

## Architecture overview

This is a Kratos-based Go backend with protobuf-first APIs, Ent for persistence, JWT authentication, Casbin authorization, and an in-process cron worker.

### High-level flow
1. API contracts are defined in `api/protos/`.
2. Generated transport bindings in `api/gen/go/` are implemented by handlers in `internal/service/`.
3. Handlers delegate business logic to use cases in `internal/biz/`.
4. Use cases depend on repository interfaces implemented in `internal/data/`.
5. `cmd/base-server/` wires everything together with Google Wire and starts the Kratos app.

### Important directories
- `cmd/base-server/` — app entrypoint, Wire setup, embedded OpenAPI assets
- `api/protos/` — source protobuf definitions
- `api/gen/go/` — generated protobuf/Kratos code
- `internal/service/` — transport-facing service implementations
- `internal/biz/` — core use cases and auth logic
- `internal/data/` — repositories, Ent client setup, cache setup, schema definitions
- `internal/server/` — HTTP/gRPC server construction and middleware, plus cron worker
- `internal/conf/` — config schema proto and generated config structs
- `configs/` — local runtime YAML config
- `deploy/helm/base-server/` — Helm chart and embedded runtime config templates
- `third_party/` — protobuf dependencies used during code generation

## Repo-specific architecture details

### Bootstrap and runtime
- The main entrypoint is `cmd/base-server/main.go`.
- Config is loaded from the `-conf` flag, which defaults to `./configs`.
- HTTP is started as part of the Kratos app.
- gRPC server wiring exists in `internal/server/grpc.go`, but it is currently not started because `gs` is commented out in `cmd/base-server/main.go`.
- A cron worker is also started as part of the app and loads jobs from config.

### API layer
- The main API surface is in `api/protos/base_api/v1/base.proto`.
- Additional transport definitions exist for SSE and file upload.
- Swagger UI is served from embedded generated assets at `/docs/` by `internal/server/http.go`.
- Avoid adding ad-hoc HTTP routes; add endpoints in proto files first, then regenerate.

### Business and auth
- `internal/biz/base.go` contains the main application use cases: login, tokens, users, roles, menus, departments, APIs, resources, and logs.
- `internal/biz/auth.go` handles Casbin integration and policy reload/rebuild logic.
- Authentication uses JWT HS256.
- Authorization is Casbin-backed and policy is sourced from database-backed entities.

### Data layer
- `internal/data/data.go` opens the database, creates the Ent client, enables SQL debug logging, auto-creates schema on startup, and initializes a local Ristretto-backed cache.
- Repository implementations live in `internal/data/base.go`.
- Ent schema sources live in `internal/data/schema/`.

### Config model
- `internal/conf/conf.proto` is the source of truth for runtime config.
- The main config sections are: `server`, `logger`, `data`, `auth`, `menus`, `job`, and `llm`.
- Keep `configs/config.yaml` aligned with `conf.proto`.

## Generated code boundaries

Do not hand-edit generated files. Edit the source, then regenerate.

- `api/gen/go/**` — generated from `api/protos/**` via `make api`
- `cmd/base-server/assets/openapi.yaml` — generated via `make api` or `make openapi`
- `cmd/base-server/wire_gen.go` — generated via `make wire`
- `internal/conf/conf.pb.go` — generated via `make config`
- `internal/data/ent/**` — generated via `make ent`

Edit these source locations instead:
- `api/protos/**`
- `cmd/base-server/wire.go`
- `internal/conf/conf.proto`
- `internal/data/schema/**`

## Common change workflows

### Adding or changing an API
1. Edit the proto in `api/protos/base_api/v1/`.
2. Run `make api`.
3. Update handlers in `internal/service/`.
4. If transport behavior changed, verify the generated OpenAPI output via `/docs/`.

### Changing config
1. Edit `internal/conf/conf.proto`.
2. Run `make config`.
3. Update `configs/config.yaml` to match.

### Changing database schema
1. Edit `internal/data/schema/`.
2. Run `make ent`.
3. Update repository logic in `internal/data/base.go` or use case logic as needed.

### Changing dependency injection
1. Edit `cmd/base-server/wire.go`.
2. Run `make wire`.

## Repo-specific gotchas

- `README.md` still contains generic Kratos template instructions; prefer the actual binary name `base-server` and the Makefile in this repo.
- `configs/config.yaml` and Helm config templates contain placeholder credentials and connection strings; treat them as sensitive config, not safe defaults to copy elsewhere.
- The Helm chart mounts `config.yaml` to `/data/conf/config.yaml` and embeds Casbin config under `/app/authconf/`.
- The Helm service template uses a fixed NodePort `30080`.
- `README.md` documents a Postgres sequence-reset workaround for `casbin_rules_id_seq` after manual imports/migrations.
- CI currently only builds Docker artifacts on tag pushes matching `v*`; there is no normal test/lint workflow in GitHub Actions.
- `go.mod` targets Go `1.24.0` with toolchain `go1.24.4`, while the Docker builder image uses Go `1.23.1`; keep this mismatch in mind if container builds fail.
