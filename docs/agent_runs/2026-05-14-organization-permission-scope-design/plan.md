# Organization Permission Scope Design

Date: 2026-05-14

## Scope

- Active version line: `base-server/monorepo` + `vben-admin/monorepo`.
- Change level: Level 2 design increment. The eventual implementation affects authorization semantics, but this iteration is docs-only.
- Goal: write a reviewable design for global API catalog, organization-scoped role permissions, organization available permission scope, and organization-admin anti-escalation rules.

## Current Decision

- API catalog remains global and is not copied per organization.
- Resource groups and menus remain global master data.
- Organizations receive an available permission scope.
- Organization admins can only assign permissions inside the current organization, inside the organization available scope, and inside their own grantable scope.
- `organization_id` should become the Casbin domain in the target model.
- Missing `x-scope-id` falls back to the default organization.
- Departments remain data-scope only and do not enter gateway Casbin authorization.
- `data_scope` belongs to roles and is enforced in service/biz data filtering after gateway/auth API allow/deny.
- Super admin is a global root and must not depend on the request domain.

## Backend Impact

- This iteration only adds documentation.
- Future implementation will touch `admin` data/schema/API/usecase, `auth` Casbin model/projection handling, `gateway` authorization request semantics, and permission seed/projection sync.

## Frontend Impact

- This iteration does not change frontend code.
- Future implementation will change role permission forms, organization detail permission-scope management, and visibility of global API/resource/menu management pages.

## Verification

- Docs-only check: `git diff --check -- base-server/docs/organization-permission-scope-design.md base-server/docs/README.md base-server/docs/agent_runs/2026-05-14-organization-permission-scope-design/plan.md`

## Results

- Added `docs/organization-permission-scope-design.md` as the reviewable target design.
- Updated `docs/README.md` to include the new design document in the docs index.
- Verification passed: `git -C base-server diff --check -- docs/organization-permission-scope-design.md docs/README.md docs/agent_runs/2026-05-14-organization-permission-scope-design/plan.md`.

## Follow-up 2026-05-15 Data Scope Review

- User feedback:
  - Add role `data_scope` values: `all`, `self_dept`, `self_dept_and_child`, `self`, `custom_depts`.
  - Keep gateway/auth as API allow/deny only: `user_id + organization_id + API`.
  - Enforce department/data visibility in service/biz after API authorization.
  - Preserve `p/p2/p3` and `g/g2/g3` model shape.
  - Super admin should be global root, not domain-dependent root.
- Documentation update:
  - Added the two-step permission flow.
  - Updated the target Casbin text model and strategy examples.
  - Added role `data_scope` and `data_scope_dept_ids` model guidance.
  - Added service/biz data-scope filtering flow, multi-role union semantics, migration notes, and verification cases.

## Follow-up 2026-05-15 Implementation

- User approval: implement the reviewed organization permission scope design.
- Change level: Level 2 feature increment with security-sensitive authorization behavior. It changes auth projection semantics, admin role data, and frontend role forms, but stays within the existing Go/Kratos + Vben stack.
- Dirty state policy: both child repositories already contain unrelated organization/member changes and generated files. This iteration will edit in place, avoid branch switching/worktrees, and not revert unrelated changes.
- Backend tasks:
  - Update `auth` Casbin model to use organization domain and global root.
  - Update `auth` snapshot/delta helpers to emit `g(user, role, org)`, `p2(role, org, resource, method)`, and global `g2(path, resource)`.
  - Add role `data_scope` and `data_scope_dept_ids` fields through schema/proto/repo mapping.
  - Add organization permission scope schema/API only if implementation scope can complete safely in this pass; otherwise keep the design doc as the source for that follow-up.
  - Preserve default organization no-scope fallback.
- Frontend tasks:
  - Surface role data scope fields in role form and API types after OpenAPI generation, if backend contract changes.
  - Keep global API/resource management visibility behavior unchanged unless backend exposes the organization permission-scope API in this pass.
- Verification plan:
  - Backend targeted tests: `go test ./app/auth/service/internal/biz ./app/admin/service/internal/biz ./app/admin/service/internal/data ./pkg/authx`.
  - Generation checks as needed: `make ent`, `make api`, `make frontend-api`.
  - Frontend targeted checks: existing role/user helper tests where applicable; typecheck may still hit known existing issues and must be recorded.

