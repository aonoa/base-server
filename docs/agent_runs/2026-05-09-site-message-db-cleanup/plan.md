# Site Message DB Cleanup

## Goal

Remove the legacy `receiver_type` and `receiver_ids` storage columns from the site-message database layer and keep new site-message writes on all-user delivery only.

## Classification

- Level 3 architecture/schema change
- Active version line: `base-server/master` + `vben-admin/main`
- Backend impact: yes
- Frontend impact: no direct behavior change expected

## Execution Notes

- This is a schema/storage change, so it is handled sequentially in the main worktree.
- The repository is already dirty from the in-progress site-message iteration, so no worktrees or parallel child tasks are used here.

## Requirements Update

- `sys_site_message` should no longer persist `receiver_type` or `receiver_ids`.
- New site-message writes continue to target all active users only.
- The backend should perform a one-time startup cleanup so old databases actually drop the legacy columns.
- Historical records do not need a data rewrite beyond column removal.

## Scope

- `internal/data/schema/site_message.go`
- `internal/data/data.go`
- `internal/data/base.go`
- `internal/data/site_message_test.go`
- `deploy/sql/pg_dump.sql`
- generated Ent artifacts under `internal/data/ent/**`

## Tasks

- [x] Record the schema cleanup requirement and risk posture.
- [x] Remove the legacy site-message columns from the Ent schema and backend persistence code.
- [x] Add startup cleanup SQL so existing databases drop the columns.
- [x] Regenerate Ent artifacts and refresh deployment SQL snapshots. (`deploy/sql/pg_dump.sql` did not contain a site-message table snapshot, so no file edit was required there.)
- [x] Run targeted verification and update project docs.

## Verification

- [x] `make ent`
- [x] `go test ./internal/data`
- [x] `go test ./...`
- [x] Restart backend and verify `information_schema.columns` no longer contains `receiver_type` or `receiver_ids`

## Risks

- Startup cleanup touches live database structure and must be idempotent.
- Existing historical message rows will lose the old audience metadata once the columns are dropped.

## Result

- The Ent schema and generated artifacts no longer expose `receiver_type` or `receiver_ids`.
- Backend startup now executes an idempotent `ALTER TABLE ... DROP COLUMN IF EXISTS ...` cleanup for `sys_site_message`.
- The live monolith database was verified after restart: `receiver_type` and `receiver_ids` are gone from `public.sys_site_message`.
