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
		ShortUsage: "textctl [flags] <subcommand>",
		FlagSet:    rootFlagSet,
		Subcommands: []*ffcli.Command{
			vault_actions.Create_project(mainUrl),
			vault_actions.Delete_project(mainUrl),
			vault_actions.Edit_project(mainUrl),
			vault_actions.Get_project(mainUrl),
			vault_actions.List_project(mainUrl),
			vault_actions.Create_secret(mainUrl),
			vault_actions.Delete_secret(mainUrl),
			vault_actions.Edit_secret(mainUrl),
			vault_actions.Get_secret(mainUrl)},
		Exec: func(context.Context, []string) error {
			return flag.ErrHelp
		},
	}

	if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
