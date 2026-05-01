#!/usr/bin/env bash

set -euo pipefail

USER_DB_URL="${USER_DB_URL:-postgresql://postgres:postgres.666@127.0.0.1:25432/user?sslmode=disable}"
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

file="$tmp_dir/user_role_bindings.csv"

echo "Checking source and target database connectivity..."
psql "$USER_DB_URL" -v ON_ERROR_STOP=1 -c "SELECT current_database();" >/dev/null
psql "$ADMIN_DB_URL" -v ON_ERROR_STOP=1 -c "SELECT current_database();" >/dev/null

echo "Exporting non-empty user role bindings from user database..."
psql "$USER_DB_URL" -v ON_ERROR_STOP=1 -c "\\copy (SELECT id::text AS user_id, role_id FROM sys_user WHERE role_id IS NOT NULL) TO '${file}' WITH (FORMAT csv)"

echo "Replacing admin user-role binding data..."
psql "$ADMIN_DB_URL" -v ON_ERROR_STOP=1 <<'SQL'
TRUNCATE TABLE sys_user_role_binding RESTART IDENTITY;
CREATE TEMP TABLE tmp_user_role_binding (
  user_id text,
  role_id bigint
);
SQL

psql "$ADMIN_DB_URL" -v ON_ERROR_STOP=1 -c "\\copy tmp_user_role_binding (user_id, role_id) FROM '${file}' WITH (FORMAT csv)"

psql "$ADMIN_DB_URL" -v ON_ERROR_STOP=1 <<'SQL'
INSERT INTO sys_user_role_binding (user_id, role_id, create_time, update_time)
SELECT user_id, role_id, NOW(), NOW()
FROM tmp_user_role_binding
ON CONFLICT (user_id, role_id) DO NOTHING;
SQL

echo "Migration completed."
