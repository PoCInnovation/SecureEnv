package project_test

import (
	"errors"
	"maps"
	"slices"
	"testing"

	"github.com/PoCInnovation/SecureEnv/internal/domain"
	"github.com/PoCInnovation/SecureEnv/internal/project"
	"github.com/PoCInnovation/SecureEnv/internal/project/projecttest"
)

func newService(t *testing.T) (*project.Service, *projecttest.MemStore) {
	t.Helper()
	store := projecttest.NewMemStore()
	return project.NewService(store), store
}

func TestCreate(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	svc, store := newService(t)

	if err := svc.Create(ctx, "backend"); err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if !store.Has("backend") {
		t.Fatal("project was not stored")
	}
	if err := svc.Create(ctx, "backend"); !errors.Is(err, domain.ErrProjectExists) {
		t.Fatalf("second Create() error = %v, want ErrProjectExists", err)
	}
	if err := svc.Create(ctx, "bad/name"); !errors.Is(err, domain.ErrInvalidProjectName) {
		t.Fatalf("Create(bad/name) error = %v, want ErrInvalidProjectName", err)
	}
}

func TestList(t *testing.T) {
	t.Parallel()
	svc, store := newService(t)
	store.Seed("b", map[string]string{})
	store.Seed("a", map[string]string{})

	names, err := svc.List(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, len(names))
	for i, n := range names {
		got[i] = n.String()
	}
	if !slices.Equal(got, []string{"a", "b"}) {
		t.Fatalf("List() = %v", got)
	}
}

func TestSetVariable(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	svc, store := newService(t)
	store.Seed("app", map[string]string{"KEEP": "1"})

	version, err := svc.SetVariable(ctx, "app", "NEW", "2")
	if err != nil {
		t.Fatalf("SetVariable() error: %v", err)
	}
	if version != 2 {
		t.Errorf("version = %d, want 2", version)
	}
	if got := store.Latest("app"); !maps.Equal(got, map[string]string{"KEEP": "1", "NEW": "2"}) {
		t.Fatalf("stored = %v", got)
	}
}

func TestSetVariableValidation(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	svc, store := newService(t)
	store.Seed("app", map[string]string{})

	if _, err := svc.SetVariable(ctx, "missing", "A", "1"); !errors.Is(err, domain.ErrProjectNotFound) {
		t.Errorf("unknown project error = %v", err)
	}
	if _, err := svc.SetVariable(ctx, "app", "SECURE_ENV_TOKEN", "x"); !errors.Is(err, domain.ErrReservedVariableKey) {
		t.Errorf("reserved key error = %v", err)
	}
	if _, err := svc.SetVariable(ctx, "app", "not-valid", "x"); !errors.Is(err, domain.ErrInvalidVariableKey) {
		t.Errorf("invalid key error = %v", err)
	}
}

func TestSetVariableRetriesOnConcurrentWrite(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	svc, store := newService(t)
	store.Seed("app", map[string]string{})

	// Another client writes between our read and our write.
	store.BeforeNextWrite(func() { store.Seed("app", map[string]string{"OTHER": "x"}) })

	if _, err := svc.SetVariable(ctx, "app", "MINE", "y"); err != nil {
		t.Fatalf("SetVariable() error: %v", err)
	}
	if got := store.Latest("app"); !maps.Equal(got, map[string]string{"OTHER": "x", "MINE": "y"}) {
		t.Fatalf("concurrent change was lost: %v", got)
	}
}

func TestSetVariableGivesUpAfterRepeatedConflicts(t *testing.T) {
	t.Parallel()
	store := projecttest.NewMemStore()
	store.Seed("app", map[string]string{})
	store.FailWrites("app", projecttest.WriteFailure{Err: domain.ErrVersionConflict})
	svc := project.NewService(store)

	if _, err := svc.SetVariable(t.Context(), "app", "A", "1"); !errors.Is(err, domain.ErrVersionConflict) {
		t.Fatalf("error = %v, want ErrVersionConflict", err)
	}
}

func TestVariable(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	svc, store := newService(t)
	store.Seed("app", map[string]string{"A": "1"})

	value, err := svc.Variable(ctx, "app", "A")
	if err != nil || value != "1" {
		t.Fatalf("Variable() = %q, %v", value, err)
	}
	if _, err := svc.Variable(ctx, "app", "B"); !errors.Is(err, domain.ErrVariableNotFound) {
		t.Fatalf("missing variable error = %v", err)
	}
}

func TestDeleteVariable(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	svc, store := newService(t)
	store.Seed("app", map[string]string{"A": "1", "B": "2"})

	if _, err := svc.DeleteVariable(ctx, "app", "A"); err != nil {
		t.Fatalf("DeleteVariable() error: %v", err)
	}
	if got := store.Latest("app"); !maps.Equal(got, map[string]string{"B": "2"}) {
		t.Fatalf("stored = %v", got)
	}
	if _, err := svc.DeleteVariable(ctx, "app", "A"); !errors.Is(err, domain.ErrVariableNotFound) {
		t.Fatalf("second delete error = %v, want ErrVariableNotFound", err)
	}
}

