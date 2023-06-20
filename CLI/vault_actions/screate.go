package vault_actions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func Screate_var(mainUrl string) *ffcli.Command {

	screate := &ffcli.Command{
		Name:       "screate",
		ShortUsage: "screate [<arg> ...]",
		ShortHelp:  "Create a secret.",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 3 {
				return fmt.Errorf("create requires 3 arguments, name, key and value but you provided %d", n)
			}
			screate(args[0], args[1], args[2], mainUrl)
			return nil
		},
	}
	return screate
}

func screate(name string, key string, value string, mainUrl string) {

	url := mainUrl + "/" + name + "/var/" + key
	println(url)
	bodyjson := map[string]interface{}{
		"Value": value,
	}
	jsonData, err := json.Marshal(bodyjson)
	bodyBuffer := bytes.NewBuffer(jsonData)
	req, res := http.NewRequest("POST", url, bodyBuffer)
	if res != nil {
		fmt.Println(res)
		return
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Secret créé avec succès.")
	} else {
		fmt.Println("La création du secret à échoué. Code de statut :", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	println(body)
	defer resp.Body.Close()

	println("Secret craeted successfully")
}
