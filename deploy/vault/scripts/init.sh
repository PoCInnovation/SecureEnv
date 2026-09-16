#!/bin/sh
# Initialise Vault once. The unseal keys and the initial root token are
# printed to stdout only: store them in a password manager, never on the
# server. Tune the key split with KEY_SHARES and KEY_THRESHOLD.
set -eu

if vault status -format=json 2>/dev/null | grep -q '"initialized": true'; then
  echo "Vault is already initialized." >&2
  exit 0
fi

vault operator init \
  -key-shares="${KEY_SHARES:-5}" \
  -key-threshold="${KEY_THRESHOLD:-3}"

echo >&2
echo "Save the unseal keys and root token above now, they will not be shown again." >&2
