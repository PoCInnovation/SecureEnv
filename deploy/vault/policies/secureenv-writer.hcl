# Full access to SecureEnv projects stored under secret/secureenv/.
# Matches SECURE_ENV_VAULT_MOUNT=secret and SECURE_ENV_VAULT_PREFIX=secureenv.

path "secret/data/secureenv/*" {
  capabilities = ["create", "read", "update"]
}

path "secret/metadata/secureenv/*" {
  capabilities = ["read", "list", "delete"]
}
