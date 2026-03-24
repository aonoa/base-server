#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
if docker compose version >/dev/null 2>&1; then
  COMPOSE=(docker compose -f "$ROOT_DIR/docker-compose.yml")
else
  COMPOSE=(docker-compose -f "$ROOT_DIR/docker-compose.yml")
fi
POSTGRES_EXEC=("${COMPOSE[@]}" exec -T postgres psql -v ON_ERROR_STOP=1 -U postgres)

wait_for_postgres() {
  until "${COMPOSE[@]}" exec -T postgres pg_isready -U postgres -d postgres >/dev/null 2>&1; do
    sleep 2
  done
}

wait_for_table() {
  local db="$1"
  local table="$2"
  until "${COMPOSE[@]}" exec -T postgres psql -U postgres -d "$db" -tAc "SELECT to_regclass('public.${table}') IS NOT NULL" | grep -q t; do
    sleep 2
  done
}

wait_for_policy_reload() {
  until "${COMPOSE[@]}" exec -T postgres psql -U postgres -d auth -tAc "SELECT COUNT(*) > 0 FROM casbin_rules" | grep -q t; do
    sleep 2
  done
}

apply_seed() {
  local db="$1"
  local file="$2"
  printf 'Applying %s\n' "$file"
  "${POSTGRES_EXEC[@]}" -d "$db" < "$file"
}

wait_for_postgres
wait_for_table auth sys_role
wait_for_table user sys_user
wait_for_table admin sys_menu
wait_for_table admin sys_dept

apply_seed auth "$ROOT_DIR/deploy/sql/seed/auth.sql"
apply_seed admin "$ROOT_DIR/deploy/sql/seed/admin.sql"
apply_seed user "$ROOT_DIR/deploy/sql/seed/user.sql"

"${COMPOSE[@]}" restart auth >/dev/null
wait_for_policy_reload

printf '\nSeed completed.\n'
printf 'Default users:\n'
printf '  - vben / 123456  (root)\n'
printf '  - jack / 123456  (admin)\n'
printf '\nIf you need a clean rerun, remove volumes with:\n'
printf '  %s down -v\n' "${COMPOSE[*]}"
