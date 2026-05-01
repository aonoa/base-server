#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
APP_COMPOSE_FILE="$ROOT_DIR/docker-compose.yml"
DEV_ENV_COMPOSE_FILE="$ROOT_DIR/docker-compose-env.yml"
SEED_MODE="${SEED_MODE:-auto}"

if docker-compose version >/dev/null 2>&1; then
  COMPOSE_BIN=(docker-compose)
elif docker compose version >/dev/null 2>&1; then
  COMPOSE_BIN=(docker compose)
else
  echo "docker compose / docker-compose not found"
  exit 1
fi

compose_cmd() {
  local compose_file="$1"
  shift
  "${COMPOSE_BIN[@]}" -f "$compose_file" "$@"
}

compose_service_id() {
  local compose_file="$1"
  local service="$2"
  compose_cmd "$compose_file" ps -q "$service" 2>/dev/null | tail -n 1
}

compose_has_running_service() {
  local compose_file="$1"
  local service="$2"
  local container_id
  container_id="$(compose_service_id "$compose_file" "$service")"
  [ -n "$container_id" ]
}

detect_mode() {
  case "$SEED_MODE" in
    compose|dev)
      printf '%s' "$SEED_MODE"
      return 0
      ;;
    auto)
      if compose_has_running_service "$APP_COMPOSE_FILE" admin; then
        printf '%s' "compose"
      else
        printf '%s' "dev"
      fi
      return 0
      ;;
    *)
      echo "Unsupported SEED_MODE=$SEED_MODE, expected auto|compose|dev" >&2
      exit 1
      ;;
  esac
}

MODE="$(detect_mode)"
if [ "$MODE" = "compose" ]; then
  POSTGRES_COMPOSE_FILE="$APP_COMPOSE_FILE"
else
  POSTGRES_COMPOSE_FILE="$DEV_ENV_COMPOSE_FILE"
fi

POSTGRES_EXEC=("${COMPOSE_BIN[@]}" -f "$POSTGRES_COMPOSE_FILE" exec -T postgres psql -v ON_ERROR_STOP=1 -U postgres)

wait_for_postgres() {
  until compose_cmd "$POSTGRES_COMPOSE_FILE" exec -T postgres pg_isready -U postgres -d postgres >/dev/null 2>&1; do
    sleep 2
  done
}

wait_for_table() {
  local db="$1"
  local table="$2"
  until compose_cmd "$POSTGRES_COMPOSE_FILE" exec -T postgres psql -U postgres -d "$db" -tAc "SELECT to_regclass('public.${table}') IS NOT NULL" | grep -q t; do
    sleep 2
  done
}

wait_for_policy_projection() {
  until compose_cmd "$POSTGRES_COMPOSE_FILE" exec -T postgres psql -U postgres -d auth -tAc "SELECT EXISTS (SELECT 1 FROM casbin_rules WHERE ptype = 'g' AND v0 = 'f4f9e258-fa13-4467-95fb-c86019a377f9' AND v1 = 'role:root') AND EXISTS (SELECT 1 FROM casbin_rules WHERE ptype = 'g' AND v0 = 'a0bb672a-a4b1-4ec9-807a-ba11e000d2a4' AND v1 = 'role:admin') AND EXISTS (SELECT 1 FROM casbin_rules WHERE ptype = 'g2' AND v0 = '/admin-api/v1/menus' AND v1 = 'api:menu')" | grep -q t; do
    sleep 2
  done
}

wait_for_local_port() {
  local port="$1"
  local name="$2"
  local attempts="${3:-60}"
  local i
  for ((i = 0; i < attempts; i++)); do
    if bash -c "exec 3<>/dev/tcp/127.0.0.1/${port}" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  echo "Timed out waiting for local ${name} on 127.0.0.1:${port}" >&2
  return 1
}

apply_seed() {
  local db="$1"
  local file="$2"
  printf 'Applying %s\n' "$file"
  "${POSTGRES_EXEC[@]}" -d "$db" < "$file"
}

wait_for_postgres
wait_for_table admin sys_role
wait_for_table admin sys_api_resources
wait_for_table admin sys_user_role_binding
wait_for_table user sys_user
wait_for_table admin sys_menu
wait_for_table admin sys_dept
wait_for_table admin sys_business_domain
wait_for_table admin sys_service_registry
wait_for_table admin sys_projection_source_status

apply_seed admin "$ROOT_DIR/deploy/sql/seed/admin.sql"
apply_seed user "$ROOT_DIR/deploy/sql/seed/user.sql"
apply_seed auth "$ROOT_DIR/deploy/sql/seed/auth.sql"

if [ "$MODE" = "compose" ]; then
  compose_cmd "$APP_COMPOSE_FILE" restart admin >/dev/null
else
  wait_for_local_port 9020 "auth service"
  printf 'Syncing admin permission snapshot in dev mode\n'
  (cd "$ROOT_DIR" && make sync-admin-projection)
fi

wait_for_table auth casbin_rules
wait_for_policy_projection

printf '\nSeed completed.\n'
printf 'Default users:\n'
printf '  - vben / 123456  (root)\n'
printf '  - jack / 123456  (admin)\n'
printf 'Seed mode: %s\n' "$MODE"
printf '\nIf you need a clean rerun, remove volumes with:\n'
printf '  %s -f %s down -v\n' "${COMPOSE_BIN[*]}" "$POSTGRES_COMPOSE_FILE"
