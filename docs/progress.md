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
- `GetUserInfo` now returns the real repeated role list so frontend role checks and backend-menu mode remain aligned
- Role-menu bootstrap now backfills `root`/`admin` with the manager menu while keeping ordinary users on the hidden inbox route only
- Draft-to-publish and draft-to-schedule updates now avoid duplicate nullable timestamp assignments, so "publish now" no longer hits the `published_time` SQL error
- Site-message receipts now support toggling an already-read message back to unread, which resets the badge count again until the user handles it

## Active Blockers

- No bootstrap blocker inside the repo
- Untracked `logs/` directory remains outside source control
- Frontend full typecheck on the paired `vben-admin/main` repo still has unrelated pre-existing errors, so end-to-end compile verification remains partially blocked outside this repo

## Next Planned Iteration

- Keep future implementation work document-driven
