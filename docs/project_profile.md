# Project Profile

## Identity

- Repository: `base-server`
- Project type: Go/Kratos backend service
- Stack: Go 1.24, Kratos v2, Ent ORM, Casbin, JWT, PostgreSQL
- Main transport: HTTP on `:8000`
- Secondary transport: gRPC on `:9000` is wired but not started in `cmd/base-server/main.go`

## Current State

- Branch: `master`
- Baseline commit: `c988100f8955b794bd92d61880c68fa8eded4cd8`
- Remote: `origin` points to `git@github.com:aonoa/base-server.git`
- Worktree: single worktree at the repo root
- Git status: clean except for untracked `logs/`
- Version line: this document describes the monolithic backend on `master`; the microservice backend lives on the separate `monorepo` branch.

## Sibling Version Line

- Monolithic pairing: `base-server/master` + `vben-admin/main`
- Microservice pairing: `base-server/monorepo` + `vben-admin/monorepo`
- Shared coordination docs live one directory above this repo in `../REPO_STRUCTURE.md` and `../STARTUP_AND_INTEGRATION.md`
- Do not mix `/basic-api/*` monolith APIs with `/auth-api/*`, `/user-api/*`, `/admin-api/*`, or `/common-api/*` microservice APIs in the same branch line.
- Feature additions, modifications, and refactors must be evaluated as full-stack work against the paired frontend branch, even when the code change is ultimately backend-only.

## Goal

Provide an admin-oriented backend with HTTP APIs, auth, RBAC, data access, jobs, and generated OpenAPI/Ent artifacts.

## Target Users

- Backend maintainers
- API consumers
- Frontend integrators using the generated OpenAPI contract
- Operators deploying the Helm chart or Docker image

## Core Workflows

- Edit proto or schema sources, then regenerate generated code
- Run the service from `cmd/base-server/main.go` with a config directory
- Build a binary or Docker image for release
- Sync OpenAPI output to the frontend generator
- Check paired frontend API modules, types, routes, views, stores, and permissions whenever backend contracts or behavior change
- Use Helm templates for deployment

## Non-Goals

- No hand-editing of generated protobuf, Ent, or Wire output
- No storing real secrets in repo-managed config
- No parallel service rewrite without updating proto, data, and middleware together

## Success Criteria

- API and generated code stay in sync
- Config schema matches runtime YAML
- Auth, RBAC, and logging behavior remain consistent with the documented contract
- Changes are verified with the narrowest relevant command set before handoff
- Backend and frontend impact are both reported for feature work
