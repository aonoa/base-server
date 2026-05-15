# Organization Scoped User Management

Date: 2026-05-13

## Scope

- Active version line: `base-server/monorepo` + `vben-admin/monorepo`.
- Change level: Level 2 feature increment.
- Goal: user management follows the current organization context. Default organization remains the all-users organization; a non-default current organization shows and manages only its members.

## Backend Impact

- `admin` owns `sys_organization` and `sys_user_organization`.
- Existing `/admin-api/v1/organizations/{organization_id}/members` is the source for organization-scoped visible users.
- Existing `user` APIs continue owning global user profile CRUD.
- No backend API contract change is planned unless implementation discovers a missing operation.

## Frontend Impact

- `views/system/user/**` should load visible rows from the current organization members endpoint.
- New users created while a concrete organization is active should be added to that organization membership after the global user is created.
- Role binding display and filters remain backed by admin user-role binding APIs.

## Verification

- Backend targeted tests: `go test ./app/admin/service/internal/...`
- Frontend targeted checks: run the narrowest available type/lint check for `apps/web-antd`; if known project-wide `vue-tsc` failures block it, record that result.

## Results

- Backend code change: none. Existing admin organization member APIs are reused.
- Frontend code change: user management now reads current organization members for list visibility; new users are explicitly added to the default organization and, when different, the current organization after global user creation; new user role bindings include the default organization's `default` role; removing a row in a non-default organization removes membership, while default organization keeps global delete behavior.
- Backend verification: `go test ./app/admin/service/internal/...` passed.
- Frontend verification: `pnpm exec vitest run apps/web-antd/src/views/system/user/helpers.test.ts --dom` passed with 6 tests.
- Frontend verification: `pnpm -F @vben/web-antd run typecheck` failed on known existing files listed in `vben-admin/docs/project/known-issues.md`; no reported error referenced the touched user-management files.

## Follow-up 2026-05-14

- Requirement refinement: newly created users must also receive the default organization's `default` role as a backend rule, not only from the current frontend add-user modal.
- Backend change plan:
  - Keep frontend behavior unchanged as a caller-side redundancy.
  - In `admin`, merge the default organization's `default` role into explicit user-role upserts.
  - In `admin` read/projection paths, treat users without explicit bindings as holding the default organization's `default` role so menus and auth projection stay consistent.
- Frontend change plan: none expected for this refinement because current `views/system/user/**` already submits the required default role when creating users.
- Verification refinement: extend `go test ./app/admin/service/internal/...` with targeted tests for default-role fallback and projection behavior.

## Follow-up 2026-05-14 Results

- Backend code change:
  - `admin` now merges the default organization's `default` role into user-role upserts.
  - `admin` now returns a synthetic default-role binding for users with no explicit role binding, and includes that fallback in permission projection snapshots.
  - `auth` now allows default-role API access for unbound users only when no explicit role binding exists for the checked subject.
- Frontend code change: none in this refinement. Existing add-user modal logic remains as caller-side redundancy.
- Backend verification:
  - `go test ./app/admin/service/internal/...` passed.
  - `go test ./app/auth/service/internal/...` passed.
- Residual risk:
  - This refinement preserves current global `sys_user_role_binding` semantics. If later work introduces per-organization user-role binding storage, this fallback logic should be revisited together with the projection model.

## Risks

- Current worktrees contain existing uncommitted organization and generated API changes. This iteration must not revert unrelated changes.
- Full end-to-end verification may require running gateway/admin/user services and database seed outside this edit session.

## Follow-up 2026-05-14 Refresh 404

- Bug report:
  - Refreshing the browser could show an application 404 while backend routes were still present.
- Root cause:
  - `GetCurrentUserMenus` used only the current organization's explicit user-role bindings.
  - If a user switched to a non-default organization that had no explicit role binding, the menu tree became empty even though gateway authorization still allowed access through the default/global role fallback.
  - `/admin-api/v1/roles*` was also missing from the admin API resource catalog, so the user page could have a menu but fail its role-list request for `admin` users.
- Backend fix:
  - Current-user menu generation now falls back to the default organization's role bindings when the current organization has no explicit binding.
  - Admin role-list/create/update/delete endpoints are added to bootstrap and seed permission catalogs under the `role` resource group.
- Verification target:
  - Run focused admin tests.
  - Apply seed or bootstrap/sync projection in the running local environment, then verify menu and role APIs for `jack` and `vben`.
