package vaultstore_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	vault "github.com/hashicorp/vault/api"
)

// fakeVault emulates the subset of the Vault HTTP API used by the store:
// KV v2 data and metadata endpoints with check-and-set, and sys/health.
type fakeVault struct {
	mount string

	mu      sync.Mutex
	secrets map[string][]map[string]any
	// fail, when non-zero, makes every KV request answer with this status.
	fail       int
	failErrors []string
}

func newFakeVault(t *testing.T) (*fakeVault, *vault.Client) {
	t.Helper()

	fake := &fakeVault{mount: "secret", secrets: map[string][]map[string]any{}}
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)

	config := vault.DefaultConfig()
	config.Address = server.URL
	config.MaxRetries = 0
	client, err := vault.NewClient(config)
	if err != nil {
		t.Fatal(err)
	}
	client.SetToken("test-token")
	return fake, client
}

func (f *fakeVault) failWith(status int, errs ...string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fail, f.failErrors = status, errs
}

func (f *fakeVault) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/v1/sys/health" {
		writeJSON(w, http.StatusOK, map[string]any{"initialized": true, "sealed": false})
		return
	}
	if r.Header.Get("X-Vault-Token") == "" {
		writeJSON(w, http.StatusForbidden, map[string]any{"errors": []string{"permission denied"}})
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if f.fail != 0 {
		writeJSON(w, f.fail, map[string]any{"errors": f.failErrors})
		return
	}

	rest, ok := strings.CutPrefix(r.URL.Path, "/v1/"+f.mount+"/")
	if !ok {
		http.NotFound(w, r)
		return
	}
	kind, secretPath, _ := strings.Cut(rest, "/")

	switch {
	case kind == "metadata" && r.Method == http.MethodGet && r.URL.Query().Get("list") == "true":
		f.list(w, strings.TrimSuffix(secretPath, "/"))
	case kind == "metadata" && r.Method == http.MethodGet:
		f.metadata(w, secretPath)
	case kind == "metadata" && r.Method == http.MethodDelete:
		delete(f.secrets, secretPath)
		w.WriteHeader(http.StatusNoContent)
	case kind == "data" && r.Method == http.MethodGet:
		f.read(w, r, secretPath)
	case kind == "data" && (r.Method == http.MethodPut || r.Method == http.MethodPost):
		f.write(w, r, secretPath)
	default:
		http.Error(w, "unsupported", http.StatusMethodNotAllowed)
	}
}

func (f *fakeVault) list(w http.ResponseWriter, prefix string) {
	if prefix != "" {
		prefix += "/"
	}
	seen := map[string]bool{}
	for secretPath := range f.secrets {
		rest, ok := strings.CutPrefix(secretPath, prefix)
		if !ok {
			continue
		}
		if head, _, nested := strings.Cut(rest, "/"); nested {
			rest = head + "/"
		}
		seen[rest] = true
	}
	if len(seen) == 0 {
		writeJSON(w, http.StatusNotFound, map[string]any{"errors": []string{}})
		return
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"keys": keys}})
}

func (f *fakeVault) metadata(w http.ResponseWriter, secretPath string) {
	versions, ok := f.secrets[secretPath]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"errors": []string{}})
		return
	}
	versionMap := map[string]any{}
	for i := range versions {
		versionMap[strconv.Itoa(i+1)] = versionMetadata(i + 1)
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{
		"current_version":      len(versions),
		"oldest_version":       1,
		"max_versions":         0,
		"cas_required":         false,
		"delete_version_after": "0s",
		"custom_metadata":      nil,
		"created_time":         timestamp(),
		"updated_time":         timestamp(),
		"versions":             versionMap,
	}})
}

func (f *fakeVault) read(w http.ResponseWriter, r *http.Request, secretPath string) {
	versions, ok := f.secrets[secretPath]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"errors": []string{}})
		return
	}
	version := len(versions)
	if raw := r.URL.Query().Get("version"); raw != "" && raw != "0" {
		version, _ = strconv.Atoi(raw)
	}
	if version < 1 || version > len(versions) {
		writeJSON(w, http.StatusNotFound, map[string]any{"errors": []string{}})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{
		"data":     versions[version-1],
		"metadata": versionMetadata(version),
	}})
}

func (f *fakeVault) write(w http.ResponseWriter, r *http.Request, secretPath string) {
	var body struct {
		Data    map[string]any `json:"data"`
		Options struct {
			CAS *int `json:"cas"`
		} `json:"options"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"errors": []string{err.Error()}})
		return
	}
	versions := f.secrets[secretPath]
	if body.Options.CAS != nil && *body.Options.CAS != len(versions) {
		writeJSON(w, http.StatusBadRequest, map[string]any{"errors": []string{
			"check-and-set parameter did not match the current version",
		}})
		return
	}
	f.secrets[secretPath] = append(versions, body.Data)
	writeJSON(w, http.StatusOK, map[string]any{"data": versionMetadata(len(versions) + 1)})
}

func versionMetadata(version int) map[string]any {
	return map[string]any{
		"version":         version,
		"created_time":    timestamp(),
		"deletion_time":   "",
		"destroyed":       false,
		"custom_metadata": nil,
	}
}

func timestamp() string { return time.Unix(1700000000, 0).UTC().Format(time.RFC3339Nano) }

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
