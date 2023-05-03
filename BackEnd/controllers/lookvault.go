package controllers

import (
	"context"
	"fmt"
	"log"

	vault "github.com/hashicorp/vault/api"
)

func LookVault(message1 string, message2 string) string {
	config := vault.DefaultConfig()

	config.Address = "http://vault-docker:8200"

	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
		return "error"
	}

	// Authenticate
	client.SetToken("dev-only-token")

	// Read a secret from the default mount path for KV v2 in dev mode, "secret"
	secret, err := client.KVv2("secret").Get(context.Background(), "my-secret-password")
	if err != nil {
		fmt.Println("unable to read secret:", err)
		return "error"
	}

	value, ok := secret.Data[message1].(string)
	if !ok {
		fmt.Println("value type assertion failed:", secret.Data[message1], secret.Data[message1])
		return "No key"
	}

	if value == "" {
		fmt.Println("No secret")
		return "No secret"
	}

	fmt.Println("Access granted!")
	return value
}
