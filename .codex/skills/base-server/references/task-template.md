# Task Template

## Before Editing

- Read the project docs listed in `AGENTS.md`.
- Check `git status --short --branch` and worktree list.
- Identify the active frontend/backend version line and inspect paired `../vben-admin` impact.
- Identify exact write paths and whether generated files are involved.
- Mark secrets, generated files, logs, and build outputs as out of scope unless explicitly requested.

## During Editing

- Keep source edits in the owning boundary.
- If a generated file is needed, regenerate it from the source.
- Avoid expanding hardcoded config or secrets.

## Before Handoff

- Run the smallest relevant generation and verification commands.
- Include paired frontend verification when the user flow, API client, type contract, auth, upload, or SSE behavior can be affected.
- Update project docs when decisions, assumptions, or task boundaries changed.
- Report verification gaps if database, Docker, Helm, or external services prevent a full smoke test.
