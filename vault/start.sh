#!/bin/bash

if [ -f "/vault/config/.init_data" ]; then
    echo "File '/vault/config/.init_data' already exists. Skipping initialization."
else
    VAULT_CACERT='/vault/certs/cert.pem' VAULT_ADDR='https://secure-env.poc-innovation.com:8200/' vault operator init -format "json" -key-shares 1 -key-threshold 1 > /vault/config/.init_data
fi

VAULT_CACERT='/vault/certs/cert.pem' VAULT_ADDR='https://secure-env.poc-innovation.com:8200/' vault operator unseal $(cat /vault/config/.init_data | jq -r '.unseal_keys_b64[0]')
