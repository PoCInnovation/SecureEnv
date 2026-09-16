package cli

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/PoCInnovation/SecureEnv/internal/dotenv"
	"github.com/PoCInnovation/SecureEnv/internal/gitremote"
)

// Settings read from the environment or the local .env file.
const (
	KeyAPIURL        = "SECURE_ENV_API_URL"
	KeyToken         = "SECURE_ENV_TOKEN" //nolint:gosec // environment variable name, not a credential
	KeyProject       = "SECURE_ENV_PROJECT"
	keyLegacyProject = "SECURE_ENV_PROJECT_NAME"
	keyVaultToken    = "VAULT_TOKEN"

	defaultAPIURL = "http://127.0.0.1:8080"
)

var errNoProject = errors.New("no project selected")

// app resolves settings and dependencies for commands.
type app struct {
	env Env

	// Global flags.
	apiURL  string
	project string
	file    string
}

func (a *app) envPath() string {
	if filepath.IsAbs(a.file) {
		return a.file
	}
	return filepath.Join(a.env.Dir, a.file)
}

// setting returns the first non empty value among the flag, the process
// environment and the .env file.
func (a *app) setting(flagValue string, keys ...string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	for _, key := range keys {
		if value := a.env.Getenv(key); value != "" {
			return value, nil
		}
	}
	file, err := dotenv.Load(a.envPath())
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	values := file.Values()
	for _, key := range keys {
		if value := values[key]; value != "" {
			return value, nil
		}
	}
	return "", nil
}

func (a *app) api() (API, error) {
	url, err := a.setting(a.apiURL, KeyAPIURL)
	if err != nil {
		return nil, err
	}
	if url == "" {
		url = defaultAPIURL
	}
	token, err := a.setting("", KeyToken, keyVaultToken)
	if err != nil {
		return nil, err
	}
	return a.env.NewAPI(url, token)
}

// projectName resolves the project: explicit argument, -project flag,
// environment, .env file, then the git origin remote.
func (a *app) projectName(ctx context.Context, explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	name, err := a.setting(a.project, KeyProject, keyLegacyProject)
	if err != nil || name != "" {
		return name, err
	}
	return a.projectFromGit(ctx)
}

func (a *app) projectFromGit(ctx context.Context) (string, error) {
	origin, err := a.env.OriginURL(ctx, a.env.Dir)
	if err != nil {
		return "", fmt.Errorf("%w: %w", errNoProject, err)
	}
	return gitremote.ProjectName(origin)
}
