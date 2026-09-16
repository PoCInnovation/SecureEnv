#!/bin/sh
# Smoke test a running SecureEnv deployment with the CLI.
#
#   SECURE_ENV_TOKEN=<writer token> scripts/smoke.sh <secureenv binary> <API URL>
#
# SECURE_ENV_CA_CERT is honoured for APIs using a private CA. The test project
# is deleted at the end.
set -eu

bin="$1"
export SECURE_ENV_API_URL="$2"
: "${SECURE_ENV_TOKEN:?SECURE_ENV_TOKEN must be set}"

project="smoke-$(date +%s)-$$"
work=$(mktemp -d)
trap '"$bin" project delete -yes "$project" > /dev/null 2>&1 || true; rm -rf "$work"' EXIT
cd "$work"

fail() {
  echo "smoke: $*" >&2
  exit 1
}

"$bin" project create "$project"
"$bin" init "$project"
printf 'DATABASE_URL="postgres://app@db:5432/app?sslmode=require"\n' >> .env
"$bin" push
"$bin" status | grep -q "Up to date." || fail "status after push is not up to date"

printf 'rotated-value' | "$bin" var set API_KEY
[ "$("$bin" var get API_KEY)" = "rotated-value" ] || fail "var get returned an unexpected value"

"$bin" pull
grep -q '^API_KEY="rotated-value"$' .env || fail "pull did not write API_KEY"
"$bin" project list | grep -qx "$project" || fail "project missing from project list"

echo "smoke: OK ($project)"
