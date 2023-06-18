package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	vault "github.com/hashicorp/vault/api"
)

func List_projects(client *vault.Client) (string, int) {
	secretList, err := client.Logical().List("secret/metadata/")
	if err != nil {
		fmt.Println("unable to read engine:", err)
		return "error read engine", http.StatusInternalServerError
	}

	if len(secretList.Data) == 0 {
		fmt.Println("No Project")
		return "No Project", http.StatusOK
	}

	secretList.Data["projects"] = secretList.Data["keys"]
	delete(secretList.Data, "keys")

	jsonData, err := json.Marshal(secretList.Data)

	return string(jsonData), http.StatusOK
}
