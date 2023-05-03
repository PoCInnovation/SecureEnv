package controllers

import (
	"context"
	"fmt"
	"log"
	"os"

	vault "github.com/hashicorp/vault/api"
)

func CallVault(message1 string, message2 string) string {
	var token string = os.Getenv("TOKEN_VAULT")
	config := vault.DefaultConfig()

	config.Address = "http://vault-docker:8200"

	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
		return "error"
	}

	// Authenticate``
	client.SetToken(token)

	secretData := map[string]interface{}{
		message1: message2,
	}

	// Write a secret
	_, err = client.KVv2("secret").Put(context.Background(), "my-secret-password", secretData)
	if err != nil {
		fmt.Println("unable to write secret: ", err)
		return "error"
	}

	fmt.Println("Secret written successfully.")

	// Read a secret from the default mount path for KV v2 in dev mode, "secret"
	secret, err := client.KVv2("secret").Get(context.Background(), "my-secret-password")
	if err != nil {
		fmt.Println("unable to read secret:", err)
		return "error"
	}

	value, ok := secret.Data[message1].(string)
	if !ok {
		fmt.Println("value type assertion failed:", secret.Data[message1], secret.Data[message1])
		return "error"
	}

	if value == "" {
		fmt.Println("No secret")
		return "error"
	}

	fmt.Println("Access granted!")
	return value
}
