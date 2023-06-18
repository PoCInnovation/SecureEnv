package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	vault "github.com/hashicorp/vault/api"
)

func List_vars(client *vault.Client, name_project string) (string, int) {
	// Ask to engine "secret"
	secret, err := client.KVv2("secret").Get(context.Background(), name_project)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "secret not found: at") {
			return "project not found", http.StatusNotFound
		}
		fmt.Println("unable to read secret:", err)
		return "error read secret", http.StatusInternalServerError
	}

	if len(secret.Data) == 0 {
		return "No secret", http.StatusOK
	}

	jsonData, err := json.Marshal(secret.Data)

	return string(jsonData), http.StatusOK
}

func Add_vars(client *vault.Client, name_project string, var_name string, var_data string) (string, int) {
	var var_struct = map[string]interface{}{
		var_name: var_data,
	}

	// Ask to engine version of project
	var temp_json, statusCode = List_vars(client, name_project)
	if statusCode == http.StatusNotFound {
		return "project not found", statusCode
	} else if statusCode >= http.StatusBadRequest {
		return "error", statusCode
	}

	// Edit engine
	var data map[string]interface{}
	_ = json.Unmarshal([]byte(temp_json), &data)
	for key, value := range var_struct {
		data[key] = value
	}

	// Send new version concat with old version
	_, err := client.KVv2("secret").Put(context.Background(), name_project, data)
	if err != nil {
		fmt.Println("unable to write secret: ", err)
		return "error", http.StatusInternalServerError
	}

	return "write successfull", http.StatusOK
}

func Edit_vars(client *vault.Client, name_project string, var_name string, var_data string) (string, int) {
	// Ask the engine version of the project
	var temp_json, statusCode = List_vars(client, name_project)
	if statusCode == http.StatusNotFound {
		return "project not found", statusCode
	} else if statusCode >= http.StatusBadRequest {
		return "error", statusCode
	}

	// Edit the engine
	var data map[string]interface{}
	_ = json.Unmarshal([]byte(temp_json), &data)

	// Edit var
	if _, ok := data[var_name]; ok {
		data[var_name] = var_data
	} else {
		// Variable not found
		return "variable not found", http.StatusBadRequest
	}

	// Send the updated data
	_, err := client.KVv2("secret").Put(context.Background(), name_project, data)
	if err != nil {
		fmt.Println("unable to write secret: ", err)
		return "error", http.StatusInternalServerError
	}

	return "write successful", http.StatusOK
}

func Del_vars(client *vault.Client, name_project string, var_name string) (string, int) {
	// Ask the engine version of the project
	var temp_json, statusCode = List_vars(client, name_project)
	if statusCode == http.StatusNotFound {
		return "project not found", statusCode
	} else if statusCode >= http.StatusBadRequest {
		return "error", statusCode
	}

	var data map[string]interface{}
	_ = json.Unmarshal([]byte(temp_json), &data)

	// Check if the variable exists
	if _, ok := data[var_name]; ok {
		delete(data, var_name)
	} else {
		return "variable not found", http.StatusBadRequest
	}

	// Send the updated data
	_, err := client.KVv2("secret").Put(context.Background(), name_project, data)
	if err != nil {
		fmt.Println("unable to write secret: ", err)
		return "error", http.StatusInternalServerError
	}

	return "variable removed successfully", http.StatusOK
}
