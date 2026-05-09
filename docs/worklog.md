# Worklog

## 2026-05-06

- Scanned Git state, remotes, worktree, manifests, CI, and the repository layout.
- Confirmed `base-server` is a separate repository from `vben-admin` and bootstrapped it independently.
- Documented source boundaries for proto, config, Ent, service, server, deployment, and shared helper layers.
- Recorded the repository’s sensitive config and Helm placeholders as risk items.
- Deferred source edits; only bootstrap docs and agent metadata were updated.
- Generated the project-specific Codex skill in the repository at `.codex/skills/base-server/`.
- Removed the user-level copy at `~/.codex/skills/base-server/`.
- Refreshed root coordination docs to separate monolithic `master/main` and microservice `monorepo/monorepo` version lines.
- Added the rule that feature additions, modifications, and refactors must assess both backend and paired frontend impact within the active version line.
- Migrated the logger adapter to `internal/logx/` and kept JSON structured output.
- Added a repo-local clean filter and pre-commit hook to redact `configs/config.yaml` into `CHANGE_ME_*` placeholders at commit time while preserving local worktree values.

## 2026-05-07

- Added site-message domain support on `master`: Ent schema, repo/usecase/service methods, proto/OpenAPI contract, and generated outputs.
- Added monolithic `/basic-api/notice/*` endpoints for inbox list, unread count, single/all read mutation, publish, management list, recall, and delete-pending queries.
- Kept Casbin API-group mapping coarse for v1 by attaching the new notice endpoints to the existing `default` API group, while leaving publish/list management protected in backend usecase logic.
- After user clarification, changed the message-center entry plan from a frontend supplemental route to backend-managed database menu data.
- Added backend logic that ensures the hidden `/messages` route row exists in `sys_menu` and that role-menu data preserves access to that route for backend-menu mode navigation.
- After the follow-up product change, added a second built-in menu row at `/system/site-message` for management users only and updated role-menu normalization so `root`/`admin` receive it while ordinary roles do not.
- Extended site-message persistence with lifecycle states (`draft`, `scheduled`, `published`, `recalled`) and lazy scheduled promotion so pending records can be managed without a separate scheduler.
- Fixed `GetUserInfoReply.roles` from a singular object to a repeated role list and populated it from the user-role relation inside the service flow.
- Fixed the draft update path so immediate publish no longer assigns `published_time` twice; added a regression test that covers draft-to-publish and draft-to-schedule transitions.
- Added a mark-unread endpoint for site-message receipts, with backend and frontend wiring plus regression coverage for read-to-unread toggling.
- Updated the built-in site-message menu bootstrap so inbox and manager routes point to separate frontend component files, and older persisted menu rows are repaired on startup.
- Tightened the site-message management authority boundary by assigning the management APIs to a dedicated backend API resource group and auto-binding that resource to `admin`/`root` during startup.

## 2026-05-09

- Recorded the product decision to remove the targeted-user compose path from site-message management and kept the iteration sequential because both repos were already dirty in the same feature area.
- Changed backend site-message creation normalization so new drafts, scheduled messages, and immediate publishes are always stored as all-user messages.
- Cleared targeted-user request IDs inside the site-message usecase so non-UI callers cannot keep using the removed product path by sending legacy payloads.
- Kept historical targeted-user records intact rather than rewriting old data in place.
- Updated site-message data-layer regression tests to seed active users and verify that new publish/schedule flows fan out to all active users instead of honoring targeted-user payloads.
- Removed `receiver_type` and `receiver_ids` from the site-message Ent schema and regenerated the Ent artifacts.
- Added an idempotent startup cleanup SQL step that drops the two legacy columns from `sys_site_message` in existing PostgreSQL databases.
- Restarted the backend on the monolith line and verified the live `test1.public.sys_site_message` table no longer contains the dropped columns.
- Deleted the legacy `receiverType` and `receiverIds` fields from the site-message proto contract, regenerated Go/OpenAPI outputs, and synced the paired frontend generated client.
- Simplified the service mapping and site-message tests so the removed API fields are no longer referenced anywhere in runtime code.
