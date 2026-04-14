#!/usr/bin/env bash

set -euo pipefail

USER_DB_URL="${USER_DB_URL:-postgresql://postgres:postgres.666@127.0.0.1:25432/user?sslmode=disable}"

if ! command -v psql >/dev/null 2>&1; then
  echo "psql is required but was not found in PATH" >&2
  exit 1
fi

psql "$USER_DB_URL" -v ON_ERROR_STOP=1 -c "ALTER TABLE sys_user DROP COLUMN IF EXISTS role_id;"

echo "Dropped sys_user.role_id from user database."
