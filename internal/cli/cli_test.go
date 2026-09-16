package cli_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"maps"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PoCInnovation/SecureEnv/internal/apiclient"
	"github.com/PoCInnovation/SecureEnv/internal/cli"
	"github.com/PoCInnovation/SecureEnv/internal/domain"
	"github.com/PoCInnovation/SecureEnv/internal/dotenv"
	"github.com/PoCInnovation/SecureEnv/internal/httpapi"
	"github.com/PoCInnovation/SecureEnv/internal/project"
	"github.com/PoCInnovation/SecureEnv/internal/project/projecttest"
)

const validToken = "hvs.test"

type healthy struct{}

func (healthy) Ping(context.Context) error { return nil }

// harness runs the CLI against the real HTTP API backed by an in-memory store.
type harness struct {
	t       *testing.T
	store   *projecttest.MemStore
	dir     string
	env     map[string]string
	origin  string
	apiURL  string
	gotURLs []string
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	store := projecttest.NewMemStore()
	svc := project.NewService(store)
	server := httptest.NewServer(httpapi.NewHandler(
		func(token string) (httpapi.ProjectService, error) {
			if token != validToken {
				return nil, domain.ErrForbidden
			}
			return svc, nil
		},
		healthy{},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	))
	t.Cleanup(server.Close)

	return &harness{
		t:      t,
		store:  store,
		dir:    t.TempDir(),
		env:    map[string]string{cli.KeyToken: validToken, cli.KeyAPIURL: server.URL},
		origin: "git@github.com:PoCInnovation/SecureEnv.git",
		apiURL: server.URL,
	}
}

type result struct {
	code   int
	stdout string
	stderr string
}

