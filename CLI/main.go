package main

import (
	"cli/vault_actions"
	"context"
	"flag"
	"fmt"
	vault "github.com/hashicorp/vault/api"
	"github.com/peterbourgon/ff/v3/ffcli"
	"log"
	"os"
)

func main() {

	var (
		rootFlagSet = flag.NewFlagSet("textctl", flag.ExitOnError)
	)

	config := vault.DefaultConfig()
	config.Address = "http://127.0.0.1:8200"
	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
	}
	client.SetToken("hvs.XArAWGBtDHkC2TyETD0qTpQV")

	screate := &ffcli.Command{
		Name:       "screate",
		ShortUsage: "screate [<arg> ...]",
		ShortHelp:  "Create a new secret.",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 3 {
				return fmt.Errorf("create requires 3 arguments, name, key and content")
			}
			secretData := map[string]interface{}{
				args[1]: args[2],
			}
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
				return fmt.Errorf("create requires 1 arguments, but you provided %d", n)
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

	root := &ffcli.Command{
		ShortUsage:  "textctl [flags] <subcommand>",
		FlagSet:     rootFlagSet,
		Subcommands: []*ffcli.Command{sdelete, screate, sget},
		Exec: func(context.Context, []string) error {
			return flag.ErrHelp
		},
	}

	if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
