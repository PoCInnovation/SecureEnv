# Vault server configuration for SecureEnv.
# https://developer.hashicorp.com/vault/docs/configuration

ui = true

# Containers cannot lock memory (cap_ipc_lock was removed from the official
# image in 2.0.2) and integrated storage recommends disabling mlock anyway.
# Disable swap on the host instead.
disable_mlock = true

storage "raft" {
  path    = "/vault/file"
  node_id = "secureenv-1"
}

listener "tcp" {
  address         = "0.0.0.0:8200"
  tls_cert_file   = "/vault/certs/vault.pem"
  tls_key_file    = "/vault/certs/vault-key.pem"
  tls_min_version = "tls13"
}

api_addr     = "https://vault:8200"
cluster_addr = "https://vault:8201"
