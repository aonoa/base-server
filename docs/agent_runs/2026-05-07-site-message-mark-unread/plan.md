# Site Message Mark Unread

## Goal

Allow a user to mark an already-read site message as unread so the unread badge can continue reminding them until they handle the message.

## Classification

- Level 2 feature increment
- Active version line: `base-server/master` + `vben-admin/main`
- Backend impact: add a user-scoped read-state mutation endpoint
- Frontend impact: add a "mark unread" action in the site-message inbox

## Scope

- `api/protos/base_api/v1/base.proto`
- Generated backend API/OpenAPI artifacts
- `internal/biz/base.go`
- `internal/biz/auth.go`
- `internal/data/base.go`
- `internal/service/base.go`
- `internal/data/site_message_test.go`
- Paired `vben-admin/main` OpenAPI/client wrapper and inbox view
- Backend/frontend project docs

## Tasks

- [x] Add the mark-unread RPC and route to the monolith API contract.
- [x] Implement the usecase and repository mutation by setting `is_read=false` and resetting `read_time`.
- [x] Add a regression test for read-to-unread receipt state and unread count.
- [x] Add the frontend API wrapper and inbox action for read records.
- [x] Run targeted verification and update project docs.

## Verification

- `go test ./internal/data -run 'TestCreateSiteMessage_UpdateDraftTransitions|TestMarkSiteMessageUnread' -count=1`
- `go test ./...`
- `pnpm -C /home/mini/OpenSource/framework/vben-admin --filter @vben/web-antd run typecheck` (fails on pre-existing repo issues outside the touched files)

## Result

- Users can now toggle a received message back to unread from the inbox.
- The unread badge count is recalculated from the backend after the toggle, so the reminder persists again.
