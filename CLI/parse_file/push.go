package parse_file

import (
	"bufio"
	"cli/vault_actions"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

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

func getAllSecrets(name string, mainUrl string) map[string]interface{} {

	url := mainUrl + "/" + name + "/var"

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var data map[string]interface{}
	_ = json.Unmarshal(body, &data)
	return data
}

func Push(mainUrl string) bool {

	config := Parsefile()
	fmt.Printf("Pushing to project %s...\n", config.Project)
	lines := linesInFile(".env")
	newData := map[string]interface{}{}
	vault_actions.Project_create(config.Project, mainUrl)
	for _, line := range lines {
		variable := create_var(line)
		newData[variable.Key] = variable.Value
		if !isVariableSecureEnvLocal(variable) {
			vault_actions.Secret_create(config.Project, variable.Key, variable.Value, mainUrl)
		}
	}
	oldData := getAllSecrets(config.Project, mainUrl)
	for key := range oldData {
		if _, ok := newData[key]; !ok {
			vault_actions.Secret_delete(config.Project, key, mainUrl)
		}
	}
	fmt.Println("Pushed successfully.")
	return true
}

func Push_project(mainUrl string) *ffcli.Command {

	pcreate := &ffcli.Command{
		Name:       "push",
		ShortUsage: "push",
		ShortHelp:  "Push to the project [SECURE_ENV_PROJECT_NAME] all variables written in the .env file execpt SECURE_ENV variables.",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 0 {
				return fmt.Errorf("push requires 0 arguments but you provided %d", n)
			}
			Push(mainUrl)
			return nil
		},
	}
	return pcreate
}
