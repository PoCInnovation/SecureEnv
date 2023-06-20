package vault_actions

import (
	"context"
	"fmt"
	"github.com/peterbourgon/ff/v3/ffcli"
	"io/ioutil"
	"log"
	"net/http"
)

func Sget_var(mainUrl string) *ffcli.Command {

	sget := &ffcli.Command{
		Name:       "sget",
		ShortUsage: "sget [<arg> ...]",
		ShortHelp:  "get the content of a secret.",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 1 {
				return fmt.Errorf("create requires 1 arguments, name of the project but you provided %d", n)
			}
			sget(args[0], mainUrl)
			return nil
		},
	}
	return sget
}

func sget(name string, mainUrl string) {

	url := mainUrl + "/" + name + "/var"

	println(url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	sb := string(body)
	log.Printf(sb)
}
