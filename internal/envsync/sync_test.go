package envsync_test

import (
	"context"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/PoCInnovation/SecureEnv/internal/domain"
	"github.com/PoCInnovation/SecureEnv/internal/dotenv"
	"github.com/PoCInnovation/SecureEnv/internal/envsync"
	"github.com/PoCInnovation/SecureEnv/internal/project"
	"github.com/PoCInnovation/SecureEnv/internal/project/projecttest"
)

// remote adapts project.Service to envsync.Remote, like the API client does.
type remote struct{ svc *project.Service }

func (r remote) Variables(ctx context.Context, name string) (domain.Snapshot, error) {
	return r.svc.Variables(ctx, name)
}

func (r remote) ReplaceVariables(ctx context.Context, name string, vars map[string]string, expected domain.Version) (domain.Version, error) {
	return r.svc.ReplaceVariables(ctx, name, vars, expected)
}

func setup(t *testing.T, local string) (*envsync.Syncer, *projecttest.MemStore, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".env")
	if local != "" {
		if err := os.WriteFile(path, []byte(local), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	store := projecttest.NewMemStore()
	return envsync.New(remote{project.NewService(store)}, path), store, path
}

func readEnv(t *testing.T, path string) map[string]string {
	t.Helper()
	file, err := dotenv.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return file.Values()
}

func TestStatus(t *testing.T) {
	t.Parallel()
	syncer, store, _ := setup(t, "SECURE_ENV_PROJECT=app\nSAME=1\nCHANGED=local\nLOCAL_ONLY=x\n")
	store.Seed("app", map[string]string{"SAME": "1", "CHANGED": "remote", "REMOTE_ONLY": "y"})

	status, err := syncer.Status(t.Context(), "app")
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}
	if status.RemoteVersion != 1 {
		t.Errorf("RemoteVersion = %d", status.RemoteVersion)
	}
	if !slices.Equal(status.Changes.Added, []string{"LOCAL_ONLY"}) ||
		!slices.Equal(status.Changes.Modified, []string{"CHANGED"}) ||
		!slices.Equal(status.Changes.Removed, []string{"REMOTE_ONLY"}) {
		t.Fatalf("Changes = %+v", status.Changes)
	}
}

func TestStatusWithoutLocalFile(t *testing.T) {
	t.Parallel()
	syncer, store, _ := setup(t, "")
	store.Seed("app", map[string]string{"A": "1"})

	status, err := syncer.Status(t.Context(), "app")
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}
	if !slices.Equal(status.Changes.Removed, []string{"A"}) {
		t.Fatalf("Changes = %+v", status.Changes)
	}
}

func TestStatusRejectsInvalidLocalKeys(t *testing.T) {
	t.Parallel()
	syncer, store, _ := setup(t, "not-valid=1\n")
	store.Seed("app", map[string]string{})

	if _, err := syncer.Status(t.Context(), "app"); !errors.Is(err, domain.ErrInvalidVariableKey) {
		t.Fatalf("Status() error = %v, want ErrInvalidVariableKey", err)
	}
}

func TestPull(t *testing.T) {
	t.Parallel()
	syncer, store, path := setup(t, "SECURE_ENV_PROJECT=app\nSECURE_ENV_TOKEN=t\n")
	store.Seed("app", map[string]string{"A": "1", "B": "2"})

	result, err := syncer.Pull(t.Context(), "app", false)
	if err != nil {
		t.Fatalf("Pull() error: %v", err)
	}
	if result.RemoteVersion != 1 {
		t.Errorf("RemoteVersion = %d", result.RemoteVersion)
	}
	want := map[string]string{"SECURE_ENV_PROJECT": "app", "SECURE_ENV_TOKEN": "t", "A": "1", "B": "2"}
	if got := readEnv(t, path); !maps.Equal(got, want) {
		t.Fatalf(".env = %v, want %v", got, want)
	}
}

func TestPullCreatesMissingFile(t *testing.T) {
	t.Parallel()
	syncer, store, path := setup(t, "")
	store.Seed("app", map[string]string{"A": "1"})

	if _, err := syncer.Pull(t.Context(), "app", false); err != nil {
		t.Fatalf("Pull() error: %v", err)
	}
	if got := readEnv(t, path); !maps.Equal(got, map[string]string{"A": "1"}) {
		t.Fatalf(".env = %v", got)
	}
}

