package vault_actions

import (
	"context"
	"fmt"
	vault "github.com/hashicorp/vault/api"
	"log"
)

func Screate(name string, secretData map[string]interface{}, client *vault.Client) bool {
	existingSecret, err := client.KVv2("secret").Get(context.Background(), name)
	if err != nil {
		log.Fatalf("unable to read secret: %v", err)
	}

	if existingSecret != nil {
		existingData := existingSecret.Data
		for k, v := range secretData {
			existingData[k] = v
		}

		_, err = client.KVv2("secret").Put(context.Background(), name, existingData)
		if err != nil {
			log.Fatalf("unable to write secret: %v", err)
		}
		fmt.Println("deja un secret existant")

	} else {
		_, err = client.KVv2("secret").Put(context.Background(), name, secretData)
		if err != nil {
			log.Fatalf("unable to write secret: %v", err)
		}
		fmt.Println("pas de secret existant")
	}

	fmt.Println("Secret written successfully.")
	return true
}
