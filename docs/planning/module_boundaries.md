# Module Boundaries

| Area | Paths | Responsibility | Notes |
| --- | --- | --- | --- |
| Entrypoint | `cmd/base-server/` | Program startup, config loading, Wire graph, embedded OpenAPI | High-conflict when transport or dependency graph changes |
| Protobuf source | `api/protos/base_api/v1/` | HTTP/gRPC API contract source | Edit here first for API changes |
| Generated API | `api/gen/go/` | Generated protobuf and transport bindings | Do not edit by hand |
| Config schema | `internal/conf/conf.proto` | Runtime config contract | Keep YAML aligned with this file |
| Domain/business | `internal/biz/` | Use cases, auth, RBAC, tree-building, repository interfaces | Owns orchestration logic |
| Data access | `internal/data/` and `internal/data/schema/` | Ent client, repositories, schemas, cache setup | Schema changes require regeneration |
| Transport handlers | `internal/service/` | HTTP/gRPC service implementations | Must track proto changes |
| Server wiring | `internal/server/` | Middleware, HTTP/gRPC setup, response encoding, cron worker | Changing middleware changes runtime policy |
| Logging | `internal/logx/` | Zap/Kratos logger adapter, JSON encoder, stdout/file sink setup | Logger changes affect runtime observability |
| Shared helpers | `internal/tools/`, `internal/types/`, `pkg/` | Utility functions and cross-layer helpers | Keep these lightweight and reusable |
| Deployment | `deploy/helm/base-server/` | Helm chart, configmap, service, deployment | Contains sensitive config placeholders |
| Runtime config | `configs/config.yaml` | Local default configuration | Must match `conf.proto` |

## Boundary Rules

- Change proto first for API surface work, then regenerate and update service implementations.
- Change config schema and YAML together.
- Change Ent schemas in `internal/data/schema/`, then regenerate `internal/data/ent/`.
- Keep `cmd/base-server/assets/openapi.yaml` and `wire_gen.go` strictly generated.
- Keep secrets out of repo-managed config whenever possible.
- Treat feature additions, modifications, and refactors as full-stack by default: inspect the paired `../vben-admin` branch for API clients, pages, routes, stores, permissions, upload/SSE paths, and visible behavior.
- If a backend change does not require frontend changes, document the reason in the handoff.
- Keep this full-stack analysis inside the active version line; `master` pairs with `vben-admin/main`, while `monorepo` pairs with `vben-admin/monorepo`.
