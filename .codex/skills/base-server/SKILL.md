---
name: base-server
description: Use when working on the base-server Go/Kratos backend, including APIs, Ent data models, config, auth middleware, deployment, docs, and verification.
---

# base-server

## Core Principle

This project uses document-driven iterative development. Repository docs are the source of truth; chat is temporary. Requirements, decisions, assumptions, open questions, task boundaries, verification commands, and iteration results must be written to project docs before they are treated as durable state.

## Full-Stack Product Rule

Treat `base-server/master` + `/home/mini/OpenSource/framework/vben-admin/main` as one complete monolithic product line for feature work. When adding, modifying, or refactoring a feature, inspect the paired frontend impact before editing and report both backend and frontend impact at handoff.

If backend API, auth, RBAC, fields, pagination, upload, SSE, errors, or OpenAPI behavior changes, plan the matching frontend client/type/view/route/store update in the paired `vben-admin` repo on the same version line. If no frontend change is required, state why.

Do not mix in microservice branch assumptions or API prefixes unless both repos are on `monorepo`.

## Read First

- `/home/mini/OpenSource/framework/AGENTS.md`
- `/home/mini/OpenSource/framework/REPO_STRUCTURE.md`
- `/home/mini/OpenSource/framework/STARTUP_AND_INTEGRATION.md`
- `AGENTS.md`
- `docs/project_profile.md`
- `docs/architecture/current_architecture.md`
- `docs/architecture/tech_stack.md`
- `docs/planning/module_boundaries.md`
- `docs/planning/verification.md`
- `docs/planning/risk_register.md`
- `docs/progress.md`
- `docs/worklog.md`

## Module Boundaries

- These rules describe the monolithic `master` branch. The microservice backend lives on `monorepo`.
- `cmd/base-server/`: entrypoint, Wire graph, embedded OpenAPI
- `api/protos/`: source API contract
- `api/gen/go/`: generated protobuf and Kratos bindings
- `internal/conf/`: config schema source and generated types
- `internal/data/schema/`: Ent schema sources
- `internal/data/ent/`: generated Ent output
- `internal/biz/`: use cases, auth, RBAC, repository interfaces
- `internal/service/`: transport handlers
- `internal/server/`: server wiring, middleware, response encoding, cron worker
- `deploy/helm/base-server/`: Helm deployment and runtime config

## Shared High-Conflict Files

- `go.mod`
- `go.sum`
- `Makefile`
- `Dockerfile`
- `configs/config.yaml`
- `api/protos/**`
- `api/gen/go/**`
- `internal/conf/conf.proto`
- `internal/conf/conf.pb.go`
- `internal/data/schema/**`
- `internal/data/ent/**`
- `cmd/base-server/main.go`
- `cmd/base-server/wire.go`
- `cmd/base-server/wire_gen.go`
- `cmd/base-server/assets/openapi.yaml`
- `internal/server/**`
- `deploy/helm/base-server/**`
- `docs/project_profile.md`
- `docs/architecture/**`
- `docs/planning/**`
- `docs/progress.md`
- `docs/worklog.md`

Only modify these when the active task explicitly authorizes it.

## Verification

See `docs/planning/verification.md` for the current command matrix.

Use the narrowest generation or verification command that covers the touched boundary. Regenerate before compiling when proto, config schema, Ent schema, or Wire sources changed.

Do not use microservice-only commands such as `make run-user`, `make run-auth`, `make run-admin`, `make run-common`, `make run-gateway`, or `make dev-env-up` unless the repo is on `monorepo`.

## Git Policy

- Check `git status --short --branch`, current branch/ref, and worktree list before editing.
- Do not overwrite existing dirty files unless the task explicitly targets them.
- Keep commits aligned with a single task boundary.
- Do not commit generated artifacts, logs, build outputs, or secrets unless explicitly required.
- Push only when the user asks.

## Worktree And Subagent Discipline

- Use parallel work only for disjoint write paths.
- Do not parallelize proto, Ent schema, config schema, Wire, or deployment config changes with dependent implementation work.
- Merge one child result at a time and re-run targeted verification after each merge.

## Human Escalation

Ask before changing:

- Public API, proto, or OpenAPI contract
- Config schema or deployment values
- Database schema or migration behavior
- Auth, JWT, Casbin, or middleware policy
- Build, Docker, or Helm release behavior
- Any real secret or credential handling

## Human Interrupts

- Pause the affected work immediately.
- Record the confirmed change in `docs/worklog.md` and any affected project docs.
- Re-check task boundaries, parallelism, and verification before resuming.

## Completion Report

Report:

- Goal result
- Active version line and frontend/backend impact
- Changed files
- Verification run
- Documentation updated
- Open risks
- Next recommended action