func (h *harness) run(stdin string, args ...string) result {
	h.t.Helper()
	var stdout, stderr bytes.Buffer
	code := cli.Run(h.t.Context(), args, cli.Env{
		Stdin:   strings.NewReader(stdin),
		Stdout:  &stdout,
		Stderr:  &stderr,
		Getenv:  func(key string) string { return h.env[key] },
		Dir:     h.dir,
		Version: "test",
		NewAPI: func(baseURL, token string) (cli.API, error) {
			h.gotURLs = append(h.gotURLs, baseURL)
			return apiclient.New(baseURL, token)
		},
		OriginURL: func(context.Context, string) (string, error) {
			if h.origin == "" {
				return "", errors.New("no origin")
			}
			return h.origin, nil
		},
	})
	return result{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

func (h *harness) mustRun(args ...string) result {
	h.t.Helper()
	r := h.run("", args...)
	if r.code != cli.ExitOK {
		h.t.Fatalf("secureenv %v exited %d\nstdout: %s\nstderr: %s", args, r.code, r.stdout, r.stderr)
	}
	return r
}

func (h *harness) writeEnv(content string) {
	h.t.Helper()
	if err := os.WriteFile(filepath.Join(h.dir, ".env"), []byte(content), 0o600); err != nil {
		h.t.Fatal(err)
	}
}

func (h *harness) readEnv() map[string]string {
	h.t.Helper()
	file, err := dotenv.Load(filepath.Join(h.dir, ".env"))
	if err != nil {
		h.t.Fatal(err)
	}
	return file.Values()
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("output does not contain %q:\n%s", want, got)
	}
}

func TestWorkflow(t *testing.T) {
	t.Parallel()
	h := newHarness(t)

	// Link the directory to the project derived from git and create it.
	h.mustRun("init")
	if got := h.readEnv()[cli.KeyProject]; got != "PoCInnovation_SecureEnv" {
		t.Fatalf("%s = %q", cli.KeyProject, got)
	}
	h.mustRun("project", "create")

	// Push local variables.
	h.writeEnv("SECURE_ENV_PROJECT=PoCInnovation_SecureEnv\nDATABASE_URL=postgres://db\nAPI_KEY=abc\n")
	assertContains(t, h.mustRun("status").stdout, "+ API_KEY")
	assertContains(t, h.mustRun("push").stdout, "Pushed PoCInnovation_SecureEnv (version 2)")
	assertContains(t, h.mustRun("status").stdout, "Up to date.")
	assertContains(t, h.mustRun("push").stdout, "Nothing to push")

	// Someone else changes the remote.
	h.store.Seed("PoCInnovation_SecureEnv", map[string]string{"DATABASE_URL": "postgres://db", "API_KEY": "rotated", "NEW": "1"})
	status := h.mustRun("status").stdout
	assertContains(t, status, "~ API_KEY")
	assertContains(t, status, "- NEW")

	// Pull refuses to drop the local API_KEY value, force applies it.
	r := h.run("", "pull")
	if r.code != cli.ExitError {
		t.Fatalf("pull with local changes exited %d", r.code)
	}
	assertContains(t, r.stderr, "API_KEY")
	h.mustRun("pull", "-force")
	want := map[string]string{
		"SECURE_ENV_PROJECT": "PoCInnovation_SecureEnv",
		"DATABASE_URL":       "postgres://db",
		"API_KEY":            "rotated",
		"NEW":                "1",
	}
	if got := h.readEnv(); !maps.Equal(got, want) {
		t.Fatalf(".env = %v, want %v", got, want)
	}

	// Push refuses to delete remote variables unless forced.
	h.writeEnv("SECURE_ENV_PROJECT=PoCInnovation_SecureEnv\nDATABASE_URL=postgres://db\n")
	if r := h.run("", "push"); r.code != cli.ExitError {
		t.Fatalf("push deleting remote variables exited %d", r.code)
	}
	h.mustRun("push", "-force")
	if got := h.store.Latest("PoCInnovation_SecureEnv"); !maps.Equal(got, map[string]string{"DATABASE_URL": "postgres://db"}) {
		t.Fatalf("remote = %v", got)
	}
}

func TestClone(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.store.Seed("backend", map[string]string{"A": "1"})
	h.writeEnv("SECURE_ENV_TOKEN=kept\n")

	h.mustRun("clone", "backend")

	want := map[string]string{"SECURE_ENV_TOKEN": "kept", "SECURE_ENV_PROJECT": "backend", "A": "1"}
	if got := h.readEnv(); !maps.Equal(got, want) {
		t.Fatalf(".env = %v, want %v", got, want)
	}
}

func TestProjectCommands(t *testing.T) {
	t.Parallel()
	h := newHarness(t)

	h.mustRun("project", "create", "b")
	h.mustRun("project", "create", "a")
	if got := h.mustRun("project", "list").stdout; got != "a\nb\n" {
		t.Fatalf("list = %q", got)
	}
	assertContains(t, h.mustRun("project", "info", "a").stdout, "Version:  1")

	h.mustRun("project", "rename", "a", "c")
	if got := h.mustRun("project", "list").stdout; got != "b\nc\n" {
		t.Fatalf("list after rename = %q", got)
	}

	if r := h.run("wrong\n", "project", "delete", "b"); r.code != cli.ExitError || !h.store.Has("b") {
		t.Fatalf("delete with wrong confirmation: code %d, exists %v", r.code, h.store.Has("b"))
	}
	if r := h.run("b\n", "project", "delete", "b"); r.code != cli.ExitOK || h.store.Has("b") {
		t.Fatalf("confirmed delete: code %d, stderr %s", r.code, r.stderr)
	}
	h.mustRun("project", "delete", "-yes", "c")
	if h.store.Has("c") {
		t.Fatal("project c still exists")
	}
}

func TestVarCommands(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.store.Seed("app", map[string]string{})
	h.env[cli.KeyProject] = "app"

	h.mustRun("var", "set", "PLAIN", "value")
	if r := h.run("from stdin\n", "var", "set", "SECRET"); r.code != cli.ExitOK {
		t.Fatalf("set from stdin: %s", r.stderr)
	}
	if got := h.mustRun("var", "get", "SECRET").stdout; got != "from stdin\n" {
		t.Fatalf("get = %q", got)
	}
	if got := h.mustRun("var", "list").stdout; got != "PLAIN\nSECRET\n" {
		t.Fatalf("list = %q", got)
	}
	assertContains(t, h.mustRun("var", "list", "-values").stdout, `SECRET="from stdin"`)

	h.mustRun("var", "unset", "PLAIN")
	r := h.run("", "var", "get", "PLAIN")
	if r.code != cli.ExitError {
		t.Fatalf("get deleted variable exited %d", r.code)
	}
	assertContains(t, r.stderr, "variable not found")
}

func TestSettingsPrecedence(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.store.Seed("from-flag", map[string]string{})
	h.store.Seed("from-env", map[string]string{})
	h.store.Seed("from-file", map[string]string{"F": "1"})

	h.writeEnv("SECURE_ENV_PROJECT_NAME=from-file\n")
	assertContains(t, h.mustRun("status").stdout, "Project from-file")

	h.env[cli.KeyProject] = "from-env"
	assertContains(t, h.mustRun("status").stdout, "Project from-env")

	assertContains(t, h.mustRun("-project", "from-flag", "status").stdout, "Project from-flag")

	h.mustRun("-api", h.apiURL+"/", "project", "list")
	if last := h.gotURLs[len(h.gotURLs)-1]; last != h.apiURL+"/" {
		t.Fatalf("-api flag ignored, used %q", last)
	}
}

func TestErrorsAndHints(t *testing.T) {
	t.Parallel()

	t.Run("bad token", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t)
		h.env[cli.KeyToken] = "wrong"
		r := h.run("", "project", "list")
		if r.code != cli.ExitError {
			t.Fatalf("code = %d", r.code)
		}
		assertContains(t, r.stderr, "hint: your Vault token policies")
	})

	t.Run("missing token", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t)
		delete(h.env, cli.KeyToken)
		assertContains(t, h.run("", "project", "list").stderr, "hint: set SECURE_ENV_TOKEN")
	})

	t.Run("no project", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t)
		h.origin = ""
		assertContains(t, h.run("", "status").stderr, "hint: run `secureenv init <project>`")
	})

	t.Run("usage", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t)
		for _, args := range [][]string{{}, {"nope"}, {"var", "get"}, {"push", "-unknown"}, {"project"}} {
			r := h.run("", args...)
			if r.code != cli.ExitUsage {
				t.Errorf("%v exited %d, want %d", args, r.code, cli.ExitUsage)
			}
			assertContains(t, r.stderr, "Usage: secureenv")
		}
	})

	t.Run("help", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t)
		for _, args := range [][]string{{"-h"}, {"help"}, {"var", "help"}, {"push", "-h"}} {
			r := h.run("", args...)
			if r.code != cli.ExitOK {
				t.Errorf("%v exited %d", args, r.code)
			}
			assertContains(t, r.stderr, "Usage:")
		}
	})
}

func TestVersion(t *testing.T) {
	t.Parallel()
	if got := newHarness(t).mustRun("version").stdout; got != "secureenv test\n" {
		t.Fatalf("version = %q", got)
	}
}
