package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PoCInnovation/SecureEnv/internal/apiv1"
	"github.com/PoCInnovation/SecureEnv/internal/domain"
	"github.com/PoCInnovation/SecureEnv/internal/httpapi"
)

// stubService records calls and returns canned results.
type stubService struct {
	token string
	err   error

	projects []string
	info     domain.ProjectInfo
	snapshot domain.Snapshot
	value    string
	version  domain.Version

	gotName     string
	gotTo       string
	gotKey      string
	gotValue    string
	gotVars     map[string]string
	gotExpected domain.Version
	panic       bool
}

func (s *stubService) List(context.Context) ([]domain.ProjectName, error) {
	if s.panic {
		panic("boom")
	}
	names := make([]domain.ProjectName, len(s.projects))
	for i, p := range s.projects {
		names[i], _ = domain.NewProjectName(p)
	}
	return names, s.err
}

func (s *stubService) Create(_ context.Context, name string) error {
	s.gotName = name
	return s.err
}

func (s *stubService) Info(_ context.Context, name string) (domain.ProjectInfo, error) {
	s.gotName = name
	return s.info, s.err
}

func (s *stubService) Rename(_ context.Context, from, to string) error {
	s.gotName, s.gotTo = from, to
	return s.err
}

func (s *stubService) Delete(_ context.Context, name string) error {
	s.gotName = name
	return s.err
}

func (s *stubService) Variables(_ context.Context, name string) (domain.Snapshot, error) {
	s.gotName = name
	return s.snapshot, s.err
}

func (s *stubService) Variable(_ context.Context, name, key string) (string, error) {
	s.gotName, s.gotKey = name, key
	return s.value, s.err
}

func (s *stubService) SetVariable(_ context.Context, name, key, value string) (domain.Version, error) {
	s.gotName, s.gotKey, s.gotValue = name, key, value
	return s.version, s.err
}

func (s *stubService) DeleteVariable(_ context.Context, name, key string) (domain.Version, error) {
	s.gotName, s.gotKey = name, key
	return s.version, s.err
}

func (s *stubService) ReplaceVariables(_ context.Context, name string, vars map[string]string, expected domain.Version) (domain.Version, error) {
	s.gotName, s.gotVars, s.gotExpected = name, vars, expected
	return s.version, s.err
}

type stubHealth struct{ err error }

func (h stubHealth) Ping(context.Context) error { return h.err }

func newTestHandler(t *testing.T, svc *stubService, health httpapi.HealthChecker) http.Handler {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return httpapi.NewHandler(func(token string) (httpapi.ProjectService, error) {
		svc.token = token
		return svc, nil
	}, health, logger)
}

func do(t *testing.T, h http.Handler, method, target string, body any, headers ...string) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	switch b := body.(type) {
	case nil:
	case string:
		reader = bytes.NewBufferString(b)
	default:
		raw, err := json.Marshal(b)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequestWithContext(t.Context(), method, target, reader)
	req.Header.Set("Authorization", "Bearer s.token")
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var out T
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode %T: %v (status %d)", out, err, rec.Code)
	}
	return out
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, want, rec.Body)
	}
}

func assertErrorCode(t *testing.T, rec *httptest.ResponseRecorder, status int, code apiv1.Code) {
	t.Helper()
	assertStatus(t, rec, status)
	if got := decode[apiv1.ErrorResponse](t, rec); got.Error.Code != code {
		t.Fatalf("error code = %q, want %q", got.Error.Code, code)
	}
}

func TestAuthentication(t *testing.T) {
	t.Parallel()
	svc := &stubService{}
	h := newTestHandler(t, svc, stubHealth{})

	for _, header := range []string{"", "Basic abc", "Bearer ", "Bearer    "} {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/projects", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assertErrorCode(t, rec, http.StatusUnauthorized, apiv1.CodeUnauthorized)
		if rec.Header().Get("WWW-Authenticate") == "" {
			t.Errorf("%q: missing WWW-Authenticate header", header)
		}
	}

	rec := do(t, h, http.MethodGet, "/v1/projects", nil)
	assertStatus(t, rec, http.StatusOK)
	if svc.token != "s.token" {
		t.Fatalf("service built with token %q", svc.token)
	}
}

func TestHealth(t *testing.T) {
	t.Parallel()

	healthy := newTestHandler(t, &stubService{}, stubHealth{})
	unhealthy := newTestHandler(t, &stubService{}, stubHealth{err: errors.New("sealed")})

	for _, tc := range []struct {
		h      http.Handler
		target string
		want   int
	}{
		{healthy, "/healthz", http.StatusOK},
		{unhealthy, "/healthz", http.StatusOK},
		{healthy, "/readyz", http.StatusOK},
		{unhealthy, "/readyz", http.StatusServiceUnavailable},
	} {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.target, nil)
		rec := httptest.NewRecorder()
		tc.h.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Errorf("%s = %d, want %d", tc.target, rec.Code, tc.want)
		}
	}
}

