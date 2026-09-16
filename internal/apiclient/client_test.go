package apiclient_test

import (
	"context"
	"encoding/pem"
	"errors"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/PoCInnovation/SecureEnv/internal/apiclient"
	"github.com/PoCInnovation/SecureEnv/internal/domain"
	"github.com/PoCInnovation/SecureEnv/internal/httpapi"
	"github.com/PoCInnovation/SecureEnv/internal/project"
	"github.com/PoCInnovation/SecureEnv/internal/project/projecttest"
)

type healthy struct{}

func (healthy) Ping(context.Context) error { return nil }

// newStack starts the real HTTP API on an in-memory store and returns a
// client talking to it with a valid token, the store and the server URL.
func newStack(t *testing.T) (*apiclient.Client, *projecttest.MemStore, string) {
	t.Helper()

	store := projecttest.NewMemStore()
	svc := project.NewService(store)
	handler := httpapi.NewHandler(
		func(token string) (httpapi.ProjectService, error) {
			if token != "valid" {
				return nil, domain.ErrForbidden
			}
			return svc, nil
		},
		healthy{},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := apiclient.New(server.URL+"/", "valid")
	if err != nil {
		t.Fatal(err)
	}
	return client, store, server.URL
}

func TestProjectLifecycle(t *testing.T) {
	t.Parallel()
	client, store, _ := newStack(t)
	ctx := t.Context()

	if err := client.CreateProject(ctx, "app"); err != nil {
		t.Fatalf("CreateProject() error: %v", err)
	}
	if err := client.CreateProject(ctx, "app"); !errors.Is(err, domain.ErrProjectExists) {
		t.Fatalf("duplicate CreateProject() error = %v, want ErrProjectExists", err)
	}

	names, err := client.ListProjects(ctx)
	if err != nil || !slices.Equal(names, []string{"app"}) {
		t.Fatalf("ListProjects() = %v, %v", names, err)
	}

	info, err := client.Project(ctx, "app")
	if err != nil || info.Name != "app" || info.CurrentVersion != 1 {
		t.Fatalf("Project() = %+v, %v", info, err)
	}

	if err := client.RenameProject(ctx, "app", "renamed"); err != nil {
		t.Fatalf("RenameProject() error: %v", err)
	}
	if store.Has("app") || !store.Has("renamed") {
		t.Fatal("rename not applied")
	}

	if err := client.DeleteProject(ctx, "renamed"); err != nil {
		t.Fatalf("DeleteProject() error: %v", err)
	}
	if _, err := client.Project(ctx, "renamed"); !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("Project() after delete error = %v, want ErrProjectNotFound", err)
	}
}

func TestVariables(t *testing.T) {
	t.Parallel()
	client, store, _ := newStack(t)
	ctx := t.Context()
	store.Seed("app", map[string]string{"A": "1"})

	version, err := client.SetVariable(ctx, "app", "B", "2")
	if err != nil || version != 2 {
		t.Fatalf("SetVariable() = %d, %v", version, err)
	}

	value, err := client.Variable(ctx, "app", "B")
	if err != nil || value != "2" {
		t.Fatalf("Variable() = %q, %v", value, err)
	}

	snapshot, err := client.Variables(ctx, "app")
	if err != nil || snapshot.Version != 2 || !maps.Equal(snapshot.Variables.Map(), map[string]string{"A": "1", "B": "2"}) {
		t.Fatalf("Variables() = %+v, %v", snapshot, err)
	}

	if _, err := client.ReplaceVariables(ctx, "app", map[string]string{"C": "3"}, 1); !errors.Is(err, domain.ErrVersionConflict) {
		t.Fatalf("stale ReplaceVariables() error = %v, want ErrVersionConflict", err)
	}
	version, err = client.ReplaceVariables(ctx, "app", map[string]string{"C": "3"}, snapshot.Version)
	if err != nil || version != 3 {
		t.Fatalf("ReplaceVariables() = %d, %v", version, err)
	}
	if _, err := client.ReplaceVariables(ctx, "app", map[string]string{"D": "4"}, domain.AnyVersion); err != nil {
		t.Fatalf("ReplaceVariables(AnyVersion) error: %v", err)
	}

	if _, err := client.DeleteVariable(ctx, "app", "D"); err != nil {
		t.Fatalf("DeleteVariable() error: %v", err)
	}
	if _, err := client.Variable(ctx, "app", "D"); !errors.Is(err, domain.ErrVariableNotFound) {
		t.Fatalf("Variable() after delete error = %v, want ErrVariableNotFound", err)
	}
}