func TestReplaceVariables(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	t.Run("with expected version", func(t *testing.T) {
		t.Parallel()
		svc, store := newService(t)
		store.Seed("app", map[string]string{"OLD": "1"})

		version, err := svc.ReplaceVariables(ctx, "app", map[string]string{"NEW": "2"}, 1)
		if err != nil {
			t.Fatalf("ReplaceVariables() error: %v", err)
		}
		if version != 2 {
			t.Errorf("version = %d, want 2", version)
		}
		if got := store.Latest("app"); !maps.Equal(got, map[string]string{"NEW": "2"}) {
			t.Fatalf("stored = %v", got)
		}
	})

	t.Run("stale version is rejected", func(t *testing.T) {
		t.Parallel()
		svc, store := newService(t)
		store.Seed("app", map[string]string{}, map[string]string{"A": "1"})

		_, err := svc.ReplaceVariables(ctx, "app", map[string]string{}, 1)
		if !errors.Is(err, domain.ErrVersionConflict) {
			t.Fatalf("error = %v, want ErrVersionConflict", err)
		}
	})

	t.Run("any version", func(t *testing.T) {
		t.Parallel()
		svc, store := newService(t)
		store.Seed("app", map[string]string{}, map[string]string{"A": "1"})

		if _, err := svc.ReplaceVariables(ctx, "app", map[string]string{"B": "2"}, domain.AnyVersion); err != nil {
			t.Fatalf("error: %v", err)
		}
	})

	t.Run("unknown project", func(t *testing.T) {
		t.Parallel()
		svc, _ := newService(t)

		_, err := svc.ReplaceVariables(ctx, "ghost", map[string]string{}, domain.AnyVersion)
		if !errors.Is(err, domain.ErrProjectNotFound) {
			t.Fatalf("error = %v, want ErrProjectNotFound", err)
		}
	})

	t.Run("invalid keys", func(t *testing.T) {
		t.Parallel()
		svc, store := newService(t)
		store.Seed("app", map[string]string{})

		_, err := svc.ReplaceVariables(ctx, "app", map[string]string{"SECURE_ENV_X": "1"}, domain.AnyVersion)
		if !errors.Is(err, domain.ErrReservedVariableKey) {
			t.Fatalf("error = %v, want ErrReservedVariableKey", err)
		}
	})
}

func TestRename(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	t.Run("copies the whole history", func(t *testing.T) {
		t.Parallel()
		svc, store := newService(t)
		store.Seed("old", map[string]string{}, map[string]string{"A": "1"}, map[string]string{"A": "2"})

		if err := svc.Rename(ctx, "old", "new"); err != nil {
			t.Fatalf("Rename() error: %v", err)
		}
		if store.Has("old") {
			t.Error("old project still exists")
		}
		history, err := store.History(ctx, mustName(t, "new"))
		if err != nil {
			t.Fatal(err)
		}
		if len(history) != 3 {
			t.Fatalf("history length = %d, want 3", len(history))
		}
		if got := store.Latest("new"); !maps.Equal(got, map[string]string{"A": "2"}) {
			t.Fatalf("latest = %v", got)
		}
	})

	t.Run("target already exists", func(t *testing.T) {
		t.Parallel()
		svc, store := newService(t)
		store.Seed("old", map[string]string{"A": "1"})
		store.Seed("taken", map[string]string{"B": "2"})

		if err := svc.Rename(ctx, "old", "taken"); !errors.Is(err, domain.ErrProjectExists) {
			t.Fatalf("error = %v, want ErrProjectExists", err)
		}
		if !store.Has("old") || !maps.Equal(store.Latest("taken"), map[string]string{"B": "2"}) {
			t.Fatal("projects must be left untouched")
		}
	})

	t.Run("partial copy is rolled back", func(t *testing.T) {
		t.Parallel()
		svc, store := newService(t)
		store.Seed("old", map[string]string{"A": "1"}, map[string]string{"A": "2"})

		boom := errors.New("vault unavailable")
		store.FailWrites("new", projecttest.WriteFailure{After: 1, Err: boom})

		if err := svc.Rename(ctx, "old", "new"); !errors.Is(err, boom) {
			t.Fatalf("error = %v, want %v", err, boom)
		}
		if store.Has("new") {
			t.Error("partially copied project was not removed")
		}
		if !store.Has("old") {
			t.Error("source project must be kept on failure")
		}
	})

	t.Run("same name", func(t *testing.T) {
		t.Parallel()
		svc, store := newService(t)
		store.Seed("app", map[string]string{})

		if err := svc.Rename(ctx, "app", "app"); !errors.Is(err, domain.ErrProjectExists) {
			t.Fatalf("error = %v, want ErrProjectExists", err)
		}
	})
}

func TestDelete(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	svc, store := newService(t)
	store.Seed("app", map[string]string{})

	if err := svc.Delete(ctx, "app"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	if err := svc.Delete(ctx, "app"); !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("second Delete() error = %v", err)
	}
}

func TestInfoAndVariables(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	svc, store := newService(t)
	store.Seed("app", map[string]string{}, map[string]string{"A": "1"})

	info, err := svc.Info(ctx, "app")
	if err != nil || info.CurrentVersion != 2 {
		t.Fatalf("Info() = %+v, %v", info, err)
	}
	snapshot, err := svc.Variables(ctx, "app")
	if err != nil || snapshot.Version != 2 || snapshot.Variables.Len() != 1 {
		t.Fatalf("Variables() = %+v, %v", snapshot, err)
	}
	if _, err := svc.Info(ctx, "../etc"); !errors.Is(err, domain.ErrInvalidProjectName) {
		t.Fatalf("Info(invalid) error = %v", err)
	}
}

func mustName(t *testing.T, raw string) domain.ProjectName {
	t.Helper()
	name, err := domain.NewProjectName(raw)
	if err != nil {
		t.Fatal(err)
	}
	return name
}
