package apiserver_test

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/PoCInnovation/SecureEnv/internal/apiserver"
)

func TestRunServesAndShutsDownGracefully(t *testing.T) {
	t.Setenv("VAULT_ADDR", "http://127.0.0.1:1")
	t.Setenv("VAULT_TOKEN", "must-be-ignored")

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	done := make(chan error, 1)
	go func() { done <- apiserver.Run(ctx, apiserver.Config{VaultMount: "secret"}, listener, logger) }()

	base := "http://" + listener.Addr().String()
	resp := waitFor(t, base+"/healthz")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/healthz = %d", resp.StatusCode)
	}

	// The server's own VAULT_TOKEN must never authenticate callers.
	resp, err = http.Get(base + "/v1/projects")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("/v1/projects without token = %d, want 401", resp.StatusCode)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down")
	}
}

func waitFor(t *testing.T, url string) *http.Response {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			return resp
		}
		if time.Now().After(deadline) {
			t.Fatalf("server not reachable: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
