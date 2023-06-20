package main

import (
	"cli/vault_actions"
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func main() {

	rootFlagSet := flag.NewFlagSet("textctl", flag.ExitOnError)

	mainUrl := "http://0.0.0.0:8080/project"

	root := &ffcli.Command{
		ShortUsage:  "textctl [flags] <subcommand>",
		FlagSet:     rootFlagSet,
		Subcommands: []*ffcli.Command{vault_actions.Sget_var(mainUrl), vault_actions.Sdelete_var(mainUrl), vault_actions.Screate_var(mainUrl)},
		Exec: func(context.Context, []string) error {
			return flag.ErrHelp
		},
	}

	if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
