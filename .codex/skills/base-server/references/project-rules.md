# Project Rules

These rules describe `base-server/master`, the monolithic backend. For the microservice backend, use `base-server/monorepo` together with `vben-admin/monorepo` and read the root coordination docs.

For feature additions, modifications, or refactors, treat `base-server/master` and `vben-admin/main` as one product. Inspect paired frontend impact before backend edits, and keep all contract/codegen/client updates inside this version line.

## Repository Areas

| Area | Paths | Notes |
| --- | --- | --- |
| Entrypoint | `cmd/base-server/` | Startup, config load, Wire, embedded OpenAPI |
| API source | `api/protos/` | Edit before generated API outputs |
| API generated | `api/gen/go/` | Do not hand-edit |
| Config | `internal/conf/` | `conf.proto` is source; `conf.pb.go` is generated |
| Data | `internal/data/schema/`, `internal/data/ent/` | Schema is source; Ent output is generated |
| Business | `internal/biz/` | Use cases and repo interfaces |
| Services | `internal/service/` | Generated interface implementations |
| Server | `internal/server/` | Middleware, response encoding, cron, transport wiring |
| Deployment | `deploy/helm/base-server/` | Helm chart and runtime config |

## Generation Map

- `make api`: proto to Go bindings and OpenAPI.
- `make config`: config proto to Go.
- `make wire`: Wire graph to `wire_gen.go`.
- `make ent`: Ent schema to generated client/models.
- `make all`: API, config, and generic generation.

## Verification Ladder

1. Run the relevant generation command.
2. Run `go test ./...`.
3. Run `make build` or `go build ./...`.
4. For deployment changes, run `helm template base-server ./deploy/helm/base-server`.
5. For runtime checks, run `go run ./cmd/base-server -conf ./configs` only when the config and dependencies are safe for the environment.

Microservice commands and paths such as `app/*/service`, `deploy/configs/*`, and `make run-*` are out of scope on `master`.

For full-stack feature verification, add the paired frontend command that covers the affected user flow. If no frontend verification is required, record why.

## Parallel Execution Policy

- Safe: independent docs or isolated implementation files with non-overlapping write paths.
- Unsafe: proto/config/schema/Wire/deployment changes parallelized with generated output or dependent handlers.
- Always verify after each merge or child-worktree integration.
