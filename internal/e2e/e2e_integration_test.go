//go:build integration

// Package e2e runs the CLI against the API server backed by a real Vault,
// with the least-privilege policies shipped in deploy/vault/policies.
package e2e

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"io"
	"log/slog"
	"maps"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	vault "github.com/hashicorp/vault/api"

	"github.com/PoCInnovation/SecureEnv/internal/apiclient"
	"github.com/PoCInnovation/SecureEnv/internal/apiserver"
	"github.com/PoCInnovation/SecureEnv/internal/cli"
	"github.com/PoCInnovation/SecureEnv/internal/dotenv"
)

// vaultPrefix must match the paths granted by deploy/vault/policies.
const vaultPrefix = "secureenv"

func TestEndToEnd(t *testing.T) {
	if os.Getenv("VAULT_ADDR") == "" || os.Getenv("VAULT_TOKEN") == "" {
		t.Skip("VAULT_ADDR and VAULT_TOKEN are required")
	}
	ctx := t.Context()

	admin, err := vault.NewClient(vault.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	for _, policy := range []string{"secureenv-writer", "secureenv-reader"} {
		rules, err := os.ReadFile(filepath.Join("..", "..", "deploy", "vault", "policies", policy+".hcl"))
		if err != nil {
			t.Fatal(err)
		}
		if err := admin.Sys().PutPolicyWithContext(ctx, policy, string(rules)); err != nil {
			t.Fatalf("write policy %s: %v", policy, err)
		}
	}
	writer := createToken(t, admin, "secureenv-writer")
	reader := createToken(t, admin, "secureenv-reader")

	apiURL := startAPI(t)
	project := "e2e-" + strings.ToLower(rand.Text())
	renamed := project + "-renamed"
	t.Cleanup(func() {
		for _, name := range []string{project, renamed, project + "-ro"} {
			_ = admin.KVv2("secret").DeleteMetadata(context.Background(), vaultPrefix+"/"+name)
		}
	})

	dir := t.TempDir()
	h := &harness{t: t, apiURL: apiURL, dir: dir, token: writer}

	h.ok("project", "create", project)
	h.writeEnv(dir, "SECURE_ENV_PROJECT="+project+"\n"+
		`DATABASE_URL="postgres://app:p@ss=w0rd@db:5432/app?sslmode=require"`+"\n"+
		`CERT="line1\nline2"`+"\n"+
		"EMPTY=\n")

	// Push, then check what Vault actually stores.
	contains(t, h.ok("push").stdout, "Pushed "+project+" (version 2)")
	want := map[string]string{
		"DATABASE_URL": "postgres://app:p@ss=w0rd@db:5432/app?sslmode=require",
		"CERT":         "line1\nline2",
		"EMPTY":        "",
	}
	if got := vaultData(t, admin, project); !maps.Equal(got, want) {
		t.Fatalf("Vault data = %q, want %q (SECURE_ENV_* must never be stored)", got, want)
	}
	contains(t, h.ok("status").stdout, "Up to date.")

	// A secret set from stdin reaches Vault and a clone gets everything.
	h.stdin = "s3cr3t value\n"
	h.ok("-project", project, "var", "set", "API_KEY")
	h.stdin = ""
	if got := h.ok("-project", project, "var", "get", "API_KEY").stdout; got != "s3cr3t value\n" {
		t.Fatalf("var get = %q", got)
	}
	clone := &harness{t: t, apiURL: apiURL, dir: t.TempDir(), token: writer}
	clone.ok("clone", project)
	want["API_KEY"] = "s3cr3t value"
	cloned := clone.readEnv()
	delete(cloned, "SECURE_ENV_PROJECT")
	if !maps.Equal(cloned, want) {
		t.Fatalf("cloned .env = %q, want %q", cloned, want)
	}

	// The first directory misses API_KEY: push must refuse to delete it.
	if r := h.run("push"); r.code != cli.ExitError || !strings.Contains(r.stderr, "API_KEY") {
		t.Fatalf("push deleting a remote variable: code %d, stderr %s", r.code, r.stderr)
	}
	h.ok("pull")

	// The reader policy can read but not write.
	ro := &harness{t: t, apiURL: apiURL, dir: t.TempDir(), token: reader}
	if got := ro.ok("-project", project, "var", "get", "API_KEY").stdout; got != "s3cr3t value\n" {
		t.Fatalf("reader var get = %q", got)
	}
	contains(t, ro.ok("project", "list").stdout, project)
	for _, args := range [][]string{
		{"-project", project, "var", "set", "X", "1"},
		{"project", "create", project + "-ro"},
		{"project", "delete", "-yes", project},
	} {
		r := ro.run(args...)
		if r.code != cli.ExitError || !strings.Contains(r.stderr, "policies do not allow") {
			t.Errorf("reader %v: code %d, stderr %s", args, r.code, r.stderr)
		}
	}

	// An unknown token is rejected by Vault.
	bad := &harness{t: t, apiURL: apiURL, dir: t.TempDir(), token: "hvs.invalid"}
	if r := bad.run("project", "list"); r.code != cli.ExitError {
		t.Errorf("invalid token: code %d, stderr %s", r.code, r.stderr)
	}

	// Rename keeps the history, delete removes everything.
	h.ok("project", "rename", project, renamed)
	contains(t, h.ok("project", "info", renamed).stdout, "Version:  3")
	if _, err := admin.KVv2("secret").GetMetadata(ctx, vaultPrefix+"/"+project); !errors.Is(err, vault.ErrSecretNotFound) {
		t.Fatalf("old project still in Vault: %v", err)
	}
	h.ok("project", "delete", "-yes", renamed)
	if _, err := admin.KVv2("secret").GetMetadata(ctx, vaultPrefix+"/"+renamed); !errors.Is(err, vault.ErrSecretNotFound) {
		t.Fatalf("deleted project still in Vault: %v", err)
	}
}

func createToken(t *testing.T, admin *vault.Client, policy string) string {
	t.Helper()
	secret, err := admin.Auth().Token().CreateWithContext(t.Context(), &vault.TokenCreateRequest{
		Policies:    []string{policy},
		TTL:         "15m",
		DisplayName: "e2e-" + policy,
	})
	if err != nil {
		t.Fatalf("create %s token: %v", policy, err)
	}
	t.Cleanup(func() { _ = admin.Auth().Token().RevokeTreeWithContext(context.Background(), secret.Auth.ClientToken) })
	return secret.Auth.ClientToken
}

// startAPI runs the real API server in-process and waits until it reports
// Vault as ready.
func startAPI(t *testing.T) string {
	t.Helper()
	listener, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	cfg := apiserver.Config{VaultMount: "secret", VaultPrefix: vaultPrefix}
	go func() { done <- apiserver.Run(ctx, cfg, listener, slog.New(slog.NewTextHandler(io.Discard, nil))) }()
	t.Cleanup(func() {
		cancel()
		if err := <-done; err != nil {
			t.Errorf("API server: %v", err)
		}
	})

	url := "http://" + listener.Addr().String()
	deadline := time.Now().Add(10 * time.Second)
	for {
		resp, err := http.Get(url + "/readyz")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return url
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("API not ready: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func vaultData(t *testing.T, admin *vault.Client, project string) map[string]string {
	t.Helper()
	secret, err := admin.KVv2("secret").Get(t.Context(), vaultPrefix+"/"+project)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for key, value := range secret.Data {
		out[key], _ = value.(string)
	}
	return out
}

type harness struct {
	t      *testing.T
	apiURL string
	dir    string
	token  string
	stdin  string
}

type result struct {
	code           int
	stdout, stderr string
}

func (h *harness) run(args ...string) result {
	h.t.Helper()
	var stdout, stderr bytes.Buffer
	env := map[string]string{cli.KeyAPIURL: h.apiURL, cli.KeyToken: h.token}
	code := cli.Run(h.t.Context(), args, cli.Env{
		Stdin:   strings.NewReader(h.stdin),
		Stdout:  &stdout,
		Stderr:  &stderr,
		Getenv:  func(key string) string { return env[key] },
		Dir:     h.dir,
		Version: "e2e",
		NewAPI: func(cfg cli.APIConfig) (cli.API, error) {
			return apiclient.New(cfg.URL, cfg.Token)
		},
		OriginURL: func(context.Context, string) (string, error) {
			return "", errors.New("not a git repository")
		},
	})
	return result{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

func (h *harness) ok(args ...string) result {
	h.t.Helper()
	r := h.run(args...)
	if r.code != cli.ExitOK {
		h.t.Fatalf("secureenv %v exited %d\nstdout: %s\nstderr: %s", args, r.code, r.stdout, r.stderr)
	}
	return r
}

func (h *harness) writeEnv(dir, content string) {
	h.t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(content), 0o600); err != nil {
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

func contains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("output does not contain %q:\n%s", want, got)
	}
}