## Follow-up 2026-05-15 Implementation Result

- Backend implemented:
  - `auth` Casbin runtime now uses `r = sub, dom, obj, act`; `organization_id` / `scope_id` is the domain.
  - API grouping is global: `g2(path, resource_group)` with no service-qualified key.
  - Role API policies are domain-scoped: `p2(role_value, organization_id, resource_group, method)`.
  - Missing `x-scope-id` falls back to the default organization ID.
  - `root` is global: `g(user_id, root, global)`, independent of request domain.
  - Snapshot registration now rebuilds the full projection because service-prefix cleanup is no longer safe.
  - User-role binding delta now removes root bindings from the `global` domain and preserves other roles in the same organization when adding one binding.
  - `sys_role` now carries `data_scope` and `data_scope_dept_ids`; proto, Ent schema, repo mapping, seed, migration backfill, and projection DTOs were updated.
  - Role save validates allowed data-scope values and validates `custom_depts` departments belong to the role organization.
- Frontend implemented:
  - OpenAPI was regenerated from backend contract.
  - Role form exposes `data_scope` with values `all`, `self_dept`, `self_dept_and_child`, `self`, `custom_depts`.
  - `custom_depts` shows a current-organization department tree selector.
  - Role create/update payload is normalized to backend proto-name fields only to avoid duplicate camel/snake fields.
  - Role list displays data scope.
- Documentation updated:
  - `docs/permission-design.md` now describes organization-domain auth instead of service/scope key encoding.
  - `docs/auth-incremental-sync-design.md` now describes full snapshot rebuild, global API groups, organization domain, and global root.
  - `docs/monorepo-overview.md`, `docs/table-ownership.md`, and `docs/README.md` were aligned with organization-domain terminology.
- Deferred:
  - `sys_organization_permission_scope` table/API and organization-admin anti-escalation checks are still design-only in this pass.
  - Role menu/resource save currently validates data scope and organization ownership, but does not yet enforce "organization available scope ∩ actor grantable scope".

## Verification Results 2026-05-15

- Generation:
  - `make api` passed.
  - `make ent` passed.
  - `make frontend-api` completed and regenerated `vben-admin/openapi.yaml` plus generated clients.
- Backend:
  - `go test ./app/auth/service/internal/biz ./pkg/authx ./app/admin/service/internal/biz ./app/admin/service/internal/data -count=1` passed.
- Frontend:
  - `pnpm --dir /home/mini/OpenSource/framework/vben-admin exec vitest run --dom apps/web-antd/src/api/system/role-payload.test.ts apps/web-antd/src/api/system/organization-payload.test.ts` passed.
  - `pnpm --dir /home/mini/OpenSource/framework/vben-admin --filter @vben/web-antd run typecheck` failed only on known existing issues already tracked in `vben-admin/docs/project/known-issues.md`; after fixes, no remaining errors were reported from the new role payload files or touched role form file.
- Whitespace:
  - `git -C base-server diff --check -- ...` and `git -C vben-admin diff --check -- ...` passed for the touched implementation/documentation files.

## Follow-up 2026-05-15 Permission Scope Completion

- User request: continue until the organization permission scope design is fully implemented.
- Active version line: `base-server/monorepo` + `vben-admin/monorepo`.
- Change level: Level 2 feature increment. It adds admin-service API contract, persistence, role write validation, and frontend organization permission-scope management UI.
- Dirty state policy: both child repositories are already dirty with related organization/auth/frontend generated changes. Continue in place, do not switch branches, and do not revert unrelated files.
- Backend task boundaries:
  - Add `sys_organization_permission_scope` Ent schema and generated code.
  - Add admin proto APIs:
    - `GET /admin-api/v1/organizations/{organization_id}/permission-scope`
    - `PUT /admin-api/v1/organizations/{organization_id}/permission-scope`
    - `GET /admin-api/v1/permission-catalog/current`
    - `GET /admin-api/v1/organizations/{organization_id}/permission-catalog`
  - Add repo/usecase/service methods to load/save organization scope, validate referenced menus/resources, and block shrinking scope while used by roles.
  - Enforce role menu/resource permissions are within configured organization scope.
  - Enforce user role and department bindings target users who are members of the organization.
