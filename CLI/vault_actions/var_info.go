package vault_actions

import (
	"context"
	"fmt"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func Get_secret(mainUrl string) *ffcli.Command {

	sget := &ffcli.Command{
		Name:       "sget",
		ShortUsage: "sget [<arg> ...]",
		ShortHelp:  "Get the content of a secret.",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 1 {
				return fmt.Errorf("create requires 1 arguments, name of the project but you provided %d", n)
			}
			secret_get(args[0], mainUrl)
			return nil
		},
	}
	return sget
}

func Create_secret(mainUrl string) *ffcli.Command {

	screate := &ffcli.Command{
		Name:       "screate",
		ShortUsage: "screate [<arg> ...]",
		ShortHelp:  "Create a secret.",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 3 {
				return fmt.Errorf("create requires 3 arguments, name, key and value but you provided %d", n)
			}
			Secret_create(args[0], args[1], args[2], mainUrl)
			return nil
		},
	}
	return screate
}

func Delete_secret(mainUrl string) *ffcli.Command {

	sdelete := &ffcli.Command{
		Name:       "sdelete",
		ShortUsage: "sdelete [<arg> ...]",
		ShortHelp:  "Delete a secret.",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 2 {
				return fmt.Errorf("delete requires 2 arguments, name and key but you provided %d", n)
			}
			Secret_delete(args[0], args[1], mainUrl)
			return nil
		},
	}
	return sdelete
}

func Edit_secret(mainUrl string) *ffcli.Command {

	sedit := &ffcli.Command{
		Name:       "sedit",
		ShortUsage: "sedit [<arg> ...]",
		ShortHelp:  "Edit a secret.",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 3 {
				return fmt.Errorf("edit requires 3 arguments, name, key and value but you provided %d", n)
			}
			secret_edit(args[0], args[1], args[2], mainUrl)
			return nil
		},
	}
	return sedit
}
