# base-server — Agent Notes

## Source Of Truth

- Read `docs/project_profile.md`, `docs/architecture/current_architecture.md`, `docs/architecture/tech_stack.md`, `docs/planning/module_boundaries.md`, `docs/planning/verification.md`, `docs/planning/risk_register.md`, `docs/progress.md`, and `docs/worklog.md` before planning or editing.
- Treat repository docs as durable project state; chat is temporary.
- Record changed requirements, assumptions, task boundaries, or human feedback in docs before relying on them.

## What This Repo Is

- These notes describe the monolithic `master` branch. The microservice backend is on the separate `monorepo` branch.
- Read `../AGENTS.md` for cross-repository coordination and full-stack feature rules.
- Read `../REPO_STRUCTURE.md` and `../STARTUP_AND_INTEGRATION.md` before switching branches or syncing frontend API code.
- Go service using `go 1.24.0` and `toolchain go1.24.4`.
- Kratos v2 HTTP service with gRPC wired but not started.
- Ent ORM over PostgreSQL with startup schema creation.
- JWT HS256 authentication plus Casbin authorization.
- Generated OpenAPI asset served through swagger UI at `/docs/`.

## Where To Change Things

- `cmd/base-server/`: entrypoint, Wire graph, embedded OpenAPI assets.
- `api/protos/`: source protos; edit here for API changes.
- `api/gen/go/`: generated protobuf and Kratos bindings; do not edit.
- `internal/conf/`: config schema proto and generated Go.
- `internal/data/schema/`: Ent schema sources.
- `internal/data/ent/`: generated Ent code; do not edit.
- `internal/biz/`: use cases, auth, RBAC, repository interfaces.
- `internal/service/`: transport handlers implementing generated interfaces.
- `internal/server/`: HTTP/gRPC server wiring, middleware, response encoding, cron worker.
- `deploy/helm/base-server/`: Helm chart and embedded runtime config.

## Generated Code Boundaries

- `api/gen/go/**` comes from `api/protos/**` via `make api`.
- `cmd/base-server/assets/openapi.yaml` comes from proto via `make api` or `make openapi`.
- `cmd/base-server/wire_gen.go` comes from `make wire`.
- `internal/conf/conf.pb.go` comes from `make config`.
- `internal/data/ent/**` comes from `make ent`.

## High-Conflict Files

- `go.mod`, `go.sum`, `Makefile`, `Dockerfile`
- `configs/config.yaml`
- `api/protos/**`, `api/gen/go/**`
- `internal/conf/conf.proto`, `internal/conf/conf.pb.go`
- `internal/data/schema/**`, `internal/data/ent/**`
- `cmd/base-server/main.go`, `cmd/base-server/wire.go`, `cmd/base-server/wire_gen.go`
- `cmd/base-server/assets/openapi.yaml`
- `internal/server/**`, especially middleware and response encoding
- `deploy/helm/base-server/**`
- `docs/project_profile.md`, `docs/architecture/**`, `docs/planning/**`, `docs/progress.md`, `docs/worklog.md`

## Commands

```bash
make init
make api
make config
make wire
make ent
make all
make build
go build ./...
go test ./...
go run ./cmd/base-server -conf ./configs
docker build -t base-server .
helm template base-server ./deploy/helm/base-server
```

## Runtime

- Binary entrypoint: `cmd/base-server/main.go`.
- Config flag: `-conf`, default `./configs`.
- Docker expects config at `/data/conf`.
- HTTP listens from `server.http.addr`.
- gRPC is configured but not started unless `newApp` includes `gs`.
- Active frontend/backend API prefix on this branch is `/basic-api/*`.
- Do not use `app/*/service`, `make run-*`, or `deploy/configs/*` from the microservice branch while working on `master`.

## Full-Stack Feature Rule

- Treat `base-server/master` + `../vben-admin/main` as one complete monolithic product line for feature work.
- When adding, modifying, or refactoring a feature, inspect the paired frontend impact before editing: API modules, generated clients, views, routes, stores, permissions, upload/SSE paths, and visible user behavior.
- If a backend change affects proto, OpenAPI, fields, errors, auth, RBAC, pagination, upload, or SSE behavior, plan the matching frontend update in `../vben-admin` on the same version line.
- If the frontend does not need changes, state the reason in the completion report.
- Do not apply microservice API prefixes or service-split assumptions unless both repos are on `monorepo`.

## Git Policy

- Check `git status --short --branch`, current branch/ref, and worktree list before editing.
- Do not overwrite existing dirty files unless the task explicitly targets them.
- Keep commits aligned with a single task boundary.
- Do not commit generated artifacts, logs, build outputs, or secrets unless explicitly required.
- Push only when the user asks.

## Worktree And Parallelism

- Use parallel work only for disjoint write paths.
- Do not parallelize proto, Ent schema, config schema, Wire, or deployment config changes with dependent implementation work.
- Merge one child result at a time and verify after each merge.

## Human Escalation

Ask before changing:

- Public API, proto, or OpenAPI contract
- Config schema or deployment values
- Database schema or migration behavior
- Auth, JWT, Casbin, or middleware policy
- Build, Docker, or Helm release behavior
- Any real secret or credential handling

## Repo-Specific Risks

- `configs/config.yaml` and Helm ConfigMap templates contain live-looking config values; treat them as placeholders and do not copy real credentials into git.
- There are no `*_test.go` files in the scanned repository snapshot.
- `logs/` is untracked and should remain outside source control.

## Completion Report

Report:

- Goal result
- Active version line and frontend/backend impact
- Changed files
- Verification run
- Documentation updated
- Open risks
- Next recommended action