- Frontend task boundaries:
  - Add organization permission-scope API wrapper.
  - Add organization-list action and drawer to configure menu/resource scope.
  - Use permission catalog in the role form so role assignment UI matches backend scope.
- Verification plan:
  - Generate backend Ent/API and frontend OpenAPI client after contract/schema changes.
  - Backend targeted tests: `go test ./app/admin/service/internal/biz ./app/admin/service/internal/data ./app/auth/service/internal/biz ./pkg/authx -count=1`.
  - Frontend targeted tests: existing payload/helper tests plus any new permission-scope payload tests.
  - Record any broader typecheck limits against `vben-admin/docs/project/known-issues.md`.

## Follow-up 2026-05-15 Permission Scope Completion Result

- Backend implemented:
  - Added `sys_organization_permission_scope` Ent schema, generated Ent code, admin migration inclusion, and seed/projection wait coverage.
  - Added admin APIs for organization permission scope and permission catalog:
    - `GET /admin-api/v1/organizations/{organization_id}/permission-scope`
    - `PUT /admin-api/v1/organizations/{organization_id}/permission-scope`
    - `GET /admin-api/v1/permission-catalog/current`
    - `GET /admin-api/v1/organizations/{organization_id}/permission-catalog`
  - Added scope persistence and validation for menu/resource references.
  - Empty menu/resource scope now means no corresponding assignable permissions.
  - Scope shrink is blocked when an existing role in that organization still uses a removed menu/resource.
  - Role create/update now enforces selected menus/resources are inside configured organization scope.
  - Role create/update and user-role binding now enforce actor-grantable boundaries for menus, API resources, and role data scope.
  - Only bootstrap root user `f4f9e258-fa13-4467-95fb-c86019a377f9` can create/update/assign `root`.
  - Only bootstrap root can create, update, or delete organization master data.
  - Only bootstrap root can update an organization's permission scope.
  - Non-root organization list reads are scoped to the default organization and the actor's current organization so the member drawer can still compare "all members" with the current organization without exposing unrelated organizations.
  - Non-root actors cannot pass an explicit `organization_id` that differs from their current organization on organization-scoped role, department, member-save, permission-scope, or permission-catalog APIs. Member read keeps the required exception that an organization admin may read the default organization member source and the current organization.
  - User role and department bindings now require the target user to belong to the target organization.
  - Site-message bootstrap preserves existing role data-scope fields and only adds built-in site-message permissions when they are inside the configured organization permission scope.
- Frontend implemented:
  - Added organization permission-scope API wrappers and payload normalization tests.
  - Added an organization-list `权限范围` drawer that loads global menus/resources, loads the selected organization's scope, and saves selected menu/resource IDs.
  - Role form now reloads `GET /admin-api/v1/permission-catalog/current` each time the drawer opens, so menus/API resources match the current organization and organization scope.
  - Organization action column was widened for the added action.
- Documentation updated:
  - This plan now records the completed backend/frontend implementation and verification.
  - `docs/organization-permission-scope-design.md` status was updated from review draft to implemented design, with compatibility and completed implementation notes.

## Verification Results 2026-05-15 Permission Scope Completion

- Formatting:
  - `gofmt` passed for touched backend files.
  - `pnpm exec oxfmt --write ...` passed for touched frontend files.
- Backend:
  - `go test ./app/auth/service/internal/biz ./pkg/authx ./app/admin/service/internal/biz ./app/admin/service/internal/data -count=1` passed.
  - Added regression coverage for platform-root-only organization mutations and permission-scope updates, non-root organization-list scoping, and non-root cross-organization rejection.
- Frontend:
  - `pnpm exec vitest run --dom apps/web-antd/src/api/system/role-payload.test.ts apps/web-antd/src/api/system/organization-payload.test.ts apps/web-antd/src/views/system/organization/helpers.test.ts apps/web-antd/src/views/system/user/helpers.test.ts apps/web-antd/src/router/guard.test.ts` passed with 5 test files and 20 tests.
  - `pnpm --filter @vben/web-antd exec vue-tsc --noEmit --skipLibCheck --pretty false` still fails only on known existing files tracked in `vben-admin/docs/project/known-issues.md`; no final error references the new organization permission-scope files or the touched role form.
