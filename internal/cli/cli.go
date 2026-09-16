// Package cli implements the secureenv command line interface.
package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/PoCInnovation/SecureEnv/internal/apiv1"
	"github.com/PoCInnovation/SecureEnv/internal/domain"
	"github.com/PoCInnovation/SecureEnv/internal/envsync"
)

// Exit codes returned by Run.
const (
	ExitOK    = 0
	ExitError = 1
	ExitUsage = 2
)

// API is the remote SecureEnv API used by the CLI.
type API interface {
	envsync.Remote
	ListProjects(ctx context.Context) ([]string, error)
	CreateProject(ctx context.Context, name string) error
	Project(ctx context.Context, name string) (apiv1.Project, error)
	RenameProject(ctx context.Context, from, to string) error
	DeleteProject(ctx context.Context, name string) error
	Variable(ctx context.Context, project, key string) (string, error)
	SetVariable(ctx context.Context, project, key, value string) (domain.Version, error)
	DeleteVariable(ctx context.Context, project, key string) (domain.Version, error)
}

// Env holds everything the CLI needs from the outside world.
type Env struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	// Getenv reads process environment variables.
	Getenv func(string) string
	// Dir is the working directory; relative paths are resolved from it.
	Dir string
	// Version is printed by the version command.
	Version string
	// NewAPI builds an API client.
	NewAPI func(baseURL, token string) (API, error)
	// OriginURL returns the git origin URL of the repository holding dir.
	OriginURL func(ctx context.Context, dir string) (string, error)
}

// Run executes the command line args and returns the process exit code.
func Run(ctx context.Context, args []string, env Env) int {
	a := &app{env: env}
	root := a.commands()

	err := root.execute(ctx, args, env.Stderr)
	var usage *usageError
	switch {
	case err == nil:
		return ExitOK
	case errors.Is(err, flag.ErrHelp):
		return ExitOK
	case errors.As(err, &usage):
		fmt.Fprintf(env.Stderr, "secureenv: %s\n\n", usage.msg)
		usage.cmd.printUsage(env.Stderr)
		return ExitUsage
	default:
		fmt.Fprintf(env.Stderr, "secureenv: %s\n", err)
		if hint := hintFor(err); hint != "" {
			fmt.Fprintf(env.Stderr, "hint: %s\n", hint)
		}
		return ExitError
	}
}

func hintFor(err error) string {
	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		return "set SECURE_ENV_TOKEN (or VAULT_TOKEN) to a valid Vault token"
	case errors.Is(err, domain.ErrForbidden):
		return "your Vault token policies do not allow this operation"
	case errors.Is(err, domain.ErrVersionConflict):
		return "the project changed remotely; run `secureenv status` then pull"
	case errors.Is(err, errNoProject):
		return "run `secureenv init <project>` or pass -project"
	}
	return ""
}
