#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
if docker compose version >/dev/null 2>&1; then
  COMPOSE=(docker compose -f "$ROOT_DIR/docker-compose.yml")
else
  COMPOSE=(docker-compose -f "$ROOT_DIR/docker-compose.yml")
fi

"${COMPOSE[@]}" up -d --build "$@"

printf '\nFull stack is starting.\n'
printf 'If you are upgrading from an older PostgreSQL major version, recreate volumes first:\n'
printf '  %s down -v\n' "${COMPOSE[*]}"
printf 'Next step: %s\n' "$ROOT_DIR/deploy/scripts/seed.sh"
printf 'Gateway: http://127.0.0.1:8000\n'
printf 'Jaeger:  http://127.0.0.1:16686\n'
printf 'Consul:  http://127.0.0.1:8500\n'
