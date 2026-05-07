# Site Message Publish Fix

## Goal

Fix the SQL error returned by the manager-side "publish now" action when a draft site message is updated to the published state.

## Classification

- Level 1 local change
- Active version line: `base-server/master` + `vben-admin/main`
- Backend impact: yes
- Frontend impact: no code or contract change is expected

## Scope

- `internal/data/base.go`
- `internal/data/site_message_test.go`
- `go.mod`
- `docs/progress.md`
- `docs/worklog.md`

## Assumption

- The failure is caused by clearing and setting the same nullable timestamp column in a single Ent update mutation.

## Tasks

- [x] Inspect the site-message update path for draft, scheduled, and published transitions.
- [x] Update the repo mutation logic so nullable lifecycle timestamps are only cleared when they are not set again in the same mutation.
- [x] Add a regression test for updating an existing draft to `scheduled` and `published`.
- [x] Run targeted backend verification and record the results.

## Verification

- `go test ./internal/data -run TestCreateSiteMessage_UpdateDraftTransitions -count=1`
- `go test ./...`

## Result

- The draft-to-publish and draft-to-schedule update paths now avoid duplicate assignments on nullable lifecycle timestamp columns.
- Regression coverage is in `internal/data/site_message_test.go`.