- Verification result:
  - `go test ./app/admin/service/internal/biz ./app/admin/service/internal/data` passed.
  - Local seed applied the admin/user/auth SQL updates. The existing projection-sync path still reports `JWT token is missing`, so the missing auth `g2` route mappings were inserted directly for this local dev database and `auth/admin/gateway` were restarted.
  - `vben` with current organization `test` now receives a non-empty menu tree including `/system/user`.
  - `jack` can call `/admin-api/v1/roles` and `/admin-api/v1/user-dept-bindings/{user_id}` through the gateway.

## Follow-up 2026-05-14 Refresh 404 Regression

- Bug report:
  - Browser refresh still lands on the application 404 page, and the browser console does not show a frontend exception.
- Current finding:
  - The startup requests can fail before dynamic menus are generated.
  - A local `jack` repro showed `403` for `/admin-api/v1/my/organizations`, `/admin-api/v1/menus/current`, and `/user-api/v1/users/{user_id}`.
  - The auth Casbin table may contain only the bootstrap root binding when admin permission projection sync fails, leaving `p2/g2` API policies unavailable to non-root users.
- Root cause to fix:
  - Admin projection snapshot building must not synthesize default user-role bindings by calling the user service without a user JWT.
  - Auth default-role fallback must match projected API route templates such as `/user-api/v1/users/{user_id}` instead of only exact paths.
- Backend change plan:
  - Keep admin projection snapshots limited to explicit user-role bindings; default-role fallback remains an auth runtime rule.
  - Update auth default-role fallback to resolve matching `g2` API groups with the same route-template matcher used by Casbin authorization.
- Verification target:
  - Run focused auth/admin tests for projection and default-role fallback.
  - Run `make sync-admin-projection` against the local services and verify `casbin_rules` contains `p2/g2` policies.
  - Re-test refresh for a non-root user after restarting affected services if needed.
- Verification result:
  - `go test ./app/auth/service/internal/biz -run 'TestCheckAuthorizationFallsBackToAdminProjectionDefaultRole|TestCheckAuthorizationFallsBackToDefaultRoleForUnboundUser|TestCheckAuthorizationDoesNotOverrideExplicitBinding|TestCheckAuthorizationUsesServiceNamespace|TestRegisterPermissionSnapshot' -count=1` passed.
  - `go test ./pkg/authx ./app/auth/service/internal/biz ./app/admin/service/internal/data ./app/admin/service/internal/biz` passed.
  - `make sync-admin-projection` passed; auth `casbin_rules` now contains `g`, `g2`, and `p2` rows, and admin `sys_projection_source_status` is `admin|snapshot|synced`.
  - Rebuilt and restarted `auth`, `admin`, and `gateway`; `admin` startup projection sync now logs `admin registered permission snapshot to auth` instead of `JWT token is missing`.
  - `jack` gateway checks returned `200` for `/admin-api/v1/my/organizations`, `/admin-api/v1/menus/current`, and `/user-api/v1/users/{user_id}`.
  - Browser refresh verification with `jack` on `/system/user` returned the `用户管理 - Vben Admin Antd` page, no 4xx startup requests, and no 404 fallback.

## Follow-up 2026-05-14 Permission Documentation Sync

- User request:
  - Align documentation with the current permission implementation.
- Change classification:
  - Level 1 docs-only synchronization.
- Current finding:
  - `docs/permission-design.md` still said organization did not participate in gateway `scope_id` authorization.
  - Current code sends the validated frontend current organization as `x-scope-id`; gateway forwards it into `auth.CheckAuthorization`; `admin` projects `sys_user_role_binding.organization_id` as auth `scope_id`.
- Documentation change:
  - `docs/permission-design.md` now documents organization-scoped role bindings, `x-scope-id`, auth subject key shape, default organization role fallback, menu fallback, and organization member removal cleanup.
  - `docs/table-ownership.md` now documents `sys_user_role_binding.organization_id` as the auth `scope_id` source and clarifies organization/member cleanup behavior.
  - `docs/monorepo-overview.md` now records that current organization participates in the permission chain.
- Code change:
  - None. This iteration only updates documentation to match the current implementation.
- Verification:
  - `git diff --check -- docs/permission-design.md docs/table-ownership.md docs/monorepo-overview.md docs/agent_runs/2026-05-13-organization-user-scope/plan.md` passed.
  - Stale wording search for the old "organization does not participate in scope authorization" model passed after excluding `**/agent_runs/**`.
