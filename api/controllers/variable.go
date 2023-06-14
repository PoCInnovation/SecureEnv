package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	vault "github.com/hashicorp/vault/api"
)

const adr_vault = "https://secure-env.poc-innovation.com:8200"

var token string = os.Getenv("TOKEN_VPS_VAULT")

func List_vars(name_project string) string {
	// Auth -> Vault

	config := vault.DefaultConfig()

	config.Address = adr_vault

	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
		return "error client auth"
	}

	client.SetToken(token)

	//Ask to engine "secret"
	secret, err := client.KVv2("secret").Get(context.Background(), name_project)
	if err != nil {
		fmt.Println("unable to read secret:", err)
		return "error read secret"
	}

	if len(secret.Data) == 0 {
		fmt.Println("No secret")
		return "No secret"
	}

	jsonData, err := json.Marshal(secret.Data)

	return string(jsonData)
}

func Add_vars(name_project string, var_name string, var_data string) string {
	// Auth -> Vault

	config := vault.DefaultConfig()

	config.Address = adr_vault

	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
		return "error"
	}

	client.SetToken(token)

	var var_struct = map[string]interface{}{
		var_name: var_data,
	}

	//Ask to engine version of project
	var temp_json = List_vars(name_project)

	var data map[string]interface{}
	_ = json.Unmarshal([]byte(temp_json), &data)

	for key, value := range var_struct {
		data[key] = value
	}

	//Send new version concat with old version
	_, err = client.KVv2("secret").Put(context.Background(), name_project, data)
	if err != nil {
		fmt.Println("unable to write secret: ", err)
		return "error"
	}

	return "write successfull"
}

func Edit_vars(name_project string, var_name string, var_data string) string {
	// Auth -> Vault

	config := vault.DefaultConfig()

	config.Address = adr_vault

	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
		return "error"
	}

	client.SetToken(token)

	// Ask the engine version of the project
	var temp_json = List_vars(name_project)

	var data map[string]interface{}
	_ = json.Unmarshal([]byte(temp_json), &data)

	// Edit the variable value
	if _, ok := data[var_name]; ok {
		data[var_name] = var_data
	} else {
		// Variable not found
		return "variable not found"
	}

	// Send the updated data
	_, err = client.KVv2("secret").Put(context.Background(), name_project, data)
	if err != nil {
		fmt.Println("unable to write secret: ", err)
		return "error"
	}

	return "write successful"
}

func Del_vars(name_project string, var_name string) string {
	// Auth -> Vault

	config := vault.DefaultConfig()

	config.Address = adr_vault

	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
		return "error"
	}

	client.SetToken(token)

	// Ask the engine version of the project
	var temp_json = List_vars(name_project)

	var data map[string]interface{}
	_ = json.Unmarshal([]byte(temp_json), &data)

	// Check if the variable exists
	if _, ok := data[var_name]; ok {
		// Remove the variable
		delete(data, var_name)
	} else {
		// Variable not found
		return "variable not found"
	}

	// Send the updated data
	_, err = client.KVv2("secret").Put(context.Background(), name_project, data)
	if err != nil {
		fmt.Println("unable to write secret: ", err)
		return "error"
	}

	return "variable removed successfully"
}
