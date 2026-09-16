#!/bin/sh
# Generate a local CA and certificates for Vault and the API. For testing
# only: use certificates from your PKI in production (see deploy/README.md for
# the file ownership the containers need).
set -eu

dir="$(cd "$(dirname "$0")/../certs" && pwd)"
cd "$dir"

openssl req -x509 -newkey rsa:4096 -nodes -days 365 -subj "/CN=SecureEnv local CA" \
  -keyout ca-key.pem -out ca.pem

issue() {
  name="$1"; san="$2"
  openssl req -newkey rsa:2048 -nodes -subj "/CN=$name" -keyout "$name-key.pem" -out "$name.csr"
  printf 'subjectAltName=%s\n' "$san" > "$name.ext"
  openssl x509 -req -in "$name.csr" -CA ca.pem -CAkey ca-key.pem -CAcreateserial \
    -days 365 -extfile "$name.ext" -out "$name.pem"
  rm "$name.csr" "$name.ext"
}

issue vault "DNS:vault,DNS:localhost,IP:127.0.0.1"
issue api "DNS:api,DNS:localhost,IP:127.0.0.1"
# The containers run as their own users (vault: uid 100, API: uid 65532), so
# bind-mounted keys must be readable by them. Acceptable for throwaway test
# certificates only; ca-key.pem is never mounted and stays private.
chmod 644 vault-key.pem api-key.pem
chmod 600 ca-key.pem
echo "Certificates written to $dir" >&2
