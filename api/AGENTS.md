# api/ — protobuf + generated bindings

## Source vs generated
- Edit: `api/protos/**.proto`
- Do not edit: `api/gen/go/**` (regenerated)

## What’s in here
- `api/protos/auth/service/v1/` — auth service contracts
- `api/protos/user/service/v1/` — user service contracts
- `api/protos/admin/service/v1/` — admin service contracts
- `api/protos/common/service/v1/` — common service contracts
- `api/gen/go/**` — generated Go bindings
- `api/openapi/**` — generated OpenAPI output

## Generate
- From repo root:
  - `make api` (pb.go + kratos http/grpc + errors + OpenAPI)
  - `make errors` (errors only)
  - `make openapi` (OpenAPI only; output goes to `api/openapi/`)

## Protoc plugins used
- `protoc-gen-go` → `*.pb.go`
- `protoc-gen-go-grpc` → `*_grpc.pb.go`
- `protoc-gen-go-http` → `*_http.pb.go`
- `protoc-gen-go-errors` → `*_errors.pb.go`
- `protoc-gen-openapi` → `api/openapi/openapi.yaml`

## Common workflow
1. Edit the owning proto under `api/protos/*/service/v1/*.proto`.
2. Run `make api`.
3. Update handlers in the owning split service under `app/*/service/internal/service/`.
4. Re-check generated output under `api/gen/go/` and `api/openapi/`.

## Boundaries to keep clean
- Never hand-edit anything under `api/gen/go/`.
- If you need a new endpoint, add it in proto first; do not add ad-hoc HTTP routes.
