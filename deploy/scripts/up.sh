#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
FULL_STACK_CONTAINERS=(base-postgres base-redis base-jaeger base-consul base-user base-auth base-admin base-common base-gateway)
if docker compose version >/dev/null 2>&1; then
  COMPOSE=(docker compose -f "$ROOT_DIR/docker-compose.yml")
  COMPOSE_NEEDS_RECREATE_WORKAROUND=0
else
  COMPOSE=(docker-compose -f "$ROOT_DIR/docker-compose.yml")
  COMPOSE_NEEDS_RECREATE_WORKAROUND=1
fi

if [ "$COMPOSE_NEEDS_RECREATE_WORKAROUND" = "1" ]; then
  printf 'docker-compose v1 detected, removing existing containers to avoid recreate bug\n'
  for name in "${FULL_STACK_CONTAINERS[@]}"; do
    ids=$(docker ps -aq --filter "name=${name}")
    if [ -n "$ids" ]; then
      docker rm -f $ids >/dev/null 2>&1 || true
    fi
  done
fi

"${COMPOSE[@]}" up -d --build "$@"

printf '\nFull stack is starting.\n'
printf 'If you are upgrading from an older PostgreSQL major version, recreate volumes first:\n'
printf '  %s down -v\n' "${COMPOSE[*]}"
printf 'Next step: %s\n' "$ROOT_DIR/deploy/scripts/seed.sh"
printf 'Gateway: http://127.0.0.1:8000\n'
printf 'Jaeger:  http://127.0.0.1:16686\n'
printf 'Consul:  http://127.0.0.1:8500\n'
