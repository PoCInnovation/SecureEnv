//go:build integration

package vaultstore_test

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"

	vault "github.com/hashicorp/vault/api"

	"github.com/PoCInnovation/SecureEnv/internal/domain"
	"github.com/PoCInnovation/SecureEnv/internal/vaultstore"
)

// healthQueryParams are documented on
// https://developer.hashicorp.com/vault/api-docs/system/health but not
// declared in the generated OpenAPI document.
var healthQueryParams = []string{
	"standbyok", "perfstandbyok", "activecode", "standbycode", "drsecondarycode",
	"haunhealthycode", "performancestandbycode", "removedcode", "sealedcode", "uninitcode",
}

// TestVaultAPIConformance records the HTTP traffic of the store against a real
// Vault and checks it twice:
//
//  1. every request matches an operation, parameter and body field declared in
//     the OpenAPI document Vault serves at sys/internal/specs/openapi;
//  2. replaying the same requests on the in-memory Vault emulator used by the
//     unit tests yields the same status codes and response fields, so unit
//     tests exercise a faithful copy of the Vault API.
func TestVaultAPIConformance(t *testing.T) {
	if os.Getenv("VAULT_ADDR") == "" || os.Getenv("VAULT_TOKEN") == "" {
		t.Skip("VAULT_ADDR and VAULT_TOKEN are required")
	}

	recorder, client := recordingClient(t)
	prefix := "secureenv-conformance/" + strings.ToLower(rand.Text())
	t.Cleanup(func() { cleanup(t, client, prefix) })

	runScenario(t, vaultstore.New(client, vaultstore.Options{Prefix: prefix}))

	calls := recorder.snapshot()
	if len(calls) == 0 {
		t.Fatal("no request recorded")
	}

	spec := fetchOpenAPI(t, client)
	for _, call := range calls {
		spec.check(t, call)
	}

	compareWithEmulator(t, calls)
}

// runScenario exercises every Store method, including the error paths.
func runScenario(t *testing.T, store *vaultstore.Store) {
	t.Helper()
	ctx := t.Context()
	name := appProject(t)
	vars := func(raw map[string]string) domain.Variables {
		v, err := domain.NewVariables(raw)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}

	steps := []struct {
		desc string
		run  func() error
		want error
	}{
		{"ping", func() error { return store.Ping(ctx) }, nil},
		{"list empty", func() error { _, err := store.List(ctx); return err }, nil},
		{"read missing", func() error { _, err := store.Read(ctx, name); return err }, domain.ErrProjectNotFound},
		{"info missing", func() error { _, err := store.Info(ctx, name); return err }, domain.ErrProjectNotFound},
		{"create", func() error {
			_, err := store.Write(ctx, name, vars(map[string]string{"A": "1"}), domain.NoVersion)
			return err
		}, nil},
		{"create again", func() error {
			_, err := store.Write(ctx, name, domain.Variables{}, domain.NoVersion)
			return err
		}, domain.ErrVersionConflict},
		{"update", func() error {
			_, err := store.Write(ctx, name, vars(map[string]string{"A": "2", "B": "x"}), 1)
			return err
		}, nil},
		{"read", func() error { _, err := store.Read(ctx, name); return err }, nil},
		{"info", func() error { _, err := store.Info(ctx, name); return err }, nil},
		{"history", func() error { _, err := store.History(ctx, name); return err }, nil},
		{"list", func() error { _, err := store.List(ctx); return err }, nil},
		{"delete", func() error { return store.Delete(ctx, name) }, nil},
		{"delete missing", func() error { return store.Delete(ctx, name) }, domain.ErrProjectNotFound},
	}
	for _, step := range steps {
		if err := step.run(); (step.want == nil && err != nil) || (step.want != nil && !errors.Is(err, step.want)) {
			t.Fatalf("%s: error = %v, want %v", step.desc, err, step.want)
		}
	}
}

type call struct {
	method   string
	path     string // without the /v1 prefix
	query    url.Values
	header   http.Header
	body     []byte
	status   int
	response []byte
}

type recorder struct {
	base  http.RoundTripper
	mu    sync.Mutex
	calls []call
}

