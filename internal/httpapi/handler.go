// Package httpapi exposes the project use cases over HTTP (API v1).
//
// Every /v1 request must carry "Authorization: Bearer <vault token>". The
// token is forwarded to Vault, so what a caller may do is decided by Vault
// policies rather than by the API.
package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/PoCInnovation/SecureEnv/internal/domain"
)

// ProjectService is the subset of project.Service the API needs.
type ProjectService interface {
	List(ctx context.Context) ([]domain.ProjectName, error)
	Create(ctx context.Context, name string) error
	Info(ctx context.Context, name string) (domain.ProjectInfo, error)
	Rename(ctx context.Context, from, to string) error
	Delete(ctx context.Context, name string) error
	Variables(ctx context.Context, name string) (domain.Snapshot, error)
	Variable(ctx context.Context, name, key string) (string, error)
	SetVariable(ctx context.Context, name, key, value string) (domain.Version, error)
	DeleteVariable(ctx context.Context, name, key string) (domain.Version, error)
	ReplaceVariables(ctx context.Context, name string, vars map[string]string, expected domain.Version) (domain.Version, error)
}

// ServiceFactory returns a service acting with the caller's token.
type ServiceFactory func(token string) (ProjectService, error)

// HealthChecker reports whether the backing store is usable.
type HealthChecker interface {
	Ping(ctx context.Context) error
}

type server struct {
	services ServiceFactory
	health   HealthChecker
	logger   *slog.Logger
}

// NewHandler returns the root HTTP handler of the API.
func NewHandler(services ServiceFactory, health HealthChecker, logger *slog.Logger) http.Handler {
	s := &server{services: services, health: health, logger: logger}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.liveness)
	mux.HandleFunc("GET /readyz", s.readiness)

	mux.Handle("GET /v1/projects", s.authenticated(s.listProjects))
	mux.Handle("POST /v1/projects", s.authenticated(s.createProject))
	mux.Handle("GET /v1/projects/{project}", s.authenticated(s.projectInfo))
	mux.Handle("PATCH /v1/projects/{project}", s.authenticated(s.renameProject))
	mux.Handle("DELETE /v1/projects/{project}", s.authenticated(s.deleteProject))

	mux.Handle("GET /v1/projects/{project}/variables", s.authenticated(s.getVariables))
	mux.Handle("PUT /v1/projects/{project}/variables", s.authenticated(s.replaceVariables))
	mux.Handle("GET /v1/projects/{project}/variables/{key}", s.authenticated(s.getVariable))
	mux.Handle("PUT /v1/projects/{project}/variables/{key}", s.authenticated(s.setVariable))
	mux.Handle("DELETE /v1/projects/{project}/variables/{key}", s.authenticated(s.deleteVariable))

	return logRequests(logger, recoverPanics(logger, mux))
}

func (s *server) liveness(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *server) readiness(w http.ResponseWriter, r *http.Request) {
	if err := s.health.Ping(r.Context()); err != nil {
		s.logger.WarnContext(r.Context(), "not ready", slog.Any("error", err))
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}
