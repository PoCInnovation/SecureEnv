# Read-only access to SecureEnv projects: list, info, status and pull.

path "secret/data/secureenv/*" {
  capabilities = ["read"]
}

path "secret/metadata/secureenv/*" {
  capabilities = ["read", "list"]
}
