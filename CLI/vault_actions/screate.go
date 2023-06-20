package vault_actions

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func Sscreate_var(mainUrl string) *ffcli.Command {

	sdelete := &ffcli.Command{
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
	return sdelete
}

func screate(name string, key string, value string, mainUrl string) {

	url := mainUrl + "/" + name + "/var/" + key
	println(url)
	req, res := http.NewRequest("POST", url, nil)
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
		fmt.Println("La ressource a été supprimée avec succès.")
	} else {
		fmt.Println("La suppression de la ressource a échoué. Code de statut :", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	println(body)
	defer resp.Body.Close()

	println("Secret deleted")
}
