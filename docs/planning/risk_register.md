# Risk Register

| ID | Risk | Impact | Mitigation |
| --- | --- | --- | --- |
| R1 | Repo-managed config and Helm templates can accidentally receive real secrets | High | Keep `configs/config.yaml` on placeholder values; review Helm/runtime config before committing |
| R2 | Generated API, Ent, config, and Wire outputs can be overwritten | High | Edit source files only and regenerate from `make` targets |
| R3 | Schema auto-create runs on startup | High | Review data-model changes carefully before running against real databases |
| R4 | gRPC is wired but not started | Medium | Treat HTTP as the active transport unless the startup path changes |
| R5 | Casbin policy and JWT handling are coupled to middleware | High | Test auth and route coverage when touching auth or transport layers |
| R6 | Helm chart embeds runtime config and policy data | High | Review deployment changes for accidental secret exposure |
| R7 | There are no Go test files in the repository snapshot | Medium | Rely on build and manual smoke checks until tests are added |
| R8 | Monolith and microservice docs/commands can be mixed accidentally | High | Use `../REPO_STRUCTURE.md` and `../STARTUP_AND_INTEGRATION.md`; keep `master` paired with `vben-admin/main` |
| R9 | Backend-only changes can silently break frontend user flows | High | Treat feature work as full-stack; check paired frontend API clients, views, permissions, and smoke/build verification |

## Open Items

- `logs/` is untracked and should stay out of source control unless log handling is intentionally changed.
- `configs/config.yaml` is committed as a placeholder template; real database passwords, auth keys, and LLM API keys must stay in private local/runtime configuration.
