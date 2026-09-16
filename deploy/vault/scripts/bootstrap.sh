#!/bin/sh
# Configure a fresh Vault for SecureEnv. Run once with a privileged token:
#   VAULT_TOKEN=<root token> /vault/scripts/bootstrap.sh
# then revoke the root token (vault token revoke -self).
set -eu

: "${VAULT_TOKEN:?VAULT_TOKEN must be set to a privileged token}"

if ! vault secrets list -format=json | grep -q '"secret/"'; then
  vault secrets enable -path=secret -version=2 kv
fi

if ! vault audit list -format=json 2>/dev/null | grep -q '"file/"'; then
  vault audit enable file file_path=/vault/logs/audit.log
fi

vault policy write secureenv-writer /vault/policies/secureenv-writer.hcl
vault policy write secureenv-reader /vault/policies/secureenv-reader.hcl

cat >&2 <<'MSG'
Bootstrap complete. Create short-lived tokens for SecureEnv users, e.g.:

  vault token create -policy=secureenv-writer -ttl=720h -display-name=alice
  vault token create -policy=secureenv-reader -ttl=720h -display-name=ci

Then revoke the root token: vault token revoke -self
MSG
