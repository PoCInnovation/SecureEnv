package vault_actions

import (
	"context"
	"flag"
	"fmt"
	"secureenv/parse_file"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func List_project(mainUrl string) *ffcli.Command {

	plist := &ffcli.Command{
		Name:       "plist",
		ShortUsage: "plist [<arg> ...]",
		ShortHelp:  "List all projects.",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 0 {
				return fmt.Errorf("list requires 0 arguments, but you provided %d", n)
			}
			project_list(mainUrl)
			return nil
		},
	}
	return plist
}

func Get_project(mainUrl string) *ffcli.Command {
	pget := &ffcli.Command{
		Name:       "pget",
		ShortUsage: "pget [<arg> ...]",
		ShortHelp:  "Get the metadata of a project.",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 1 {
				return fmt.Errorf("create requires 1 arguments, name of the project but you provided %d", n)
			}
			project_get(args[0], mainUrl)
			return nil
		},
	}
	return pget
}

func Create_project(mainUrl string) *ffcli.Command {
	pcreate := &ffcli.Command{
		Name:       "pcreate",
		ShortUsage: "pcreate [<arg> ...]",
		ShortHelp:  "Create a project.",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 1 {
				return fmt.Errorf("create requires 1 arguments, name of the project but you provided %d", n)
			}
			project_create(args[0], mainUrl)
			return nil
		},
	}
	return pcreate
}

func Delete_project(mainUrl string) *ffcli.Command {

	pdelete := &ffcli.Command{
		Name:       "pdelete",
		ShortUsage: "pdelete [<arg> ...]",
		ShortHelp:  "Delete a project.",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 1 {
				return fmt.Errorf("delete requires 1 arguments, name of the project but you provided %d", n)
			}
			project_delete(args[0], mainUrl)
			return nil
		},
	}
	return pdelete
}

func Edit_project(mainUrl string) *ffcli.Command {

	pedit := &ffcli.Command{
		Name:       "pedit",
		ShortUsage: "pedit [<arg> ...]",
		ShortHelp:  "Rename a project.",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 2 {
				return fmt.Errorf("edit requires 2 arguments, name, and value but you provided %d", n)
			}
			project_edit(args[0], args[1], mainUrl)
			return nil
		},
	}
	return pedit
}

func Push_project(mainUrl string) *ffcli.Command {

	fs := flag.NewFlagSet("push", flag.ExitOnError)
	forceFlag := fs.Bool("f", false, "Force push")

	ppush := &ffcli.Command{
		Name:       "push",
		ShortUsage: "push [-f]",
		ShortHelp:  "Push to the project [SECURE_ENV_PROJECT_NAME] all variables written in the .env file except SECURE_ENV variables.",
		FlagSet:    fs,
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n > 0 {
				return fmt.Errorf("push requires 0 arguments but you provided %d", n)
			}

			forcePush := *forceFlag

			config := parse_file.Parsefile()
			bodyjson := parse_file.GetEnvSecrets()

			project_update(config.Project, bodyjson, mainUrl, forcePush)
			return nil
		},
	}
	return ppush
}

func Pull_project(mainUrl string) *ffcli.Command {

	ppull := &ffcli.Command{
		Name:       "pull",
		ShortUsage: "pull",
		ShortHelp:  "Pull the project specified in the .env file [SECURE_ENV_PROJECT_NAME] from Vault to the .env file",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 0 {
				return fmt.Errorf("pull requires 0 arguments but you provided %d", n)
			}
			config := parse_file.Parsefile()
			bodyjson := Secret_get(config.Project, mainUrl, 0)
			Project_pull(config, bodyjson)
			return nil
		},
	}
	return ppull
}
