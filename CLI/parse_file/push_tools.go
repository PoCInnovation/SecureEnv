package parse_file

import (
	"bufio"
	"os"
	"strings"
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

func GetEnvSecrets() map[string]interface{} {

	lines := linesInFile(".env")
	newData := map[string]interface{}{}
	for _, line := range lines {
		variable := create_var(line)
		if !isVariableSecureEnvLocal(variable) {
			newData[variable.Key] = variable.Value
		}
	}
	return newData
}
