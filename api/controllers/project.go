package controllers

import (
	"encoding/json"
	"fmt"
	"log"

	vault "github.com/hashicorp/vault/api"
)

func List_projects() string {
	// Auth -> Vault

	config := vault.DefaultConfig()

	config.Address = adr_vault

	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
		return "error"
	}

	client.SetToken(token)

	secretList, err := client.Logical().List("secret/metadata/")
	if err != nil {
		fmt.Println("unable to read engine:", err)
		return "error read engine"
	}

	if len(secretList.Data) == 0 {
		fmt.Println("No secret")
		return "No secret"
	}

	secretList.Data["projects"] = secretList.Data["keys"]
	delete(secretList.Data, "keys")

	jsonData, err := json.Marshal(secretList.Data)

	return string(jsonData)
}
