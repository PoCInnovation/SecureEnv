package main

import (
	"cli/parse_file"
	"cli/vault_actions"
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	vault "github.com/hashicorp/vault/api"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func main() {

	var (
		rootFlagSet = flag.NewFlagSet("textctl", flag.ExitOnError)
	)

	secure_env := parse_file.Parsefile()

	config := vault.DefaultConfig()
	config.Address = secure_env.Host
	client, err := vault.NewClient(config)

	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
	}
	client.SetToken(secure_env.Token)

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
			vault_actions.Screate(args[0], secretData, client)
			return nil
		},
	}

	sdelete := &ffcli.Command{
		Name:       "sdelete",
		ShortUsage: "sdelete [<arg> ...]",
		ShortHelp:  "Delete a secret.",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 1 {
				return fmt.Errorf("create requires 1 arguments, name and key but you provided %d", n)
			}
			vault_actions.Sdelete(args[0], client)
			return nil
		},
	}

	sget := &ffcli.Command{
		Name:       "sget",
		ShortUsage: "sget [<arg> ...]",
		ShortHelp:  "get the content of a a secret.",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 2 {
				return fmt.Errorf("create requires 2 arguments, name and key")
			}
			vault_actions.Sget(args[0], args[1], client)
			return nil
		},
	}

	push := &ffcli.Command{
		Name:       "push",
		ShortUsage: "push",
		ShortHelp:  "Push a file to Vault.",
		Exec: func(_ context.Context, args []string) error {
			vault_actions.Push(client)
			return nil
		},
	}

	root := &ffcli.Command{
		ShortUsage:  "textctl [flags] <subcommand>",
		FlagSet:     rootFlagSet,
		Subcommands: []*ffcli.Command{sdelete, screate, sget, push},
		Exec: func(context.Context, []string) error {
			return flag.ErrHelp
		},
	}

	if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
