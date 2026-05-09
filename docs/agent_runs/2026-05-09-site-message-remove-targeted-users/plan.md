# Site Message Remove Targeted Users

## Goal

Remove the targeted-user compose flow from site-message management and normalize new site-message creation to all-user delivery only.

## Classification

- Level 2 feature increment
- Active version line: `base-server/master` + `vben-admin/main`
- Backend impact: yes
- Frontend impact: yes

## Git And Execution Notes

- `base-server` and `vben-admin` both already contain unrelated in-progress changes from the active site-message iteration.
- Automatic parallel worktree execution is skipped for this iteration because the worktrees are not clean and the touched files overlap the current WIP.
- This iteration is executed sequentially in the main worktrees.

## Requirements Update

- Site-message management no longer offers a "specified users" audience mode.
- New drafts, scheduled messages, and immediate publishes are treated as all-user messages only.
- Backend must ignore targeted-user audience input for new mutations so non-UI callers cannot continue using that product path accidentally.
- Existing historical records are not migrated destructively in this iteration.

## Scope

- `internal/biz/base.go`
- `internal/data/base.go`
- `internal/data/site_message_test.go`
- `docs/progress.md`
- `docs/worklog.md`
- paired frontend site-message API wrapper, management page, and iteration docs in `vben-admin/main`

## Tasks

- [x] Record the requirement change and execution constraints.
- [x] Remove targeted-user handling from backend normalization and persistence for new mutations.
- [x] Add or update regression coverage for all-user-only site-message creation.
- [x] Remove targeted-user UI and payload shaping from the paired frontend manager page.
- [x] Run targeted verification and update project docs.

## Verification

- [x] `go test ./internal/data`
- [x] `go test ./...`
- [x] `pnpm -C /home/mini/OpenSource/framework/vben-admin --filter @vben/web-antd exec vue-tsc --noEmit` (still fails on unrelated pre-existing frontend repo issues in copilot, generated request, dashboard, and existing modal typings)

## Risks

- Historical published records may still contain targeted-user metadata from earlier behavior; this iteration leaves them intact to avoid destructive data rewrite.
- Frontend full typecheck is known to have unrelated existing failures outside the site-message files.

## Result

- Backend site-message creation now forces all-user delivery for new mutations and ignores legacy targeted-user input.
- Data-layer regression coverage now proves publish and schedule flows expand to all active users.
- The paired frontend manager page no longer exposes targeted-user compose controls and now always submits all-user messages.
