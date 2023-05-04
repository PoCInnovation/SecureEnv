package vault_actions

import (
	"context"
	"fmt"
	vault "github.com/hashicorp/vault/api"
	"log"
)

func Sget(name string, secretData map[string]interface{}, client *vault.Client) bool {

	_, err := client.KVv2("secret").Put(context.Background(), name, secretData)
	if err != nil {
		log.Fatalf("unable to write secret: %v", err)
	}

	fmt.Println("Secret written successfully.")
	return true
}