func TestListProjects(t *testing.T) {
	t.Parallel()
	h := newTestHandler(t, &stubService{projects: []string{"a", "b"}}, stubHealth{})

	rec := do(t, h, http.MethodGet, "/v1/projects", nil)
	assertStatus(t, rec, http.StatusOK)
	if got := decode[apiv1.ProjectList](t, rec); len(got.Projects) != 2 || got.Projects[0] != "a" {
		t.Fatalf("body = %+v", got)
	}
}

func TestListProjectsEmptyIsArray(t *testing.T) {
	t.Parallel()
	h := newTestHandler(t, &stubService{}, stubHealth{})

	rec := do(t, h, http.MethodGet, "/v1/projects", nil)
	if body := rec.Body.String(); body != "{\"projects\":[]}\n" {
		t.Fatalf("body = %q", body)
	}
}

func TestCreateProject(t *testing.T) {
	t.Parallel()
	svc := &stubService{}
	h := newTestHandler(t, svc, stubHealth{})

	rec := do(t, h, http.MethodPost, "/v1/projects", apiv1.ProjectNameRequest{Name: "app"})
	assertStatus(t, rec, http.StatusCreated)
	if svc.gotName != "app" {
		t.Errorf("created %q", svc.gotName)
	}
	if loc := rec.Header().Get("Location"); loc != "/v1/projects/app" {
		t.Errorf("Location = %q", loc)
	}

	svc.err = domain.ErrProjectExists
	assertErrorCode(t, do(t, h, http.MethodPost, "/v1/projects", apiv1.ProjectNameRequest{Name: "app"}), http.StatusConflict, apiv1.CodeProjectExists)
}

func TestInvalidBodies(t *testing.T) {
	t.Parallel()
	h := newTestHandler(t, &stubService{}, stubHealth{})

	for name, body := range map[string]string{
		"malformed":      "{",
		"unknown fields": `{"name":"a","extra":1}`,
		"trailing data":  `{"name":"a"}{}`,
		"wrong type":     `{"name":1}`,
	} {
		t.Run(name, func(t *testing.T) {
			assertErrorCode(t, do(t, h, http.MethodPost, "/v1/projects", body), http.StatusBadRequest, apiv1.CodeInvalidRequest)
		})
	}

	huge := `{"name":"` + string(bytes.Repeat([]byte("a"), 2<<20)) + `"}`
	assertErrorCode(t, do(t, h, http.MethodPost, "/v1/projects", huge), http.StatusRequestEntityTooLarge, apiv1.CodeInvalidRequest)
}

func TestProjectInfo(t *testing.T) {
	t.Parallel()
	name, _ := domain.NewProjectName("app")
	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	svc := &stubService{info: domain.ProjectInfo{Name: name, CurrentVersion: 4, CreatedAt: created, UpdatedAt: created}}
	h := newTestHandler(t, svc, stubHealth{})

	rec := do(t, h, http.MethodGet, "/v1/projects/app", nil)
	assertStatus(t, rec, http.StatusOK)
	got := decode[apiv1.Project](t, rec)
	if got.Name != "app" || got.CurrentVersion != 4 || !got.CreatedAt.Equal(created) {
		t.Fatalf("body = %+v", got)
	}

	svc.err = domain.ErrProjectNotFound
	assertErrorCode(t, do(t, h, http.MethodGet, "/v1/projects/app", nil), http.StatusNotFound, apiv1.CodeProjectNotFound)
}

func TestRenameAndDeleteProject(t *testing.T) {
	t.Parallel()
	svc := &stubService{}
	h := newTestHandler(t, svc, stubHealth{})

	assertStatus(t, do(t, h, http.MethodPatch, "/v1/projects/old", apiv1.ProjectNameRequest{Name: "new"}), http.StatusNoContent)
	if svc.gotName != "old" || svc.gotTo != "new" {
		t.Errorf("renamed %q to %q", svc.gotName, svc.gotTo)
	}

	assertStatus(t, do(t, h, http.MethodDelete, "/v1/projects/app", nil), http.StatusNoContent)
	if svc.gotName != "app" {
		t.Errorf("deleted %q", svc.gotName)
	}
}

