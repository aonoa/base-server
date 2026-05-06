# Current Architecture

## Major Modules

- This architecture describes the monolithic `master` branch, not the `monorepo` microservice branch.
- `cmd/base-server/`: process entrypoint, config loading, Wire graph, embedded OpenAPI asset
- `api/protos/base_api/v1/`: source protobuf API definitions
- `api/gen/go/`: generated protobuf/Kratos bindings
- `internal/conf/`: config schema source and generated config types
- `internal/data/schema/`: Ent schema sources
- `internal/data/ent/`: generated Ent client and models
- `internal/biz/`: use cases, repositories interfaces, auth logic, menu/tree logic
- `internal/service/`: transport handlers for HTTP/gRPC
- `internal/server/`: server wiring, middleware, response encoding, cron worker
- `internal/log/`: logger setup
- `internal/tools/`: utility helpers
- `internal/types/`: shared type aliases or DTO helpers
- `pkg/`: shared helpers for transport/context handling
- `deploy/helm/base-server/`: Helm chart and embedded runtime config

## Data Flow

1. `main.go` loads YAML config into `conf.Bootstrap`.
2. Wire assembles logger, data, biz, service, and server providers.
3. `internal/data` opens PostgreSQL, initializes Ent, and creates schema on startup.
4. `internal/biz` coordinates auth, token generation, RBAC, and business rules.
5. `internal/service` implements the generated protobuf service interfaces.
6. `internal/server` registers HTTP, gRPC, middleware, swagger UI, and cron worker.
7. The OpenAPI asset is generated from proto and exposed through `/docs/`.

## External Integrations

- PostgreSQL via `pgx`
- Casbin for authorization
- JWT HS256 for authentication/session handling
- Ent ORM for persistence
- Kratos HTTP/gRPC transport and middleware
- Swagger UI for the OpenAPI asset
- Helm and Docker for deployment packaging

## Public Contracts

- HTTP routes are defined in `api/protos/base_api/v1/*.proto`
- Config schema is defined in `internal/conf/conf.proto`
- OpenAPI output is generated to `cmd/base-server/assets/openapi.yaml`
- Helm config embeds runtime config and Casbin policy/model files
- The active API prefix for this branch is `/basic-api/*`; microservice prefixes belong to `monorepo`.

## Risks And Constraints

- Generated code is a write-once output layer and should not be edited by hand
- `configs/config.yaml` and Helm templates currently contain placeholder-like secrets and live credentials
- `internal/data/data.go` auto-creates schema on startup, so data-model changes affect runtime behavior immediately
- gRPC is wired but not started, so the effective public surface is HTTP-first
