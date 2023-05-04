package vault_actions

import (
	"context"
	vault "github.com/hashicorp/vault/api"
	"log"
)

func Sdelete(name string, client *vault.Client) {

	err := client.KVv2("secret").Delete(context.Background(), name)
	if err != nil {
		log.Fatalf("unable to delete secret: %v", err)
	} else {
		log.Printf("secret %s deleted", name)
	}
}
