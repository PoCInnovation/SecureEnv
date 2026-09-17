//go:build integration

package vaultstore_test

import (
	"context"
	"crypto/rand"
	"os"
	"testing"

	vault "github.com/hashicorp/vault/api"

	"github.com/PoCInnovation/SecureEnv/internal/project"
	"github.com/PoCInnovation/SecureEnv/internal/project/storetest"
	"github.com/PoCInnovation/SecureEnv/internal/vaultstore"
)

// TestVaultContract runs the store contract against a real Vault server.
// It needs VAULT_ADDR and VAULT_TOKEN, e.g. a dev server:
//
//	docker run --rm -p 8200:8200 -e VAULT_DEV_ROOT_TOKEN_ID=root hashicorp/vault
//	VAULT_ADDR=http://127.0.0.1:8200 VAULT_TOKEN=root go test -tags integration ./internal/vaultstore/
func TestVaultContract(t *testing.T) {
	if os.Getenv("VAULT_ADDR") == "" || os.Getenv("VAULT_TOKEN") == "" {
		t.Skip("VAULT_ADDR and VAULT_TOKEN are required")
	}

	client, err := vault.NewClient(vault.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}

	storetest.Run(t, func(t *testing.T) project.Store {
		// Every test gets its own folder so runs never collide.
		prefix := "secureenv-test/" + rand.Text()
		store := vaultstore.New(client, vaultstore.Options{Prefix: prefix})
		t.Cleanup(func() { cleanup(t, client, prefix) })
		return store
	})
}

func cleanup(t *testing.T, client *vault.Client, prefix string) {
	t.Helper()
	// t.Context is already cancelled when cleanup functions run.
	ctx := context.Background()
	list, err := client.Logical().ListWithContext(ctx, "secret/metadata/"+prefix)
	if err != nil || list == nil {
		return
	}
	keys, _ := list.Data["keys"].([]any)
	for _, key := range keys {
		_ = client.KVv2("secret").DeleteMetadata(ctx, prefix+"/"+key.(string))
	}
}