func (r *recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		var err error
		if body, err = io.ReadAll(req.Body); err != nil {
			return nil, err
		}
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	resp, err := r.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	response, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewReader(response))

	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, call{
		method:   req.Method,
		path:     strings.TrimPrefix(req.URL.Path, "/v1"),
		query:    req.URL.Query(),
		header:   req.Header.Clone(),
		body:     body,
		status:   resp.StatusCode,
		response: response,
	})
	return resp, nil
}

func (r *recorder) snapshot() []call {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.calls)
}

func recordingClient(t *testing.T) (*recorder, *vault.Client) {
	t.Helper()
	config := vault.DefaultConfig()
	config.MaxRetries = 0
	rec := &recorder{base: config.HttpClient.Transport}
	config.HttpClient.Transport = rec
	client, err := vault.NewClient(config)
	if err != nil {
		t.Fatal(err)
	}
	return rec, client
}

// openAPI is the subset of an OpenAPI 3 document needed for the checks.
type openAPI struct {
	Paths      map[string]map[string]json.RawMessage `json:"paths"`
	Components struct {
		Schemas map[string]schema `json:"schemas"`
	} `json:"components"`
}

type schema struct {
	Ref        string            `json:"$ref"`
	Properties map[string]schema `json:"properties"`
}

type operation struct {
	OperationID string `json:"operationId"`
	Parameters  []struct {
		Name string `json:"name"`
		In   string `json:"in"`
	} `json:"parameters"`
	RequestBody struct {
		Content map[string]struct {
			Schema schema `json:"schema"`
		} `json:"content"`
	} `json:"requestBody"`
}

func fetchOpenAPI(t *testing.T, client *vault.Client) *openAPI {
	t.Helper()
	secret, err := client.Logical().ReadRawWithContext(t.Context(), "sys/internal/specs/openapi")
	if err != nil {
		t.Fatalf("fetch OpenAPI document: %v", err)
	}
	defer secret.Body.Close()
	var spec openAPI
	if err := json.NewDecoder(secret.Body).Decode(&spec); err != nil {
		t.Fatalf("decode OpenAPI document: %v", err)
	}
	if len(spec.Paths) == 0 {
		t.Fatal("OpenAPI document has no paths")
	}
	return &spec
}

// check asserts that c is a request the Vault API declares.
//
// It follows the conventions of https://developer.hashicorp.com/vault/api-docs:
// "Vault currently considers PUT and POST to be synonyms", and LIST operations
// may be sent as GET with ?list=true (declared on the path with a trailing /).
func (s *openAPI) check(t *testing.T, c call) {
	t.Helper()
	path, method := c.path, strings.ToLower(c.method)
	if method == "put" {
		method = "post"
	}
	if method == "get" && c.query.Get("list") == "true" && !strings.HasSuffix(path, "/") {
		path += "/"
	}

	template, item := s.match(path)
	if item == nil {
		t.Errorf("%s %s: no matching path in the Vault OpenAPI document", c.method, c.path)
		return
	}
	raw, ok := item[method]
	if !ok {
		t.Errorf("%s %s: %s does not declare %s", c.method, c.path, template, c.method)
		return
	}
	var op operation
	if err := json.Unmarshal(raw, &op); err != nil {
		t.Fatal(err)
	}

	// Vault accepts request fields as query parameters on reads, e.g. the
	// documented ?version= of kv-v2-read, so fields of the path's request
	// schemas are valid query parameters too.
	allowed := s.requestFields(item)
	for _, p := range op.Parameters {
		allowed = append(allowed, p.Name)
	}
	if template == "/sys/health" {
		allowed = append(allowed, healthQueryParams...)
	}
	for param := range c.query {
		if !slices.Contains(allowed, param) {
			t.Errorf("%s %s (%s): query parameter %q is not declared", c.method, c.path, op.OperationID, param)
		}
	}

	if len(c.body) == 0 {
		return
	}
	var body map[string]any
	if err := json.Unmarshal(c.body, &body); err != nil {
		t.Errorf("%s %s: body is not a JSON object: %v", c.method, c.path, err)
		return
	}
	fields := s.fields(op.RequestBody.Content["application/json"].Schema)
	for field := range body {
		if !slices.Contains(fields, field) {
			t.Errorf("%s %s (%s): body field %q is not declared", c.method, c.path, op.OperationID, field)
		}
	}
}

