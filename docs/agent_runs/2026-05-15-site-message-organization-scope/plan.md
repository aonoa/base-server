# Site Message Organization Scope Iteration

## Goal

Make site messages scoped to the current organization. The default organization already contains all users, so publishing while the current organization is default is equivalent to all-user publishing. Publishing from any other organization only creates receipts for active members of that organization.

## Change Level

Level 2 feature increment: storage semantics, service integration, API contract, generated clients, UI labels, and regression tests change.

## Active Version Line

- Backend: `base-server/monorepo`
- Frontend: `vben-admin/monorepo`
- API prefixes: `/common-api/*` for site messages, `/admin-api/*` for organization membership lookup.

## Current Baseline

- Both repositories are clean and ahead of remote by one local commit from the previous organization permission iteration.
- Existing site message implementation broadcasts to all active users through `user.GetUserList`.
- `sys_site_message` and `sys_site_message_receipt` have no organization column.

## Accepted Behavior

- Creating, updating draft, scheduling, and immediate publishing of site messages uses the request current organization.
- Drafts and scheduled messages store the organization at save time.
- Scheduled publish resolves recipients using the saved organization, not whatever organization a later request happens to carry.
- Manage list only shows messages for the current organization.
- Inbox list, unread count, mark read/unread, and mark all read operate only in the current organization.
- Default organization needs no special global scope because it contains all users.
- Frontend displays target range as current organization members.
- Site message management permission is checked against the same current organization that receives the message.
- Internal service calls that depend on request scope forward both `Authorization` and `x-organization-id`.

## Backend Tasks

- Add `organization_id` to `SiteMessage` and `SiteMessageReceipt` schemas and indexes.
- Add generated Ent fields through `make ent`.
- Add internal admin RPC for organization member user IDs or reuse an existing safe contract if suitable.
- Change common repo to resolve organization from `x-organization-id` with default fallback, store it, and resolve active organization members through admin/user data.
- Update common proto replies to include organization id where useful for UI/debugging.
- Update targeted tests for organization-scoped receipts and scheduled publishing.

## Frontend Tasks

- Regenerate OpenAPI client after backend proto changes.
- Update site message API wrapper types to include organization id.
- Update manage page copy from generic all-user to current organization members.
- Use existing organization store to show current organization name.

## Verification Strategy

Backend targeted:

```bash
go test ./pkg/authx ./app/common/service/internal/data ./app/common/service/internal/biz ./app/admin/service/internal/biz ./app/admin/service/internal/data -count=1
```

Frontend targeted:

```bash
pnpm exec vitest run --dom apps/web-antd/src/api/system/organization-payload.test.ts apps/web-antd/src/views/system/organization/helpers.test.ts apps/web-antd/src/views/system/user/helpers.test.ts apps/web-antd/src/router/guard.test.ts
```

Additional checks:

```bash
git diff --check
```

Full frontend `vue-tsc` is a known pre-existing failure and not a completion gate for this iteration.

## Risks

- Common service currently has no durable default organization constant except auth; fallback will use the known fixed default ID already documented in the platform.
- Existing databases will get new columns via Ent migration. Project is not in production, so no old-data compatibility work is required.
- Generated OpenAPI may remove stale sample service files as in the previous iteration.

## Implementation Result

- `sys_site_message` and `sys_site_message_receipt` now include `organization_id`.
- `common` stores the current organization on draft, scheduled, and published messages.
- Published messages create receipts only for active members returned by `admin.ListOrganizationMemberUserIds`.
- Scheduled messages resolve recipients using the saved `organization_id`.
- Inbox list, unread count, mark read/unread, mark all read, manage list, recall, delete, and publish pending are scoped to current organization.
- Site message management access now resolves user role values with an explicit organization id.
- `authx.ForwardAuthorizationContext` now forwards `x-organization-id` so internal gRPC calls keep the frontend-selected organization scope.
- Frontend message management shows the target range as the current organization members and refreshes message pages when the organization changes.

## Verification Result

- `go test ./pkg/authx ./app/common/service/internal/data ./app/common/service/internal/biz ./app/admin/service/internal/biz ./app/admin/service/internal/data -count=1` passed.
- `pnpm exec eslint apps/web-antd/src/api/system/site-message.ts apps/web-antd/src/views/_core/messages/manage.vue apps/web-antd/src/views/_core/messages/inbox.vue` passed.
- `pnpm exec vitest run --dom apps/web-antd/src/api/system/organization-payload.test.ts apps/web-antd/src/views/system/organization/helpers.test.ts apps/web-antd/src/views/system/user/helpers.test.ts apps/web-antd/src/router/guard.test.ts` passed.
- `git diff --check` passed in both repositories.
