package vault_actions

import (
	"context"
	"fmt"
	"github.com/peterbourgon/ff/v3/ffcli"
	"io/ioutil"
	"net/http"
)

func Sdelete_var(mainUrl string) *ffcli.Command {

	sdelete := &ffcli.Command{
		Name:       "sdelete",
		ShortUsage: "sdelete [<arg> ...]",
		ShortHelp:  "Delete a secret.",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 2 {
				return fmt.Errorf("delete requires 2 arguments, name and key but you provided %d", n)
			}
			sdelete(args[0], args[1], mainUrl)
			return nil
		},
	}
	return sdelete
}

func sdelete(name string, key string, mainUrl string) {

	url := mainUrl + "/" + name + "/var/" + key
	println(url)
	req, res := http.NewRequest("DELETE", url, nil)
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
