# internal/data/ — Ent + repositories

## Source vs generated (Ent)
- Edit schemas: `internal/data/schema/*.go`
- Edit templates: `internal/data/template/*.tmpl`
- Do not edit: `internal/data/ent/**` (generated)

## Generate Ent
- From repo root: `make ent`

## If Ent changes don’t show up
- Make sure you edited `internal/data/schema/` (not `internal/data/ent/`).
- Re-run `make ent` and recompile.

## DB init + migrations
- `internal/data/data.go`:
  - opens DB using `conf.Data.Database.{Driver,Source}`
  - creates Ent client
  - runs `client.Schema.Create(...)` on startup (auto-migration)

## Caching (local)
- `internal/data/data.go` creates a Ristretto-backed `gocache` manager.

## Repo implementation
- `internal/data/base.go` implements `biz.BaseRepo` using Ent queries.
