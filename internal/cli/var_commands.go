package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/PoCInnovation/SecureEnv/internal/dotenv"
)

const maxStdinValue = 1 << 20

func (a *app) varListCommand() *command {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	values := fs.Bool("values", false, "print values in .env format instead of names only")

	return &command{
		name:    "list",
		summary: "List the variables of the current project",
		flags:   fs,
		run: func(ctx context.Context, _ []string) error {
			project, api, err := a.projectAndAPI(ctx)
			if err != nil {
				return err
			}
			snapshot, err := api.Variables(ctx, project)
			if err != nil {
				return err
			}
			if *values {
				return dotenv.New(snapshot.Variables.Map()).Format(a.env.Stdout)
			}
			for _, key := range snapshot.Variables.Keys() {
				fmt.Fprintln(a.env.Stdout, key)
			}
			return nil
		},
	}
}

func (a *app) varGetCommand() *command {
	return &command{
		name:    "get",
		args:    "<key>",
		summary: "Print the value of a variable",
		minArgs: 1,
		maxArgs: 1,
		run: func(ctx context.Context, args []string) error {
			project, api, err := a.projectAndAPI(ctx)
			if err != nil {
				return err
			}
			value, err := api.Variable(ctx, project, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(a.env.Stdout, value)
			return nil
		},
	}
}

func (a *app) varSetCommand() *command {
	return &command{
		name:    "set",
		args:    "<key> [value]",
		summary: "Set a variable; without value it is read from stdin, keeping it out of shell history",
		minArgs: 1,
		maxArgs: 2,
		run: func(ctx context.Context, args []string) error {
			value, err := a.valueArg(args)
			if err != nil {
				return err
			}
			project, api, err := a.projectAndAPI(ctx)
			if err != nil {
				return err
			}
			version, err := api.SetVariable(ctx, project, args[0], value)
			if err != nil {
				return err
			}
			fmt.Fprintf(a.env.Stdout, "Set %s in %s (version %d)\n", args[0], project, version)
			return nil
		},
	}
}

func (a *app) varUnsetCommand() *command {
	return &command{
		name:    "unset",
		args:    "<key>",
		summary: "Delete a variable",
		minArgs: 1,
		maxArgs: 1,
		run: func(ctx context.Context, args []string) error {
			project, api, err := a.projectAndAPI(ctx)
			if err != nil {
				return err
			}
			version, err := api.DeleteVariable(ctx, project, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(a.env.Stdout, "Deleted %s from %s (version %d)\n", args[0], project, version)
			return nil
		},
	}
}

func (a *app) valueArg(args []string) (string, error) {
	if len(args) == 2 {
		return args[1], nil
	}
	raw, err := io.ReadAll(io.LimitReader(a.env.Stdin, maxStdinValue+1))
	if err != nil {
		return "", fmt.Errorf("read value from stdin: %w", err)
	}
	if len(raw) > maxStdinValue {
		return "", errors.New("value read from stdin is too large")
	}
	return strings.TrimSuffix(strings.TrimSuffix(string(raw), "\n"), "\r"), nil
}

func (a *app) projectAndAPI(ctx context.Context) (string, API, error) {
	project, err := a.projectName(ctx, "")
	if err != nil {
		return "", nil, err
	}
	api, err := a.api()
	if err != nil {
		return "", nil, err
	}
	return project, api, nil
}
