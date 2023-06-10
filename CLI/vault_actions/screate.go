package vault_actions

import (
	"context"
	"fmt"
	"log"

	vault "github.com/hashicorp/vault/api"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func Screate_var(client *vault.Client) *ffcli.Command {

	screate := &ffcli.Command{
		Name:       "screate",
		ShortUsage: "screate [<arg> ...]",
		ShortHelp:  "Create a new secret.",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 3 {
				return fmt.Errorf("create requires 3 arguments, name, key and content")
			}

			secret, err := client.KVv2("secret").Get(context.Background(), args[0])
			if err != nil {
				log.Fatalf("unable to read secret: %v", err)
			}

			secretData := secret.Data
			secretData[args[1]] = args[2]
			fmt.Println(secretData)
			Screate(args[0], secretData, client)
			return nil
		},
	}
	return screate
}

func Screate(name string, secretData map[string]interface{}, client *vault.Client) bool {

	_, err := client.KVv2("secret").Put(context.Background(), name, secretData)
	if err != nil {
		log.Fatalf("unable to write secret: %v", err)
	}

	fmt.Println("Secret written successfully.")
	return true
}
