package cli

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"
)

func (a *app) projectListCommand() *command {
	return &command{
		name:    "list",
		summary: "List projects",
		run: func(ctx context.Context, _ []string) error {
			api, err := a.api()
			if err != nil {
				return err
			}
			names, err := api.ListProjects(ctx)
			if err != nil {
				return err
			}
			for _, name := range names {
				fmt.Fprintln(a.env.Stdout, name)
			}
			return nil
		},
	}
}

func (a *app) projectCreateCommand() *command {
	return &command{
		name:    "create",
		args:    "[name]",
		summary: "Create a project (default: current project)",
		maxArgs: 1,
		run: func(ctx context.Context, args []string) error {
			name, err := a.projectName(ctx, first(args))
			if err != nil {
				return err
			}
			api, err := a.api()
			if err != nil {
				return err
			}
			if err := api.CreateProject(ctx, name); err != nil {
				return err
			}
			fmt.Fprintf(a.env.Stdout, "Created project %s\n", name)
			return nil
		},
	}
}

func (a *app) projectInfoCommand() *command {
	return &command{
		name:    "info",
		args:    "[name]",
		summary: "Show project metadata (default: current project)",
		maxArgs: 1,
		run: func(ctx context.Context, args []string) error {
			name, err := a.projectName(ctx, first(args))
			if err != nil {
				return err
			}
			api, err := a.api()
			if err != nil {
				return err
			}
			info, err := api.Project(ctx, name)
			if err != nil {
				return err
			}
			fmt.Fprintf(a.env.Stdout, "Name:     %s\nVersion:  %d\nCreated:  %s\nUpdated:  %s\n",
				info.Name, info.CurrentVersion, info.CreatedAt.Format(time.RFC3339), info.UpdatedAt.Format(time.RFC3339))
			return nil
		},
	}
}

func (a *app) projectRenameCommand() *command {
	return &command{
		name:    "rename",
		args:    "<old> <new>",
		summary: "Rename a project, keeping its history",
		minArgs: 2,
		maxArgs: 2,
		run: func(ctx context.Context, args []string) error {
			api, err := a.api()
			if err != nil {
				return err
			}
			if err := api.RenameProject(ctx, args[0], args[1]); err != nil {
				return err
			}
			fmt.Fprintf(a.env.Stdout, "Renamed project %s to %s\n", args[0], args[1])
			return nil
		},
	}
}

func (a *app) projectDeleteCommand() *command {
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	yes := fs.Bool("yes", false, "do not ask for confirmation")

	return &command{
		name:    "delete",
		args:    "<name>",
		summary: "Permanently delete a project and its history",
		flags:   fs,
		minArgs: 1,
		maxArgs: 1,
		run: func(ctx context.Context, args []string) error {
			name := args[0]
			if !*yes {
				if err := a.confirm(fmt.Sprintf("This permanently deletes %s and all its versions. Type the project name to confirm: ", name), name); err != nil {
					return err
				}
			}
			api, err := a.api()
			if err != nil {
				return err
			}
			if err := api.DeleteProject(ctx, name); err != nil {
				return err
			}
			fmt.Fprintf(a.env.Stdout, "Deleted project %s\n", name)
			return nil
		},
	}
}

var errAborted = errors.New("aborted")

func (a *app) confirm(prompt, expected string) error {
	fmt.Fprint(a.env.Stderr, prompt)
	line, err := bufio.NewReader(a.env.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return errAborted
	}
	if strings.TrimSpace(line) != expected {
		return errAborted
	}
	return nil
}

func first(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}