- Whitespace:
  - `git diff --check -- ...` passed for the touched backend and frontend implementation/documentation files.

## Remaining Risks

- Browser-level manual verification is still recommended for the new organization permission-scope drawer because it uses the shared tree component with nested menu/resource selection.
- Project-wide frontend typecheck remains blocked by existing known issues outside this iteration.

## Follow-up 2026-05-15 Clean Current Permission Model

- User feedback: the project has not been put into real use, so old data formats and old permission terminology do not need compatibility support.
- Change level: Level 2 model cleanup. This removes compatibility branches and old `scope/service` authorization wording, but keeps the accepted organization-domain permission model.
- Backend cleanup scope:
  - Remove legacy organization ID conversion rules such as `0/1/100000 -> default organization` and hash-based non-UUID conversion.
  - Keep only current-model schema setup and default organization seeding.
  - Rename auth authorization/projection organization context from `scope_id` to `organization_id`.
  - Stop sending service/resource-group metadata through gateway-to-auth authorization requests; auth decides by `path + method + organization_id` against the projected API catalog.
  - Remove unused service-qualified auth helper methods and tests built around service namespace compatibility.
  - Treat organization permission scope as explicit current-model data: empty scope denies assigning that permission type.
  - Seed/bootstrap the default organization with all current menus and resource groups so the default organization remains fully usable without relying on empty-scope compatibility.
- Frontend cleanup scope:
  - Send current organization with `x-organization-id` instead of `x-scope-id`.
  - Regenerate OpenAPI clients if backend proto/OpenAPI output changes frontend generated files.
- Documentation cleanup scope:
  - Align permission, API ownership, table ownership, and sync docs with the current organization-domain model.
  - Remove statements that describe old business-domain, service-qualified key, or scope compatibility as active behavior.
- Verification plan:
  - Backend targeted tests: `go test ./app/auth/service/internal/biz ./pkg/authx ./app/admin/service/internal/biz ./app/admin/service/internal/data ./app/gateway/service/internal/middleware/casbin -count=1`.
  - Frontend targeted tests: existing organization/role payload and router tests if generated API changes touch frontend.
  - `git diff --check` in both repositories.

## Follow-up 2026-05-15 Clean Current Permission Model Result

- Backend implemented:
  - `CheckAuthorizationRequest` now uses `organization_id`; generated Go/OpenAPI output was regenerated.
  - Gateway Casbin middleware now forwards only `user_id + path + method + x-organization-id` to `auth`; service-prefix and resource-group lookup in gateway was removed.
  - `auth` projection DTOs and admin projection mapping now use `organization_id` instead of `scope_id`, and service-qualified helper code was removed.
  - Startup migration no longer runs legacy cleanup/backfill branches for old organization IDs, old domain columns, or old user-role unique indexes.
  - Organization permission scope is no longer compatibility-open when empty; role permission validation and site-message bootstrap only use explicitly configured organization scope.
  - Default organization bootstrap/seed now explicitly inserts all current menu/resource scope rows into `sys_organization_permission_scope`.
- Frontend implemented:
  - Request client sends `x-organization-id`; no active frontend code sends `x-scope-id`.
  - OpenAPI and generated API clients were regenerated from the cleaned backend contract.
- Documentation updated:
  - `permission-design.md`, `api-ownership.md`, `table-ownership.md`, `auth-incremental-sync-design.md`, and `organization-permission-scope-design.md` now describe the current organization-domain model without old-data compatibility as active behavior.
- Verification:
  - `make api` passed.
  - `make frontend-api` passed.
  - `go test ./app/auth/service/internal/biz ./pkg/authx ./app/admin/service/internal/biz ./app/admin/service/internal/data ./app/gateway/service/internal/middleware/casbin -count=1` passed.
  - `pnpm exec vitest run --dom apps/web-antd/src/api/system/role-payload.test.ts apps/web-antd/src/api/system/organization-payload.test.ts apps/web-antd/src/views/system/organization/helpers.test.ts apps/web-antd/src/views/system/user/helpers.test.ts apps/web-antd/src/router/guard.test.ts` passed with 5 test files and 20 tests.
  - `git diff --check` passed in both repositories.
