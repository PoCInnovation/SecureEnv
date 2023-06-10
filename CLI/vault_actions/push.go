package vault_actions

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	vault "github.com/hashicorp/vault/api"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type Variable struct {
	Key   string
	Value string
}

func linesInFile(fileName string) []string {
	f, _ := os.Open(fileName)
	scanner := bufio.NewScanner(f)
	result := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		result = append(result, line)
	}
	return result
}

func create_var(data string) Variable {
	var variable Variable
	variable.Key = strings.Split(data, "=")[0]
	variable.Value = strings.Split(data, "=")[1]
	variable.Value = strings.Trim(variable.Value, "\"")
	return variable
}

func isVariableSecureEnvLocal(variable Variable) bool {
	return strings.Contains(variable.Key, "SECURE_ENV_")
}

func Spush_var(client *vault.Client) *ffcli.Command {

	push := &ffcli.Command{
		Name:       "push",
		ShortUsage: "push",
		ShortHelp:  "Push a file to Vault.",
		Exec: func(_ context.Context, args []string) error {
			push(client)
			return nil
		},
	}
	return push
}

func push(client *vault.Client) bool {

	lines := linesInFile(".env")
	secretData := map[string]interface{}{}
	for _, line := range lines {
		variable := create_var(line)
		if !isVariableSecureEnvLocal(variable) {
			secretData[variable.Key] = variable.Value
		}
	}
	Screate("secure-env", secretData, client)
	fmt.Println("Pushed successfully.")
	return true
}