func TestInvalidInputFailsBeforeSending(t *testing.T) {
	t.Parallel()

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests++ }))
	t.Cleanup(server.Close)
	client, err := apiclient.New(server.URL, "token")
	if err != nil {
		t.Fatal(err)
	}
	ctx := t.Context()

	checks := map[string]error{}
	checks["delete .."] = client.DeleteProject(ctx, "..")
	checks["create a/b"] = client.CreateProject(ctx, "a/b")
	checks["rename to bad"] = client.RenameProject(ctx, "app", "bad name")
	_, checks["get reserved key"] = client.Variable(ctx, "app", "SECURE_ENV_TOKEN")
	_, checks["set bad key"] = client.SetVariable(ctx, "app", "a-b", "x")

	for name, err := range checks {
		if err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
	if requests != 0 {
		t.Fatalf("%d requests were sent", requests)
	}
}

func TestErrors(t *testing.T) {
	t.Parallel()

	t.Run("forbidden token", func(t *testing.T) {
		t.Parallel()
		_, _, url := newStack(t)
		bad, err := apiclient.New(url, "invalid")
		if err != nil {
			t.Fatal(err)
		}
		_, err = bad.ListProjects(t.Context())
		var apiErr *apiclient.Error
		if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden || !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("non JSON error body", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "bad gateway", http.StatusBadGateway)
		}))
		t.Cleanup(server.Close)
		client, _ := apiclient.New(server.URL, "token")

		_, err := client.ListProjects(t.Context())
		var apiErr *apiclient.Error
		if !errors.As(err, &apiErr) || apiErr.Status != http.StatusBadGateway || apiErr.Message != "bad gateway" {
			t.Fatalf("error = %#v", err)
		}
	})

	t.Run("unreachable server", func(t *testing.T) {
		t.Parallel()
		client, _ := apiclient.New("http://127.0.0.1:1", "token")
		if _, err := client.ListProjects(t.Context()); err == nil {
			t.Fatal("expected a connection error")
		}
	})
}

func TestNewValidation(t *testing.T) {
	t.Parallel()

	for _, url := range []string{"", "localhost:8080", "ftp://host", "http://"} {
		if _, err := apiclient.New(url, "token"); err == nil {
			t.Errorf("New(%q) should fail", url)
		}
	}
	if _, err := apiclient.New("http://localhost:8080", ""); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("New without token error = %v, want ErrUnauthorized", err)
	}
}

func TestRequestHeaders(t *testing.T) {
	t.Parallel()

	var got http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		_, _ = w.Write([]byte(`{"projects":[]}`))
	}))
	t.Cleanup(server.Close)

	client, _ := apiclient.New(server.URL, "s3cr3t", apiclient.WithUserAgent("secureenv/test"))
	if _, err := client.ListProjects(t.Context()); err != nil {
		t.Fatal(err)
	}
	if got.Get("Authorization") != "Bearer s3cr3t" || got.Get("User-Agent") != "secureenv/test" {
		t.Fatalf("headers = %v", got)
	}
}

func TestCustomCACertificate(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"projects":["tls"]}`))
	}))
	t.Cleanup(server.Close)

	caFile := filepath.Join(t.TempDir(), "ca.pem")
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	if err := os.WriteFile(caFile, certPEM, 0o600); err != nil {
		t.Fatal(err)
	}

	untrusted, _ := apiclient.New(server.URL, "token")
	if _, err := untrusted.ListProjects(t.Context()); err == nil {
		t.Fatal("a server signed by an unknown CA must be rejected")
	}

	client, err := apiclient.New(server.URL, "token", apiclient.WithCACertFile(caFile))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	names, err := client.ListProjects(t.Context())
	if err != nil || len(names) != 1 || names[0] != "tls" {
		t.Fatalf("ListProjects() = %v, %v", names, err)
	}
}

func TestInvalidCACertificateFile(t *testing.T) {
	t.Parallel()

	missing := filepath.Join(t.TempDir(), "missing.pem")
	if _, err := apiclient.New("https://localhost", "token", apiclient.WithCACertFile(missing)); err == nil {
		t.Error("missing CA file should fail")
	}

	garbage := filepath.Join(t.TempDir(), "garbage.pem")
	if err := os.WriteFile(garbage, []byte("not a certificate"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := apiclient.New("https://localhost", "token", apiclient.WithCACertFile(garbage)); err == nil {
		t.Error("file without certificates should fail")
	}
}
