# Verification

## Setup

```bash
make init
sh scripts/git/install-config-safety.sh
```

## Generate

```bash
make api
make config
make wire
make ent
make all
```

## Build

```bash
make build
go build ./...
```

## Tests

```bash
go test ./...
```

## Manual Smoke Checks

```bash
go run ./cmd/base-server -conf ./configs
docker build -t base-server .
helm template base-server ./deploy/helm/base-server
```

## Verification Policy

- Use the narrowest generation command that matches the touched source.
- If proto, config schema, or Ent schema changed, regenerate before compiling.
- If server wiring or middleware changed, run at least a full `go test ./...` and a build.
- If deployment manifests changed, validate the Helm chart render.
- For feature additions, modifications, or refactors, include the paired frontend verification that covers the affected user flow when frontend behavior, generated clients, request/response types, auth, upload, or SSE behavior can be affected.
- If only backend verification is required, state why the paired frontend does not need a code or verification change.
- Before committing `configs/config.yaml`, install the repo-local hook/filter and confirm the staged blob contains `CHANGE_ME_*` placeholders rather than local secrets.
