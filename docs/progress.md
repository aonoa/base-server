# Progress

## Current Status

- Repository scan completed
- Bootstrap documents created for project state, architecture, boundaries, verification, and risks
- Root agent notes updated to point at the new docs
- Project-specific skill has been generated in the repository at `.codex/skills/base-server/`
- Root coordination docs now separate monolithic `master/main` and microservice `monorepo/monorepo` version lines
- Full-stack feature work is now documented as the default rule for changes in the paired frontend/backend version line
- Logger setup now lives under `internal/logx/` and keeps structured JSON output
- `configs/config.yaml` is protected by a repo-local clean filter and pre-commit hook so local secrets are redacted before entering Git
- Site-message backend is now implemented on `master`, including inbox/publish/management APIs, unread/read-state persistence, lifecycle states, lazy scheduled promotion, and generated OpenAPI output
- Backend startup/usecase flow now ensures the hidden `/messages` inbox row and the visible `/system/site-message` manager row exist in `sys_menu`
- Backend startup/usecase flow now also normalizes those persisted menu rows onto separate frontend component paths for inbox and manager pages
- Backend startup now also self-heals the `site-message-manage` API permission resource onto `admin`/`root` only, so menu access and management-interface access are both enforced on the backend side
- `GetUserInfo` now returns the real repeated role list so frontend role checks and backend-menu mode remain aligned
- Role-menu bootstrap now backfills `root`/`admin` with the manager menu while keeping ordinary users on the hidden inbox route only
- Draft-to-publish and draft-to-schedule updates now avoid duplicate nullable timestamp assignments, so "publish now" no longer hits the `published_time` SQL error
- Site-message receipts now support toggling an already-read message back to unread, which resets the badge count again until the user handles it
- Site-message creation now normalizes new drafts, scheduled tasks, and publishes to all-user delivery only, and targeted-user input is ignored on the backend
- Site-message persistence no longer keeps `receiver_type` or `receiver_ids`, and backend startup now drops those legacy columns from existing databases
- Site-message API contracts no longer expose `receiverType` or `receiverIds`, and the paired OpenAPI/client outputs have been regenerated for the monolith line

## Active Blockers

- No bootstrap blocker inside the repo
- Untracked `logs/` directory remains outside source control
- Frontend full typecheck on the paired `vben-admin/main` repo still has unrelated pre-existing errors, so end-to-end compile verification remains partially blocked outside this repo

## Next Planned Iteration

- Keep future implementation work document-driven
