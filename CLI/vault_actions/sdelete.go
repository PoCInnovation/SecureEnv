package vault_actions

import (
	"context"
	vault "github.com/hashicorp/vault/api"
	"log"
)

func Sdelete(name string, client *vault.Client) {

	_, err := client.KVv2("secret").Get(context.Background(), name)

	if err != nil {
		log.Fatalf("%v", err)
	}

	err = client.KVv2("secret").Delete(context.Background(), name)
	if err != nil {
		log.Fatalf("unable to delete secret: %v", err)
	} else {
		log.Printf("secret %s deleted", name)
	}
}