func TestGetVariables(t *testing.T) {
	t.Parallel()
	vars, _ := domain.NewVariables(map[string]string{"A": "1"})
	h := newTestHandler(t, &stubService{snapshot: domain.Snapshot{Version: 3, Variables: vars}}, stubHealth{})

	rec := do(t, h, http.MethodGet, "/v1/projects/app/variables", nil)
	assertStatus(t, rec, http.StatusOK)
	if etag := rec.Header().Get("ETag"); etag != `"3"` {
		t.Errorf("ETag = %q", etag)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control = %q, secrets must not be cached", cc)
	}
	got := decode[apiv1.Variables](t, rec)
	if got.Version != 3 || !maps.Equal(got.Variables, map[string]string{"A": "1"}) {
		t.Fatalf("body = %+v", got)
	}
}

func TestReplaceVariables(t *testing.T) {
	t.Parallel()
	svc := &stubService{version: 5}
	h := newTestHandler(t, svc, stubHealth{})
	body := apiv1.ReplaceVariablesRequest{Variables: map[string]string{"B": "2"}}

	rec := do(t, h, http.MethodPut, "/v1/projects/app/variables", body, "If-Match", `"4"`)
	assertStatus(t, rec, http.StatusOK)
	if svc.gotExpected != 4 || !maps.Equal(svc.gotVars, body.Variables) {
		t.Errorf("service got expected=%d vars=%v", svc.gotExpected, svc.gotVars)
	}
	if got := decode[apiv1.VersionResponse](t, rec); got.Version != 5 {
		t.Errorf("version = %d", got.Version)
	}
	if rec.Header().Get("ETag") != `"5"` {
		t.Errorf("ETag = %q", rec.Header().Get("ETag"))
	}

	do(t, h, http.MethodPut, "/v1/projects/app/variables", body)
	if svc.gotExpected != domain.AnyVersion {
		t.Errorf("without If-Match expected = %d, want AnyVersion", svc.gotExpected)
	}

	assertErrorCode(t, do(t, h, http.MethodPut, "/v1/projects/app/variables", body, "If-Match", "W/abc"), http.StatusBadRequest, apiv1.CodeInvalidRequest)
	assertErrorCode(t, do(t, h, http.MethodPut, "/v1/projects/app/variables", `{}`), http.StatusBadRequest, apiv1.CodeInvalidRequest)

	svc.err = domain.ErrVersionConflict
	assertErrorCode(t, do(t, h, http.MethodPut, "/v1/projects/app/variables", body, "If-Match", `"1"`), http.StatusPreconditionFailed, apiv1.CodeVersionConflict)
}

func TestSingleVariable(t *testing.T) {
	t.Parallel()
	svc := &stubService{value: "secret", version: 2}
	h := newTestHandler(t, svc, stubHealth{})

	rec := do(t, h, http.MethodGet, "/v1/projects/app/variables/KEY", nil)
	assertStatus(t, rec, http.StatusOK)
	if got := decode[apiv1.Variable](t, rec); got.Key != "KEY" || got.Value != "secret" {
		t.Fatalf("body = %+v", got)
	}

	rec = do(t, h, http.MethodPut, "/v1/projects/app/variables/KEY", apiv1.SetVariableRequest{Value: "v"})
	assertStatus(t, rec, http.StatusOK)
	if svc.gotKey != "KEY" || svc.gotValue != "v" {
		t.Errorf("set %q=%q", svc.gotKey, svc.gotValue)
	}

	rec = do(t, h, http.MethodDelete, "/v1/projects/app/variables/KEY", nil)
	assertStatus(t, rec, http.StatusOK)

	svc.err = domain.ErrVariableNotFound
	assertErrorCode(t, do(t, h, http.MethodGet, "/v1/projects/app/variables/NOPE", nil), http.StatusNotFound, apiv1.CodeVariableNotFound)
}

func TestInternalErrorsAreNotLeaked(t *testing.T) {
	t.Parallel()
	h := newTestHandler(t, &stubService{err: errors.New("vault: dial tcp 10.0.0.3:8200: connection refused")}, stubHealth{})

	rec := do(t, h, http.MethodGet, "/v1/projects", nil)
	assertStatus(t, rec, http.StatusInternalServerError)
	if got := decode[apiv1.ErrorResponse](t, rec); got.Error.Message != "internal server error" {
		t.Fatalf("message leaked internals: %q", got.Error.Message)
	}
}

func TestPanicIsRecovered(t *testing.T) {
	t.Parallel()
	h := newTestHandler(t, &stubService{panic: true}, stubHealth{})

	assertErrorCode(t, do(t, h, http.MethodGet, "/v1/projects", nil), http.StatusInternalServerError, apiv1.CodeInternal)
}

func TestUnknownRoute(t *testing.T) {
	t.Parallel()
	h := newTestHandler(t, &stubService{}, stubHealth{})

	assertStatus(t, do(t, h, http.MethodGet, "/v1/nope", nil), http.StatusNotFound)
	rec := do(t, h, http.MethodPost, "/v1/projects/app", nil)
	assertStatus(t, rec, http.StatusMethodNotAllowed)
}
