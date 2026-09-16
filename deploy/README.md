# Deployment

`compose.prod.yaml` runs Vault (integrated storage, TLS 1.3, audit log) and the API. It follows the [Vault production hardening](https://developer.hashicorp.com/vault/docs/concepts/production-hardening) baseline where a single container host allows it.

## 1. Certificates

Use certificates from your PKI and place them in `vault/certs/` (`ca.pem`, `vault.pem`, `vault-key.pem`, `api.pem`, `api-key.pem`). To try things out locally:

```bash
./vault/scripts/gen-certs.sh
```

Certificates are git-ignored. Also make sure `api_addr` and `cluster_addr` in `vault/config.hcl` match your hostnames.

## 2. Start, initialize and unseal

```bash
docker compose -f compose.prod.yaml up -d vault
docker compose -f compose.prod.yaml exec vault /vault/scripts/init.sh     # once: prints unseal keys and root token
docker compose -f compose.prod.yaml exec -it vault /vault/scripts/unseal.sh  # after every start
```

`init.sh` only prints the unseal keys and the root token. Store them in a password manager and hand the key shares to different people.

## 3. Bootstrap

```bash
docker compose -f compose.prod.yaml exec -it -e VAULT_TOKEN vault /vault/scripts/bootstrap.sh
```

This enables KV v2 at `secret/`, a file audit device at `/vault/logs/audit.log`, and two policies scoped to `secret/*/secureenv/*`:

- `secureenv-writer` can do everything the CLI can do.
- `secureenv-reader` can list, inspect, get and pull, but not change anything.

Create short-lived tokens for users and CI, then revoke the root token:

```bash
vault token create -policy=secureenv-writer -ttl=720h -display-name=alice
vault token revoke -self
```

For machines, prefer the [AppRole](https://developer.hashicorp.com/vault/docs/auth/approle) auth method, and hand the resulting token to the CLI through `SECURE_ENV_TOKEN`.

## 4. Start the API

```bash
docker compose -f compose.prod.yaml up -d api
curl --cacert vault/certs/ca.pem https://localhost:8080/readyz
```

## Host recommendations

- Disable swap: containers cannot lock memory, so `disable_mlock = true` is set.
- Restrict network access to ports 8200 and 8080, and back up the `vault-data` volume (`vault operator raft snapshot save`).
- Ship `vault-logs` to your log platform and keep access to it restricted.
