package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	vault "github.com/hashicorp/vault/api"
)

func isValidString(input string) bool {
	invalidChars := []string{":", "\\", "/", "*", "?", "\"", "<", ">", "|"}
	if input == "" {
		return false
	}
	for _, char := range invalidChars {
		if strings.Contains(input, char) {
			return false
		}
	}
	return true
}

func List_projects(client *vault.Client) (string, int) {
	secretList, err := client.Logical().List("secret/metadata/")
	if err != nil {
		fmt.Println("unable to read engine:", err)
		return "error read engine", http.StatusInternalServerError
	}

	if len(secretList.Data) == 0 {
		return "No Project", http.StatusOK
	}

	secretList.Data["projects"] = secretList.Data["keys"]
	delete(secretList.Data, "keys")

	jsonData, err := json.Marshal(secretList.Data)

	return string(jsonData), http.StatusOK
}

func Create_project(client *vault.Client, projectName string) (string, int) {
	// Check correct char of name
	if isValidString(projectName) != true {
		return "Project name contains invalid characters", http.StatusNotAcceptable
	}

	// Obtain list of project
	projectList, statusCode := List_projects(client)
	if statusCode != http.StatusOK {
		return projectList, statusCode
	}

	var data map[string]interface{}
	err := json.Unmarshal([]byte(projectList), &data)
	if err != nil {
		fmt.Println("Error decoding project list:", err)
		return "Error decoding project list", http.StatusInternalServerError
	}

	projects, ok := data["projects"].([]interface{})
	if !ok {
		fmt.Println("Invalid project list format")
		return "Invalid project list format", http.StatusInternalServerError
	}

	// Compare list of project to very if already exists
	for _, project := range projects {
		if project.(string) == projectName {
			return "Project already exists", http.StatusConflict
		}
	}

	// Setup Project
	secretData := map[string]interface{}{}
	ctx := context.Background()

	// Write the project into Vault
	_, err = client.KVv2("secret").Put(ctx, projectName, secretData)
	if err != nil {
		fmt.Println("Unable to create project:", err)
		return "Error creating project", http.StatusInternalServerError
	}
	return "Project created successfully", http.StatusCreated
}

func Edit_project(client *vault.Client, projectName string, newProjectName string) (string, int) {
	// Check correct char of name
	if isValidString(projectName) != true {
		return "Project name contains invalid characters", http.StatusNotAcceptable
	}

	// Obtain list of project
	projectList, statusCode := List_projects(client)
	if statusCode != http.StatusOK {
		return projectList, statusCode
	}

	var data map[string]interface{}
	err := json.Unmarshal([]byte(projectList), &data)
	if err != nil {
		fmt.Println("Error decoding project list:", err)
		return "Error decoding project list", http.StatusInternalServerError
	}

	projects, ok := data["projects"].([]interface{})
	if !ok {
		fmt.Println("Invalid project list format:", projects)
		return "Invalid project list format", http.StatusInternalServerError
	}

	// Compare list of project to very if already exists
	for _, project := range projects {
		if project.(string) == newProjectName {
			return "Project already exists", http.StatusConflict
		}
	}

	ctx := context.Background()

	// List all version from project
	list, err := client.KVv2("secret").GetVersionsAsList(ctx, projectName)

	// Write all version from the old project to the new project
	for i := 1; i <= len(list); i++ {
		secret, _ := client.KVv2("secret").GetVersion(ctx, projectName, i)
		_, err = client.KVv2("secret").Put(ctx, newProjectName, secret.Data)
		if err != nil {
			break
		}
	}
	if err != nil {
		fmt.Println("Unable to rename project:", err)
		return "Error while renaming the project", http.StatusInternalServerError
	} else {
		err = client.KVv2("secret").DeleteMetadata(ctx, projectName)
	}
	return "Successfully renamed project", http.StatusCreated
}

func Del_project(client *vault.Client, projectName string) (string, int) {
	// Obtain list of project
	projectList, statusCode := List_projects(client)
	if statusCode != http.StatusOK {
		return projectList, statusCode
	}

	var data map[string]interface{}
	err := json.Unmarshal([]byte(projectList), &data)
	if err != nil {
		fmt.Println("Error decoding project list:", err)
		return "Error decoding project list", http.StatusInternalServerError
	}

	projects, ok := data["projects"].([]interface{})
	if !ok {
		fmt.Println("Invalid project list format:", projects)
		return "Invalid project list format", http.StatusInternalServerError
	}

	ctx := context.Background()

	// Delete project if exist in the list of projects
	for _, project := range projects {
		if project.(string) == projectName {
			err = client.KVv2("secret").DeleteMetadata(ctx, projectName)
			if err != nil {
				fmt.Println("Error deleting project:", err)
				return "Error deleting project", http.StatusInternalServerError
			}
			return "Project deleted successfully", http.StatusOK
		}
	}

	return "Project not Found", http.StatusInternalServerError
}

func Get_project(client *vault.Client, projectName string) (string, int) {
	ctx := context.Background()

	secret, err := client.KVv2("secret").GetMetadata(ctx, projectName)
	if err != nil {
		fmt.Println("Project not Found:", err)
		return "Project not Found", http.StatusInternalServerError
	}

	jsonData, err := json.Marshal(secret)

	return string(jsonData), http.StatusOK
}
