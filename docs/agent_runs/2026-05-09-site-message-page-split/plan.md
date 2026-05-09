# Site Message Page Split

## Goal

Remove the user-facing summary field from site-message flows and split the inbox and management experiences into two different pages with separate component files.

## Classification

- Level 2 feature increment
- Active version line: `base-server/master` + `vben-admin/main`
- Backend impact: adjust built-in menu component mapping for the paired frontend pages
- Frontend impact: split the existing shared site-message page and remove summary input/display semantics

## Scope

- `internal/biz/base.go`
- `docs/progress.md`
- `docs/worklog.md`
- Paired `vben-admin/main` site-message views, API wrapper, layout notification mapping, and project docs

## Requirements Update

- Inbox route `/messages` and manager route `/system/site-message` must render from different Vue files.
- The inbox page remains user-facing read/unread handling only.
- The manager page remains admin/root-only publish and record management only.
- Summary is no longer a user-visible field in compose forms, cards, or bell notifications.
- No API contract or persistence migration is required for this iteration; backend `category` may remain as an internal/defaulted field.

## Tasks

- [x] Update backend iteration docs for the new split-page requirement.
- [x] Point built-in inbox/manage menu rows to separate frontend component paths.
- [x] Split the frontend shared page into separate inbox/manage files and shared helpers.
- [x] Remove summary validation, inputs, and display text from the frontend flow.
- [x] Run targeted verification and update project docs in both repos.

## Verification

- [x] `go test ./...`
- [x] `pnpm -C /home/mini/OpenSource/framework/vben-admin --filter @vben/web-antd exec vue-tsc --noEmit` (still fails on unrelated pre-existing frontend repo issues outside the touched site-message files)

## Risks

- Existing menu rows are persisted in `sys_menu`; startup normalization must update old component paths rather than only creating missing rows.
- Frontend targeted typecheck still has known unrelated failures, so compile-level verification on `vben-admin/main` remains partially blocked by pre-existing repo issues.

## Result

- Built-in site-message menu bootstrap now points `/messages` to `/_core/messages/inbox` and `/system/site-message` to `/_core/messages/manage`, and upgrades older persisted menu rows to those component paths during startup.
- The paired frontend now renders inbox and management from different Vue files, and the old `index.vue` compatibility page has been removed.
- User-facing summary input and display were removed from compose, cards, and bell previews; backend `category` stays internal and defaults to `system`.
- Follow-up requirement tightening moved the authority boundary fully back to the backend: frontend page code no longer checks roles for station-message management visibility, and startup now binds the management APIs to a dedicated `site-message-manage` API resource granted only to `admin`/`root`.
