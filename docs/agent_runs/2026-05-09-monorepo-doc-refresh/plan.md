# Monorepo Doc Refresh

## Goal

Refresh outdated coordination docs and add missing project-level documentation for the active `monorepo` version line.

## Classification

- Level 1 docs-only change
- Active version line: `base-server/monorepo` + `vben-admin/monorepo`
- Backend impact: docs only
- Frontend impact: docs only

## Scope

- Root coordination docs:
  - `REPO_STRUCTURE.md`
  - `STARTUP_AND_INTEGRATION.md`
- Backend repo docs:
  - `README.md`
  - `docs/README.md`
  - `docs/monorepo-overview.md`

## Tasks

- [x] Identify outdated or missing monorepo documentation.
- [x] Remove stale root-doc references to temporary checkout state.
- [x] Add a backend docs index and monorepo overview.
- [x] Link the new docs from the repository README.

## Verification

- [x] Manual consistency review of updated paths, branch names, service names, and command references
- [x] `go test ./...`
- [ ] `pnpm --dir /home/mini/OpenSource/framework/vben-admin --filter @vben/web-antd exec vue-tsc --noEmit` (tracked in the paired frontend repo docs as an existing failure set)

## Notes

- Root `framework/` is not a Git repository; root coordination docs are editable but not commit-tracked as part of either child repository.
