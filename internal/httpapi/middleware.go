package httpapi

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/PoCInnovation/SecureEnv/internal/domain"
)

const maxBodyBytes = 1 << 20

// handlerFunc is an endpoint running with an authenticated service. Returned
// errors are rendered by writeError.
type handlerFunc func(w http.ResponseWriter, r *http.Request, svc ProjectService) error

// authenticated extracts the bearer token, builds the caller's service and
// renders any error returned by next.
func (s *server) authenticated(next handlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r)
		if !ok {
			w.Header().Set("WWW-Authenticate", `Bearer realm="secureenv"`)
			writeError(s.logger, w, r, domain.ErrUnauthorized)
			return
		}

		svc, err := s.services(token)
		if err != nil {
			writeError(s.logger, w, r, err)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		if err := next(w, r, svc); err != nil {
			writeError(s.logger, w, r, err)
		}
	})
}

func bearerToken(r *http.Request) (string, bool) {
	scheme, token, found := strings.Cut(r.Header.Get("Authorization"), " ")
	token = strings.TrimSpace(token)
	if !found || !strings.EqualFold(scheme, "Bearer") || token == "" {
		return "", false
	}
	return token, true
}

func recoverPanics(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}
			if recovered == http.ErrAbortHandler {
				panic(recovered)
			}
			logger.ErrorContext(r.Context(), "panic", slog.Any("panic", recovered))
			writeError(logger, w, r, fmt.Errorf("panic: %v", recovered))
		}()
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }

// logRequests writes one structured line per request. Tokens and bodies are
// never logged.
func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)

		level := slog.LevelInfo
		if recorder.status >= http.StatusInternalServerError {
			level = slog.LevelError
		}
		logger.LogAttrs(r.Context(), level, "request",
			slog.String("method", r.Method),
			slog.String("route", r.Pattern),
			slog.String("project", r.PathValue("project")),
			slog.Int("status", recorder.status),
			slog.Duration("duration", time.Since(start)),
			slog.String("remote_addr", r.RemoteAddr),
		)
	})
}
