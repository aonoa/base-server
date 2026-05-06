# Tech Stack

## Selected Stack

| Area | Choice | Why |
| --- | --- | --- |
| Language | Go 1.24 | Matches `go.mod` and toolchain |
| Framework | Kratos v2 | Core transport, middleware, logging, and app composition |
| ORM | Ent | Generated data layer with schema-first workflow |
| AuthN | JWT HS256 | Used in middleware and token generation |
| AuthZ | Casbin | Policy-based authorization with DB-backed adapter |
| DB | PostgreSQL | Primary storage |
| Wire DI | Google Wire | Used in `cmd/base-server` |
| API spec | Protobuf + Kratos HTTP annotations | Source for HTTP and gRPC contracts |
| OpenAPI | Generated from proto | Feeds swagger UI and frontend sync |
| Deployment | Docker + Helm | Runtime packaging and cluster deployment |

## Rejected Alternatives

- Hand-written transport layers: the repo already uses generated proto bindings
- ORM-free SQL: the current repository is already built around Ent
- gRPC-only service: HTTP is the active transport and swagger UI is part of the workflow
- Static config-only auth: current implementation requires JWT and Casbin

## Operational Assumptions

- Go toolchain is `go1.24.4`
- `make` is the canonical generation/build entrypoint
- `configs/config.yaml` is the default runtime config directory input
- `cmd/base-server/assets/openapi.yaml` and `internal/data/ent/**` are generated outputs
- Helm chart config values are sensitive and should be treated as placeholders unless explicitly rotated
