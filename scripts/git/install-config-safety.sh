#!/bin/sh
set -eu

repo_root=$(git rev-parse --show-toplevel)
cd "$repo_root"

git config core.hooksPath .githooks
git config filter.redact-config.clean "sh scripts/git/redact-config-yaml.sh"
git config filter.redact-config.smudge "cat"
git config filter.redact-config.required true

echo "Configured repo-local hooks and redact-config clean filter."
