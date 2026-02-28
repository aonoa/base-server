# api/ — protobuf + generated bindings

## Source vs generated
- Edit: `api/protos/**.proto`
- Do not edit: `api/gen/go/**` (regenerated)

## What’s in here
- `api/protos/base_api/v1/` — services + messages
  - `base.proto` — main service + core types
  - `sse.proto` — SSE service
  - `upload.proto` — file upload service
  - `error_reason.proto` — error enums/reasons
- `api/protos/base_api/options/options.proto` — custom proto options
- `api/gen/go/base_api/**` — generated Go (pb + http + grpc + errors)

## Generate
- From repo root:
  - `make api` (pb.go + kratos http/grpc + errors + OpenAPI)
  - `make errors` (errors only; rarely needed standalone)
  - `make openapi` (OpenAPI only; output goes to `cmd/base-server/assets/`)

## Protoc plugins used (Makefile)
- `protoc-gen-go` → `*.pb.go`
- `protoc-gen-go-grpc` → `*_grpc.pb.go`
- `protoc-gen-go-http` → `*_http.pb.go` (Kratos HTTP transport bindings)
- `protoc-gen-go-errors` → `*_errors.pb.go`
- `protoc-gen-openapi` → `cmd/base-server/assets/openapi.yaml`

## Outputs to know
- `api/gen/go/base_api/v1/*_grpc.pb.go` — gRPC stubs
- `api/gen/go/base_api/v1/*_http.pb.go` — Kratos HTTP bindings
- `api/gen/go/base_api/v1/*_errors.pb.go` — error helpers

## Common “I changed proto, now what?”
1. Change `api/protos/base_api/v1/*.proto`.
2. Run `make api`.
3. Update handlers in `internal/service/`.
4. If you changed HTTP annotations/options, re-check swagger at `/docs/`.

## Boundaries to keep clean
- Never hand-edit anything under `api/gen/go/` (regen will overwrite).
- If you need a new endpoint: add it in proto first; do not add ad-hoc HTTP routes.