func TestPullProtectsUnpushedChanges(t *testing.T) {
	t.Parallel()

	for name, local := range map[string]string{
		"local only":     "A=1\nNEW=x\n",
		"modified local": "A=changed\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			syncer, store, path := setup(t, local)
			store.Seed("app", map[string]string{"A": "1"})

			_, err := syncer.Pull(t.Context(), "app", false)
			var unpushed *envsync.UnpushedChangesError
			if !errors.As(err, &unpushed) {
				t.Fatalf("Pull() error = %v, want *UnpushedChangesError", err)
			}
			before, _ := dotenv.Parse(strings.NewReader(local))
			if got := readEnv(t, path); !maps.Equal(got, before.Values()) {
				t.Fatal(".env must be left untouched")
			}

			if _, err := syncer.Pull(t.Context(), "app", true); err != nil {
				t.Fatalf("forced Pull() error: %v", err)
			}
			if got := readEnv(t, path); !maps.Equal(got, map[string]string{"A": "1"}) {
				t.Fatalf("forced .env = %v", got)
			}
		})
	}
}

func TestPush(t *testing.T) {
	t.Parallel()
	syncer, store, _ := setup(t, "SECURE_ENV_TOKEN=never-pushed\nA=1\nB=2\n")
	store.Seed("app", map[string]string{"A": "old"})

	result, err := syncer.Push(t.Context(), "app", false)
	if err != nil {
		t.Fatalf("Push() error: %v", err)
	}
	if result.RemoteVersion != 2 {
		t.Errorf("RemoteVersion = %d, want 2", result.RemoteVersion)
	}
	if !slices.Equal(result.Changes.Added, []string{"B"}) || !slices.Equal(result.Changes.Modified, []string{"A"}) {
		t.Errorf("Changes = %+v", result.Changes)
	}
	if got := store.Latest("app"); !maps.Equal(got, map[string]string{"A": "1", "B": "2"}) {
		t.Fatalf("remote = %v", got)
	}
}

func TestPushInSyncDoesNotWrite(t *testing.T) {
	t.Parallel()
	syncer, store, _ := setup(t, "A=1\n")
	store.Seed("app", map[string]string{"A": "1"})

	result, err := syncer.Push(t.Context(), "app", false)
	if err != nil {
		t.Fatalf("Push() error: %v", err)
	}
	if result.RemoteVersion != 1 || !result.Changes.InSync() {
		t.Fatalf("result = %+v, want untouched v1", result)
	}
}

func TestPushProtectsRemoteOnlyVariables(t *testing.T) {
	t.Parallel()
	syncer, store, _ := setup(t, "A=1\n")
	store.Seed("app", map[string]string{"A": "1", "REMOTE": "keep me"})

	_, err := syncer.Push(t.Context(), "app", false)
	var removed *envsync.RemoteOnlyVariablesError
	if !errors.As(err, &removed) || !slices.Equal(removed.Keys, []string{"REMOTE"}) {
		t.Fatalf("Push() error = %v, want *RemoteOnlyVariablesError{REMOTE}", err)
	}
	if got := store.Latest("app"); len(got) != 2 {
		t.Fatalf("remote must be left untouched, got %v", got)
	}

	if _, err := syncer.Push(t.Context(), "app", true); err != nil {
		t.Fatalf("forced Push() error: %v", err)
	}
	if got := store.Latest("app"); !maps.Equal(got, map[string]string{"A": "1"}) {
		t.Fatalf("remote = %v", got)
	}
}

func TestPushDetectsConcurrentChange(t *testing.T) {
	t.Parallel()
	syncer, store, _ := setup(t, "A=1\n")
	store.Seed("app", map[string]string{})
	store.BeforeNextWrite(func() { store.Seed("app", map[string]string{"OTHER": "x"}) })

	if _, err := syncer.Push(t.Context(), "app", true); !errors.Is(err, domain.ErrVersionConflict) {
		t.Fatalf("Push() error = %v, want ErrVersionConflict", err)
	}
}

func TestUnknownProject(t *testing.T) {
	t.Parallel()
	syncer, _, _ := setup(t, "A=1\n")

	if _, err := syncer.Push(t.Context(), "ghost", false); !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("Push() error = %v", err)
	}
	if _, err := syncer.Pull(t.Context(), "ghost", false); !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("Pull() error = %v", err)
	}
}
