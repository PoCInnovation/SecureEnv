package vault_actions

import (
	"context"
	vault "github.com/hashicorp/vault/api"
	"log"
)

func Sdelete(name string, client *vault.Client) {
	secret, err := client.KVv2("secret").Get(context.Background(), name)

	if err != nil {
		log.Fatalf("unable to read secret: %v", err)
	}

	value, ok := secret.Data[key].(string)
	if !ok {
		log.Fatalf("value type assertion failed: %T %#v", secret.Data[key], secret.Data[key])
	}

	err := client.KVv2("secret").Delete(context.Background(), name)
	if err != nil {
		log.Fatalf("unable to delete secret: %v", err)
	} else {
		log.Printf("secret %s deleted", name)
	}
}
