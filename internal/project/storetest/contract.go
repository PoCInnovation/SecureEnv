// Package storetest provides a behavioural contract every project.Store
// implementation must satisfy.
package storetest

import (
	"errors"
	"maps"
	"testing"

	"github.com/PoCInnovation/SecureEnv/internal/domain"
	"github.com/PoCInnovation/SecureEnv/internal/project"
)

// Factory returns an empty, isolated store.
type Factory func(t *testing.T) project.Store

// Run checks that stores built by newStore honour the project.Store contract.
func Run(t *testing.T, newStore Factory) {
	t.Helper()

	t.Run("empty store lists nothing", func(t *testing.T) {
		names, err := newStore(t).List(t.Context())
		if err != nil {
			t.Fatalf("List() error: %v", err)
		}
		if len(names) != 0 {
			t.Fatalf("List() = %v, want empty", names)
		}
	})

	t.Run("missing project", func(t *testing.T) {
		store := newStore(t)
		ctx := t.Context()
		name := mustName(t, "ghost")

		if _, err := store.Read(ctx, name); !errors.Is(err, domain.ErrProjectNotFound) {
			t.Errorf("Read() error = %v, want ErrProjectNotFound", err)
		}
		if _, err := store.Info(ctx, name); !errors.Is(err, domain.ErrProjectNotFound) {
			t.Errorf("Info() error = %v, want ErrProjectNotFound", err)
		}
		if _, err := store.History(ctx, name); !errors.Is(err, domain.ErrProjectNotFound) {
			t.Errorf("History() error = %v, want ErrProjectNotFound", err)
		}
		if err := store.Delete(ctx, name); !errors.Is(err, domain.ErrProjectNotFound) {
			t.Errorf("Delete() error = %v, want ErrProjectNotFound", err)
		}
	})

	t.Run("create read list", func(t *testing.T) {
		store := newStore(t)
		ctx := t.Context()
		name := mustName(t, "app")
		want := map[string]string{"URL": "postgres://db.internal:5432/app?sslmode=disable&a=b", "MULTI": "l1\nl2", "EMPTY": ""}

		version, err := store.Write(ctx, name, mustVars(t, want), domain.NoVersion)
		if err != nil {
			t.Fatalf("Write() error: %v", err)
		}
		if version != 1 {
			t.Errorf("Write() version = %d, want 1", version)
		}

		snapshot, err := store.Read(ctx, name)
		if err != nil {
			t.Fatalf("Read() error: %v", err)
		}
		if snapshot.Version != 1 || !maps.Equal(snapshot.Variables.Map(), want) {
			t.Fatalf("Read() = v%d %v, want v1 %v", snapshot.Version, snapshot.Variables.Map(), want)
		}

		info, err := store.Info(ctx, name)
		if err != nil {
			t.Fatalf("Info() error: %v", err)
		}
		if info.Name != name || info.CurrentVersion != 1 {
			t.Fatalf("Info() = %+v", info)
		}

		names, err := store.List(ctx)
		if err != nil {
			t.Fatalf("List() error: %v", err)
		}
		if len(names) != 1 || names[0] != name {
			t.Fatalf("List() = %v, want [app]", names)
		}
	})

	t.Run("check and set", func(t *testing.T) {
		store := newStore(t)
		ctx := t.Context()
		name := mustName(t, "app")

		if _, err := store.Write(ctx, name, domain.Variables{}, domain.NoVersion); err != nil {
			t.Fatal(err)
		}
		if _, err := store.Write(ctx, name, domain.Variables{}, domain.NoVersion); !errors.Is(err, domain.ErrVersionConflict) {
			t.Fatalf("create over existing error = %v, want ErrVersionConflict", err)
		}
		if _, err := store.Write(ctx, name, domain.Variables{}, 7); !errors.Is(err, domain.ErrVersionConflict) {
			t.Fatalf("stale write error = %v, want ErrVersionConflict", err)
		}
		version, err := store.Write(ctx, name, mustVars(t, map[string]string{"A": "1"}), 1)
		if err != nil {
			t.Fatalf("write at current version error: %v", err)
		}
		if version != 2 {
			t.Fatalf("version = %d, want 2", version)
		}
	})

	t.Run("history is ordered", func(t *testing.T) {
		store := newStore(t)
		ctx := t.Context()
		name := mustName(t, "app")

		expected := domain.NoVersion
		for _, value := range []string{"1", "2", "3"} {
			var err error
			expected, err = store.Write(ctx, name, mustVars(t, map[string]string{"A": value}), expected)
			if err != nil {
				t.Fatal(err)
			}
		}

		history, err := store.History(ctx, name)
		if err != nil {
			t.Fatalf("History() error: %v", err)
		}
		if len(history) != 3 {
			t.Fatalf("History() length = %d, want 3", len(history))
		}
		for i, snapshot := range history {
			wantValue := string(rune('1' + i))
			if got, _ := snapshot.Variables.Get("A"); snapshot.Version != domain.Version(i+1) || got != wantValue {
				t.Errorf("history[%d] = v%d A=%q, want v%d A=%q", i, snapshot.Version, got, i+1, wantValue)
			}
		}
	})

	t.Run("delete removes history", func(t *testing.T) {
		store := newStore(t)
		ctx := t.Context()
		name := mustName(t, "app")

		if _, err := store.Write(ctx, name, domain.Variables{}, domain.NoVersion); err != nil {
			t.Fatal(err)
		}
		if err := store.Delete(ctx, name); err != nil {
			t.Fatalf("Delete() error: %v", err)
		}
		if _, err := store.Read(ctx, name); !errors.Is(err, domain.ErrProjectNotFound) {
			t.Fatalf("Read() after delete error = %v", err)
		}
		if _, err := store.Write(ctx, name, domain.Variables{}, domain.NoVersion); err != nil {
			t.Fatalf("recreate after delete error: %v", err)
		}
	})
}

func mustName(t *testing.T, raw string) domain.ProjectName {
	t.Helper()
	name, err := domain.NewProjectName(raw)
	if err != nil {
		t.Fatal(err)
	}
	return name
}

func mustVars(t *testing.T, raw map[string]string) domain.Variables {
	t.Helper()
	vars, err := domain.NewVariables(raw)
	if err != nil {
		t.Fatal(err)
	}
	return vars
}
