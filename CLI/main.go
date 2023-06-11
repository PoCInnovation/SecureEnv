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

	rootFlagSet := flag.NewFlagSet("textctl", flag.ExitOnError)

	secure_env := parse_file.Parsefile()

	config := vault.DefaultConfig()
	config.Address = secure_env.Host
	client, err := vault.NewClient(config)

	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
	}
	client.SetToken(secure_env.Token)

	root := &ffcli.Command{
		ShortUsage: "textctl [flags] <subcommand>",
		FlagSet:    rootFlagSet,
		Subcommands: []*ffcli.Command{vault_actions.Sdelete_var(client), vault_actions.Screate_var(client),
			vault_actions.Sget_var(client), vault_actions.Spush_var(client)},
		Exec: func(context.Context, []string) error {
			return flag.ErrHelp
		},
	}

	if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
