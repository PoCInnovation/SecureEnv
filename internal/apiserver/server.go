package apiserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	vault "github.com/hashicorp/vault/api"

	"github.com/PoCInnovation/SecureEnv/internal/httpapi"
	"github.com/PoCInnovation/SecureEnv/internal/project"
	"github.com/PoCInnovation/SecureEnv/internal/vaultstore"
)

const shutdownTimeout = 15 * time.Second

// NewHandler builds the API handler on top of a Vault client. The client's own
// token is cleared: every request must bring its caller's token.
func NewHandler(client *vault.Client, cfg Config, logger *slog.Logger) http.Handler {
	client.ClearToken()
	base := vaultstore.New(client, vaultstore.Options{Mount: cfg.VaultMount, Prefix: cfg.VaultPrefix})

	services := func(token string) (httpapi.ProjectService, error) {
		store, err := base.WithToken(token)
		if err != nil {
			return nil, err
		}
		return project.NewService(store), nil
	}
	return httpapi.NewHandler(services, base, logger)
}

// Run serves the API on listener until ctx is cancelled, then drains
// in-flight requests.
func Run(ctx context.Context, cfg Config, listener net.Listener, logger *slog.Logger) error {
	client, err := vault.NewClient(vault.DefaultConfig())
	if err != nil {
		return fmt.Errorf("create vault client: %w", err)
	}

	server := &http.Server{
		Handler:           NewHandler(client, cfg, logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       2 * time.Minute,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelWarn),
	}

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("api listening", slog.String("addr", listener.Addr().String()), slog.Bool("tls", cfg.TLSEnabled()), slog.String("vault_addr", client.Address()))
		if cfg.TLSEnabled() {
			serveErr <- server.ServeTLS(listener, cfg.TLSCertFile, cfg.TLSKeyFile)
		} else {
			serveErr <- server.Serve(listener)
		}
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}

	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-serveErr; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
