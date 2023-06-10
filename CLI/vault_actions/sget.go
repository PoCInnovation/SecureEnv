package vault_actions

import (
	"context"
	"fmt"
	"log"

	vault "github.com/hashicorp/vault/api"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func Sget_var(client *vault.Client) *ffcli.Command {

	sget := &ffcli.Command{
		Name:       "sget",
		ShortUsage: "sget [<arg> ...]",
		ShortHelp:  "get the content of a secret.",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 2 {
				return fmt.Errorf("create requires 2 arguments, name and key")
			}
			sget(args[0], args[1], client)
			return nil
		},
	}
	return sget
}

func sget(name string, key string, client *vault.Client) {

	secret, err := client.KVv2("secret").Get(context.Background(), name)

	if err != nil {
		log.Fatalf("unable to read secret: %v", err)
	}

	value, ok := secret.Data[key].(string)
	if !ok {
		log.Fatalf("value type assertion failed: %T %#v", secret.Data[key], secret.Data[key])
	}

	fmt.Println("Access granted! the value of the secret is: ", value)
}
