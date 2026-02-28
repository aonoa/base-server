# internal/conf/ — config schema (proto)

## Source vs generated
- Edit: `internal/conf/conf.proto`
- Do not edit: `internal/conf/conf.pb.go`

## What’s in the schema
- `Bootstrap` is the root config message.
- Key sections: `server`, `logger`, `data`, `auth`, `menus`, `job`, `llm`.

## Generate
- From repo root: `make config`

## If config load fails
- Mismatched YAML keys/types vs `conf.proto` → update YAML or regenerate and fix structs.
- `cmd/base-server/main.go` panics on load/scan errors; look there first.

## Runtime
- `cmd/base-server/main.go` loads YAML into `conf.Bootstrap` via Kratos config.
- Keep `configs/config.yaml` aligned with `conf.proto`.
