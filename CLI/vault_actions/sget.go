package vault_actions

import (
	"context"
	"fmt"
	vault "github.com/hashicorp/vault/api"
	"log"
)

func Sget(name string, key string, client *vault.Client) {

	secret, err := client.KVv2("secret").Get(context.Background(), name)

	if err != nil {
		log.Fatalf("unable to read secret: %v", err)
	}

	value, ok := secret.Data[key].(string)
	if !ok {
		log.Fatalf("value type assertion failed: %T %#v", secret.Data[key], secret.Data[key])
	}

	fmt.Println("Access granted! the value of the secret is: ", value)
}