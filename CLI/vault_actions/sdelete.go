package vault_actions

import (
	"context"
	"fmt"
	"log"

	vault "github.com/hashicorp/vault/api"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func Sdelete_var(client *vault.Client) *ffcli.Command {

	sdelete := &ffcli.Command{
		Name:       "sdelete",
		ShortUsage: "sdelete [<arg> ...]",
		ShortHelp:  "Delete a secret.",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 1 {
				return fmt.Errorf("create requires 1 arguments, name and key but you provided %d", n)
			}
			Sdelete(args[0], client)
			return nil
		},
	}
	return sdelete
}

func Sdelete(name string, client *vault.Client) {

	_, err := client.KVv2("secret").Get(context.Background(), name)

	if err != nil {
		log.Fatalf("%v", err)
	}

	err = client.KVv2("secret").Delete(context.Background(), name)
	if err != nil {
		log.Fatalf("unable to delete secret: %v", err)
	} else {
		log.Printf("secret %s deleted", name)
	}
}
