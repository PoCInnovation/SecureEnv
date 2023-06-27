package commands

import (
	"context"
	"fmt"
	"secureenv/parse_file"
	"secureenv/vault_actions"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func is_sync(bodyjson_local map[string]interface{}, bodyjson_server map[string]interface{}) int {
	if len(bodyjson_local) != len(bodyjson_server) {
		return 1
	}

	for key, value := range bodyjson_local {
		serverValue, exists := bodyjson_server[key]
		if !exists || value != serverValue {
			return 1
		}
	}

	return 0
}

func status_print(bodyjson_local map[string]interface{}, bodyjson_server map[string]interface{}, name_project string) {
	fmt.Printf("On project %s\n", name_project)
	if is_sync(bodyjson_local, bodyjson_server) == 0 {
		fmt.Printf("Your project is up to date with %s\n", name_project)
	} else {
		fmt.Printf("Your project is not up to date with %s\n", name_project)
	}
	red := "\033[31m"
	green := "\033[32m"
	reset := "\033[0m"
	for key, value := range bodyjson_local {
		serverValue, exists := bodyjson_server[key]
		if !exists {
			fmt.Printf("\t%s%s%s %s\n", green, "add\t : ", reset, key)
		} else if value != serverValue {
			fmt.Printf("\t%s%s%s %s\n", green, "modified : ", reset, key)
		}
	}

	for key := range bodyjson_server {
		_, exists := bodyjson_local[key]
		if !exists {
			fmt.Printf("\t%s%s%s %s\n", red, "rm\t : ", reset, key)
		}
	}
}

func Status_project(mainUrl string) *ffcli.Command {

	status := &ffcli.Command{
		Name:       "status",
		ShortUsage: "status",
		ShortHelp:  "Get the status of the project between local and vault",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 0 {
				return fmt.Errorf("status requires 0 arguments but you provided %d", n)
			}
			config := parse_file.Parsefile()
			bodyjson_server := vault_actions.Secret_get(config.Project, mainUrl, 0)
			bodyjson_local := parse_file.GetEnvSecrets()
			status_print(bodyjson_local, bodyjson_server, config.Project)
			return nil
		},
	}
	return status
}
