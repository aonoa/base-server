#!/usr/bin/env bash

set -euo pipefail

AUTH_DB_URL="${AUTH_DB_URL:-postgresql://postgres:postgres.666@127.0.0.1:25432/auth?sslmode=disable}"
ADMIN_DB_URL="${ADMIN_DB_URL:-postgresql://postgres:postgres.666@127.0.0.1:25432/admin?sslmode=disable}"

if ! command -v psql >/dev/null 2>&1; then
  echo "psql is required but was not found in PATH" >&2
  exit 1
fi

tmp_dir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT

copy_table() {
  local table="$1"
  local columns="$2"
  local file="$tmp_dir/${table}.csv"

  echo "Exporting ${table} from auth database..."
  psql "$AUTH_DB_URL" -v ON_ERROR_STOP=1 -c "\\copy (SELECT ${columns} FROM ${table}) TO '${file}' WITH (FORMAT csv)"

  echo "Importing ${table} into admin database..."
  psql "$ADMIN_DB_URL" -v ON_ERROR_STOP=1 -c "\\copy ${table} (${columns}) FROM '${file}' WITH (FORMAT csv)"
}

echo "Checking source and target database connectivity..."
psql "$AUTH_DB_URL" -v ON_ERROR_STOP=1 -c "SELECT current_database();" >/dev/null
psql "$ADMIN_DB_URL" -v ON_ERROR_STOP=1 -c "SELECT current_database();" >/dev/null

echo "Replacing permission data in admin database..."
psql "$ADMIN_DB_URL" -v ON_ERROR_STOP=1 <<'SQL'
TRUNCATE TABLE
  api_resources_roles,
  resource_roles,
  sys_api_resources,
  sys_resources,
  sys_role;
SQL

copy_table "sys_role" "\"id\", \"create_time\", \"update_time\", \"name\", \"value\", \"status\", \"desc\", \"menus\""
copy_table "sys_resources" "\"id\", \"create_time\", \"update_time\", \"name\", \"type\", \"value\", \"method\", \"description\""
copy_table "sys_api_resources" "\"id\", \"create_time\", \"update_time\", \"description\", \"path\", \"method\", \"module\", \"module_description\", \"resources_group\""
copy_table "resource_roles" "\"resource_id\", \"role_id\""
copy_table "api_resources_roles" "\"api_resources_id\", \"role_id\""

echo "Migration completed."
echo "Copied tables:"
for table in sys_role sys_resources sys_api_resources resource_roles api_resources_roles; do
  echo "  - ${table}"
done
