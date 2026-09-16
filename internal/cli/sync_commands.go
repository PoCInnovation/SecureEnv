package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"

	"github.com/PoCInnovation/SecureEnv/internal/domain"
	"github.com/PoCInnovation/SecureEnv/internal/dotenv"
	"github.com/PoCInnovation/SecureEnv/internal/envsync"
)

func (a *app) initCommand() *command {
	return &command{
		name:    "init",
		args:    "[project]",
		summary: "Link the local env file to a project (default: derived from the git origin remote)",
		maxArgs: 1,
		run: func(ctx context.Context, args []string) error {
			_, err := a.link(ctx, first(args))
			return err
		},
	}
}

func (a *app) cloneCommand() *command {
	return &command{
		name:    "clone",
		args:    "[project]",
		summary: "Link the local env file to a project and pull its variables",
		maxArgs: 1,
		run: func(ctx context.Context, args []string) error {
			project, err := a.link(ctx, first(args))
			if err != nil {
				return err
			}
			return a.pull(ctx, project, false)
		},
	}
}

func (a *app) statusCommand() *command {
	return &command{
		name:    "status",
		summary: "Compare the local env file with the project",
		run: func(ctx context.Context, _ []string) error {
			project, syncer, err := a.syncer(ctx)
			if err != nil {
				return err
			}
			result, err := syncer.Status(ctx, project)
			if err != nil {
				return err
			}
			fmt.Fprintf(a.env.Stdout, "Project %s (remote version %d)\n", project, result.RemoteVersion)
			if result.Changes.InSync() {
				fmt.Fprintln(a.env.Stdout, "Up to date.")
				return nil
			}
			a.printChanges(result.Changes)
			return nil
		},
	}
}

func (a *app) pullCommand() *command {
	fs := flag.NewFlagSet("pull", flag.ContinueOnError)
	force := fs.Bool("force", false, "overwrite local changes that were not pushed")

	return &command{
		name:    "pull",
		summary: "Write the project variables to the local env file",
		flags:   fs,
		run: func(ctx context.Context, _ []string) error {
			project, err := a.projectName(ctx, "")
			if err != nil {
				return err
			}
			return a.pull(ctx, project, *force)
		},
	}
}

func (a *app) pushCommand() *command {
	fs := flag.NewFlagSet("push", flag.ContinueOnError)
	force := fs.Bool("force", false, "delete remote variables missing from the local env file")

	return &command{
		name:    "push",
		summary: "Replace the project variables with the local env file",
		flags:   fs,
		run: func(ctx context.Context, _ []string) error {
			project, syncer, err := a.syncer(ctx)
			if err != nil {
				return err
			}
			result, err := syncer.Push(ctx, project, *force)
			if err != nil {
				return err
			}
			if result.Changes.InSync() {
				fmt.Fprintf(a.env.Stdout, "Nothing to push, %s is up to date (version %d)\n", project, result.RemoteVersion)
				return nil
			}
			a.printChanges(result.Changes)
			fmt.Fprintf(a.env.Stdout, "Pushed %s (version %d)\n", project, result.RemoteVersion)
			return nil
		},
	}
}

func (a *app) versionCommand() *command {
	return &command{
		name:    "version",
		summary: "Print the version",
		run: func(context.Context, []string) error {
			fmt.Fprintln(a.env.Stdout, "secureenv", a.env.Version)
			return nil
		},
	}
}

// link records the project in the local env file.
func (a *app) link(ctx context.Context, explicit string) (string, error) {
	project := explicit
	if project == "" {
		var err error
		if project, err = a.projectFromGit(ctx); err != nil {
			return "", err
		}
	}
	if _, err := domain.NewProjectName(project); err != nil {
		return "", err
	}

	file, err := dotenv.Load(a.envPath())
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	values := file.Values()
	delete(values, keyLegacyProject)
	values[KeyProject] = project
	if a.apiURL != "" {
		values[KeyAPIURL] = a.apiURL
	}
	if err := dotenv.Save(a.envPath(), dotenv.New(values)); err != nil {
		return "", err
	}
	fmt.Fprintf(a.env.Stdout, "Linked %s to project %s\n", a.file, project)
	return project, nil
}

func (a *app) pull(ctx context.Context, project string, force bool) error {
	api, err := a.api()
	if err != nil {
		return err
	}
	result, err := envsync.New(api, a.envPath()).Pull(ctx, project, force)
	if err != nil {
		return err
	}
	fmt.Fprintf(a.env.Stdout, "Pulled %s (version %d) into %s\n", project, result.RemoteVersion, a.file)
	return nil
}

func (a *app) syncer(ctx context.Context) (string, *envsync.Syncer, error) {
	project, api, err := a.projectAndAPI(ctx)
	if err != nil {
		return "", nil, err
	}
	return project, envsync.New(api, a.envPath()), nil
}

func (a *app) printChanges(changes domain.Changes) {
	groups := []struct {
		symbol, label string
		keys          []string
	}{
		{"+", "local only", changes.Added},
		{"~", "modified locally", changes.Modified},
		{"-", "remote only", changes.Removed},
	}
	for _, group := range groups {
		for _, key := range group.keys {
			fmt.Fprintf(a.env.Stdout, "  %s %-32s %s\n", group.symbol, key, group.label)
		}
	}
}