// match finds the path template matching path, preferring literal segments.
func (s *openAPI) match(path string) (string, map[string]json.RawMessage) {
	var best string
	for template := range s.Paths {
		pattern := regexp.QuoteMeta(template)
		pattern = strings.ReplaceAll(pattern, `\{path\}`, `.+?`)
		pattern = regexp.MustCompile(`\\\{[^}]+\\\}`).ReplaceAllString(pattern, `[^/]+`)
		if regexp.MustCompile("^"+pattern+"$").MatchString(path) && len(template) > len(best) {
			best = template
		}
	}
	return best, s.Paths[best]
}

func (s *openAPI) requestFields(item map[string]json.RawMessage) []string {
	var fields []string
	for method, raw := range item {
		if !slices.Contains([]string{"get", "post", "put", "patch", "delete"}, method) {
			continue
		}
		var op operation
		if json.Unmarshal(raw, &op) == nil {
			fields = append(fields, s.fields(op.RequestBody.Content["application/json"].Schema)...)
		}
	}
	return fields
}

func (s *openAPI) fields(sc schema) []string {
	if sc.Ref != "" {
		sc = s.Components.Schemas[sc.Ref[strings.LastIndex(sc.Ref, "/")+1:]]
	}
	var fields []string
	for name := range sc.Properties {
		fields = append(fields, name)
	}
	return fields
}

// compareWithEmulator replays calls on a fresh emulator and compares the
// responses with the ones recorded from Vault.
func compareWithEmulator(t *testing.T, calls []call) {
	t.Helper()
	fake := &fakeVault{mount: "secret", secrets: map[string][]map[string]any{}}
	server := httptest.NewServer(fake)
	defer server.Close()

	for _, c := range calls {
		req, err := http.NewRequestWithContext(t.Context(), c.method, server.URL+"/v1"+c.path+"?"+c.query.Encode(), bytes.NewReader(c.body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header = c.header.Clone()
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		response, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		desc := c.method + " " + c.path
		if resp.StatusCode != c.status {
			t.Errorf("%s: emulator status %d, Vault %d", desc, resp.StatusCode, c.status)
			continue
		}
		if len(c.response) == 0 && len(response) == 0 {
			continue
		}
		var real, emulated any
		if err := json.Unmarshal(c.response, &real); err != nil {
			t.Errorf("%s: Vault body is not JSON: %s", desc, c.response)
			continue
		}
		if err := json.Unmarshal(response, &emulated); err != nil {
			t.Errorf("%s: emulator body is not JSON: %s", desc, response)
			continue
		}
		for _, diff := range subsetDiff("", emulated, real) {
			t.Errorf("%s: %s", desc, diff)
		}
	}
}

// subsetDiff reports fields of emulated that Vault does not return, or returns
// with another JSON type. Vault may return extra fields.
func subsetDiff(path string, emulated, real any) []string {
	switch e := emulated.(type) {
	case map[string]any:
		r, ok := real.(map[string]any)
		if !ok {
			return []string{fmt.Sprintf("%s: emulator returns an object, Vault %T", orRoot(path), real)}
		}
		var diffs []string
		for key, value := range e {
			realValue, exists := r[key]
			if !exists {
				diffs = append(diffs, fmt.Sprintf("%s.%s: returned by the emulator but not by Vault", orRoot(path), key))
				continue
			}
			diffs = append(diffs, subsetDiff(path+"."+key, value, realValue)...)
		}
		return diffs
	case []any:
		r, ok := real.([]any)
		if !ok {
			return []string{fmt.Sprintf("%s: emulator returns an array, Vault %T", orRoot(path), real)}
		}
		if len(e) > 0 && len(r) > 0 {
			return subsetDiff(path+"[0]", e[0], r[0])
		}
		return nil
	default:
		if fmt.Sprintf("%T", emulated) != fmt.Sprintf("%T", real) {
			return []string{fmt.Sprintf("%s: emulator returns %T, Vault %T", orRoot(path), emulated, real)}
		}
		return nil
	}
}

func orRoot(path string) string {
	if path == "" {
		return "$"
	}
	return "$" + path
}
