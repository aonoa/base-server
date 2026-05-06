#!/bin/sh
set -eu

awk '
function indent_of(line) {
  match(line, /^[[:space:]]*/)
  return RLENGTH
}

function spaces(n, out, i) {
  out = ""
  for (i = 0; i < n; i++) {
    out = out " "
  }
  return out
}

/^[[:space:]]*filename:[[:space:]]*/ {
  print spaces(indent_of($0)) "filename: ./logs/app.log"
  next
}

/^[[:space:]]*source:[[:space:]]*/ {
  print spaces(indent_of($0)) "source: \"postgresql://postgres:CHANGE_ME_DB_PASSWORD@127.0.0.1:5432/test1?sslmode=disable\""
  next
}

/^[[:space:]]*service_key:[[:space:]]*/ {
  print spaces(indent_of($0)) "service_key: CHANGE_ME_AUTH_SERVICE_KEY"
  next
}

/^[[:space:]]*api_key:[[:space:]]*/ {
  if (indent_of($0) <= 2) {
    print spaces(indent_of($0)) "api_key: CHANGE_ME_FRONTEND_API_KEY"
  } else {
    print spaces(indent_of($0)) "api_key: CHANGE_ME_LLM_API_KEY"
  }
  next
}

{ print }
'
