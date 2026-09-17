package cli

import (
	"flag"
)

func (a *app) commands() *command {
	global := flag.NewFlagSet("secureenv", flag.ContinueOnError)
	global.StringVar(&a.apiURL, "api", "", "API URL (env "+KeyAPIURL+", default "+defaultAPIURL+")")
	global.StringVar(&a.project, "project", "", "project name (env "+KeyProject+", default from the git origin remote)")
	global.StringVar(&a.file, "file", ".env", "path of the local env file")

	root := &command{
		name:    "secureenv",
		summary: "Manage and synchronise project environment variables stored in HashiCorp Vault.",
		flags:   global,
	}

	return root.add(
		a.initCommand(),
		a.cloneCommand(),
		a.statusCommand(),
		a.pullCommand(),
		a.pushCommand(),
		(&command{name: "project", summary: "Manage projects"}).add(
			a.projectListCommand(),
			a.projectCreateCommand(),
			a.projectInfoCommand(),
			a.projectRenameCommand(),
			a.projectDeleteCommand(),
		),
		(&command{name: "var", summary: "Manage the variables of the current project"}).add(
			a.varListCommand(),
			a.varGetCommand(),
			a.varSetCommand(),
			a.varUnsetCommand(),
		),
		a.versionCommand(),
	)
}
