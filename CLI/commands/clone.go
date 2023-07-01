package commands

import (
	"context"
	"fmt"
	"secureenv/vault_actions"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func Clone_project() *ffcli.Command {

	status := &ffcli.Command{
		Name:       "clone",
		ShortUsage: "clone",
		ShortHelp:  "Clone the project with name value",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 1 {
				return fmt.Errorf("clone requires 1 arguments, api address, but you provided %d", n)
			}
			url := args[0] + "/project"
			_, err := vault_actions.Project_clone(url)
			if err != nil {
				return err
			}
			return nil
		},
	}
	return status
}