- Residual risk:
  - Existing local databases with old columns or old duplicate data are no longer automatically repaired on service startup. This matches the current no-legacy-compatibility decision; use a fresh database or reseed when needed.

## Follow-up 2026-05-15 Local Database Audit

- User request: check whether the current local database still contains old columns, duplicate data, or old permission data.
- Scope: read-only SQL audit against local PostgreSQL `127.0.0.1:25432`, databases `admin` and `auth`.
- Results:
  - `admin` old tables: `sys_business_domain` and `sys_dept_role_default` do not exist.
  - `admin` old permission rows: no domain/sample API resources, no domain/sample resources, no `organization:*` permission rows found.
  - `admin` duplicate data: no duplicates for `sys_api_resources(path, method)`, `sys_role(organization_id, value)`, `sys_user_role_binding(user_id, role_id, organization_id)`, `sys_user_organization(user_id, organization_id)`, active primary organization, `sys_user_dept_membership(user_id, organization_id)`, or `sys_organization_permission_scope(organization_id, permission_type, permission_ref)`.
  - `admin` old organization IDs: no `0`, `1`, `100000`, `organization:*`, or non-UUID organization IDs found in current organization-linked tables.
  - `admin` old columns still present: `sys_role.scope_type`; `sys_user_role_binding.service`, `sys_user_role_binding.source_type`, `sys_user_role_binding.source_id`.
  - `admin` old index still present: `userrolebinding_user_id_role_id` unique index on `sys_user_role_binding(user_id, role_id)`, which conflicts with the current per-organization uniqueness model.
  - `admin` generated/hashed index names still present for current unique constraints: `organizationpermissionscope_or_c04539f8824e29ac2a9e30caa64c8d6e` and `sys_user_dept_membership_user_id_organization_id_key`; these are not data-model stale, only names differ from Ent-generated expectation.
  - `admin` current organization permission scope is empty: `sys_organization_permission_scope` has 0 rows, while `sys_menu` has 25 rows and `sys_resources` has 11 rows. This means the newly added default-organization explicit scope bootstrap has not been applied to this local DB yet.
  - `auth` Casbin rules: no duplicate rules, no service-qualified old keys, no `organization:*`, no `0/1/100000` domains, and p2/g domains are UUID/default-global as expected.
- Recommended cleanup if the user approves DB mutation:
  - Drop `sys_role.scope_type`.
  - Drop `sys_user_role_binding.service`, `source_type`, and `source_id`.
  - Drop unique index `userrolebinding_user_id_role_id`.
  - Run admin bootstrap/seed so default organization receives explicit menu/resource scope rows.

## Follow-up 2026-05-15 Local Database Cleanup Result

- User request: continue after the local database audit and clean remaining old local DB structures/data.
- Database mutations executed against local `admin` database:
  - Dropped old unique index `public.userrolebinding_user_id_role_id`.
  - Dropped old column `public.sys_role.scope_type`.
  - Dropped old columns `public.sys_user_role_binding.service`, `source_type`, and `source_id`.
  - Inserted default organization explicit permission scope rows for all current menus and resources.
- First insert attempt failed inside a transaction because `sys_organization_permission_scope.id` has no DB-side default; the transaction was rolled back automatically. The successful retry inserted deterministic IDs `default-menu-*` and `default-resource-*`.
- Verification after cleanup:
  - No unexpected columns remain in current admin permission/organization tables.
  - No targeted old columns remain.
  - Old unique index `userrolebinding_user_id_role_id` no longer exists.
  - No duplicates remain for API path/method, role organization/value, user role binding organization tuple, or organization permission scope tuple.
  - Default organization scope now covers all current `sys_menu` rows and all current `sys_resources` rows: 25 menu rows and 11 resource rows.
  - `auth.casbin_rules` still has no duplicate rules or service-qualified old keys; p2 domain remains `9f740c1b-0210-4e3a-858d-d128edea924d`.
