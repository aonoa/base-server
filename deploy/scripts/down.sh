#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
if docker compose version >/dev/null 2>&1; then
  COMPOSE=(docker compose -f "$ROOT_DIR/docker-compose.yml")
else
  COMPOSE=(docker-compose -f "$ROOT_DIR/docker-compose.yml")
fi

"${COMPOSE[@]}" down "$@"
