package domain_test

import (
	"slices"
	"testing"

	"github.com/PoCInnovation/SecureEnv/internal/domain"
)

func mustVars(t *testing.T, m map[string]string) domain.Variables {
	t.Helper()
	vars, err := domain.NewVariables(m)
	if err != nil {
		t.Fatal(err)
	}
	return vars
}

func TestDiff(t *testing.T) {
	t.Parallel()

	local := mustVars(t, map[string]string{"SAME": "x", "CHANGED": "new", "ADDED": "1", "ADDED_2": "2"})
	remote := mustVars(t, map[string]string{"SAME": "x", "CHANGED": "old", "REMOVED": "gone"})

	got := domain.Diff(local, remote)

	if want := []string{"ADDED", "ADDED_2"}; !slices.Equal(got.Added, want) {
		t.Errorf("Added = %v, want %v", got.Added, want)
	}
	if want := []string{"CHANGED"}; !slices.Equal(got.Modified, want) {
		t.Errorf("Modified = %v, want %v", got.Modified, want)
	}
	if want := []string{"REMOVED"}; !slices.Equal(got.Removed, want) {
		t.Errorf("Removed = %v, want %v", got.Removed, want)
	}
	if got.InSync() {
		t.Error("InSync() = true, want false")
	}
}

func TestDiffInSync(t *testing.T) {
	t.Parallel()

	vars := mustVars(t, map[string]string{"A": "1"})
	if got := domain.Diff(vars, vars); !got.InSync() {
		t.Fatalf("identical sets should be in sync, got %+v", got)
	}
	if got := domain.Diff(domain.Variables{}, domain.Variables{}); !got.InSync() {
		t.Fatalf("zero values should be in sync, got %+v", got)
	}
}
