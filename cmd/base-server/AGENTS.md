# cmd/base-server/ — bootstrap + Wire + OpenAPI assets

## Entry
- `main.go` loads config from `-conf` (default `./configs`) and builds the app via `wireApp(...)`.
- gRPC server is wired but currently not started (`gs` is commented in `newApp`).

## Config loading
- Reads from directory passed to `-conf` via `kratos/config/file`.
- YAML is scanned into `conf.Bootstrap` (see `internal/conf/conf.proto`).

## Wire
- Edit DI graph in `wire.go`.
- Generate: `make wire` (runs `go generate ./...`).
- Do not edit: `wire_gen.go`.

## If wire breaks
- Check provider sets in `cmd/base-server/wire.go`.
- Regenerate with `make wire` (do not run wire output edits manually).

## OpenAPI asset
- `assets/openapi.yaml` is generated from protos.
- Regenerate via `make api` / `make openapi`.
- Served at `/docs/` by `internal/server/http.go` (swagger-ui).
