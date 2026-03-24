# pkg/data/ — shared Ent layer

## Source vs generated
- Edit schemas: `pkg/data/schema/*.go`
- Edit templates: `pkg/data/template/*.tmpl`
- Do not edit: `pkg/data/ent/**` (generated)

## Generate Ent
- From repo root: `make ent`

## If Ent changes don’t show up
- Make sure you edited `pkg/data/schema/` (not `pkg/data/ent/`).
- Re-run `make ent` and recompile.

## Usage
- Split services import shared generated Ent packages from `pkg/data/ent`.
- Service-specific repositories stay under `app/*/service/internal/data/`.
