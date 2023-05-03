package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func main() {

	var (
		rootFlagSet = flag.NewFlagSet("textctl", flag.ExitOnError)
	)

	screate := &ffcli.Command{
		Name:       "screate",
		ShortUsage: "screate [<arg> ...]",
		ShortHelp:  "Create a new secret.",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 2 {
				return fmt.Errorf("create requires 2 arguments, but you provided %d", n)
			}
			fmt.Fprintf(os.Stdout, "creating a secret named : %s\ncontent : %s\n", args[0], args[1])
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
			fmt.Fprintf(os.Stdout, "secret named : %s deleted\n", args[0])
			return nil
		},
	}

	sget := &ffcli.Command{
		Name:       "sget",
		ShortUsage: "sget [<arg> ...]",
		ShortHelp:  "get the content of a a secret.",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 1 {
				return fmt.Errorf("create requires 1 arguments, but you provided %d", n)
			}
			fmt.Fprintf(os.Stdout, "secret named '%s' deleted\n", args[0])
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
