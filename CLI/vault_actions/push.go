package vault_actions

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	vault "github.com/hashicorp/vault/api"
)

type Variable struct {
	Key   string
	Value string
}

func LinesInFile(fileName string) []string {
	f, _ := os.Open(fileName)
	scanner := bufio.NewScanner(f)
	result := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		result = append(result, line)
	}
	return result
}

func Create_var(data string) Variable {
	var variable Variable
	variable.Key = strings.Split(data, "=")[0]
	variable.Value = strings.Split(data, "=")[1]
	variable.Value = strings.Trim(variable.Value, "\"")
	return variable
}

func IsVariableSecureEnvLocal(variable Variable) bool {
	return strings.Contains(variable.Key, "SECURE_ENV_")
}

func Push(client *vault.Client) bool {

	lines := LinesInFile("secure-env.env")
	secretData := map[string]interface{}{}
	for _, line := range lines {
		variable := Create_var(line)
		if !IsVariableSecureEnvLocal(variable) {
			secretData[variable.Key] = variable.Value
		}
	}
	Screate("secure-env", secretData, client)
	fmt.Println("Pushed successfully.")
	return true
}
