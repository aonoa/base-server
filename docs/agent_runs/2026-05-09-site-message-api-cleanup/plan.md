# Site Message API Cleanup

## Goal

Remove the legacy `receiverType` and `receiverIds` fields from the site-message API contract after the product and database have both converged on all-user delivery only.

## Classification

- Level 3 public API change
- Active version line: `base-server/master` + `vben-admin/main`
- Backend impact: yes
- Frontend impact: yes

## Execution Notes

- Both repositories already contain in-progress site-message work, so this API cleanup is executed sequentially in the main worktrees.
- No automatic parallel child execution is used because the changed files overlap the same feature area and generated artifacts.

## Requirements Update

- `PublishedSiteMessageItem` must no longer expose `receiverType` or `receiverIds`.
- `CreateSiteMessageRequest` must no longer accept `receiverType` or `receiverIds`.
- Backend generated API artifacts and paired frontend OpenAPI/client outputs must be regenerated in the same version line.
- Frontend handwritten wrappers and pages must stop referencing the removed fields.

## Scope

- `api/protos/base_api/v1/base.proto`
- `api/gen/go/base_api/v1/**`
- `cmd/base-server/assets/openapi.yaml`
- `internal/biz/base.go`
- `internal/data/site_message_test.go`
- paired `vben-admin/main` generated API files, handwritten site-message wrapper, manager page, and docs

## Tasks

- [x] Record the API cleanup scope and verification plan.
- [x] Remove the legacy audience fields from the backend proto and service mapping.
- [x] Regenerate backend API/OpenAPI outputs and paired frontend generated API files.
- [x] Remove remaining frontend wrapper/page references to the deleted fields.
- [x] Run targeted verification and update project docs.

## Verification

- [x] `make api`
- [x] `make frontend-api`
- [x] `go test ./...`
- [x] `pnpm -C /home/mini/OpenSource/framework/vben-admin --filter @vben/web-antd exec vue-tsc --noEmit` (still fails on unrelated pre-existing frontend repo issues in copilot, generated request, dashboard, and existing modal typings)

## Risks

- This is a public contract change inside the active version line, so backend and frontend generated artifacts must stay in sync.
- Frontend repo still has unrelated existing typecheck failures outside the site-message files.

## Result

- `CreateSiteMessageRequest` and `PublishedSiteMessageItem` no longer expose `receiverType` or `receiverIds`.
- Backend Go bindings, OpenAPI output, frontend generated API models, and handwritten site-message wrappers are now aligned to the reduced contract.
