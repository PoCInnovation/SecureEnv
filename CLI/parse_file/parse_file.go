package parse_file

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Configuration struct {
	Project string
	Server  []Server_var
}

type Server_var struct {
	Host  string
	Port  int
	Token string
}

func Parsefile() Configuration {
	var config Configuration
	data, err := ioutil.ReadFile(".secure-env")
	if err != nil {
		fmt.Println("File reading error", err)
		return config
	}
	json.Unmarshal(data, &config)
	return config
}
