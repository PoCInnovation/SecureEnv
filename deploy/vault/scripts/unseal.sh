#!/bin/sh
# Unseal Vault after each start. Keys are prompted without echo so they never
# land in shell history or process arguments.
set -eu

while vault status -format=json 2>/dev/null | grep -q '"sealed": true'; do
  vault operator unseal
done
echo "Vault is unsealed." >&2
