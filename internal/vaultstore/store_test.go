package vaultstore_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/PoCInnovation/SecureEnv/internal/domain"
	"github.com/PoCInnovation/SecureEnv/internal/project"
	"github.com/PoCInnovation/SecureEnv/internal/project/storetest"
	"github.com/PoCInnovation/SecureEnv/internal/vaultstore"
)

func TestStoreContract(t *testing.T) {
	t.Run("without prefix", func(t *testing.T) {
		storetest.Run(t, func(t *testing.T) project.Store {
			_, client := newFakeVault(t)
			return vaultstore.New(client, vaultstore.Options{})
		})
	})
	t.Run("with prefix", func(t *testing.T) {
		storetest.Run(t, func(t *testing.T) project.Store {
			_, client := newFakeVault(t)
			return vaultstore.New(client, vaultstore.Options{Mount: "/secret/", Prefix: "teams/core"})
		})
	})
}

func TestListSkipsFoldersAndForeignSecrets(t *testing.T) {
	fake, client := newFakeVault(t)
	fake.secrets["app"] = []map[string]any{{}}
	fake.secrets["folder/nested"] = []map[string]any{{}}
	fake.secrets["not a project"] = []map[string]any{{}}

	names, err := vaultstore.New(client, vaultstore.Options{}).List(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0].String() != "app" {
		t.Fatalf("List() = %v, want [app]", names)
	}
}

func TestReadRejectsNonStringValues(t *testing.T) {
	fake, client := newFakeVault(t)
	fake.secrets["app"] = []map[string]any{{"PORT": 8080}}

	_, err := vaultstore.New(client, vaultstore.Options{}).Read(t.Context(), appProject(t))
	if err == nil {
		t.Fatal("Read() should reject non string values")
	}
}

func TestErrorMapping(t *testing.T) {
	tests := []struct {
		status int
		errors []string
		want   error
	}{
		{status: http.StatusUnauthorized, want: domain.ErrUnauthorized},
		{status: http.StatusForbidden, errors: []string{"permission denied"}, want: domain.ErrForbidden},
		{status: http.StatusBadRequest, errors: []string{"check-and-set parameter did not match the current version"}, want: domain.ErrVersionConflict},
	}
	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			fake, client := newFakeVault(t)
			fake.failWith(tt.status, tt.errors...)
			store := vaultstore.New(client, vaultstore.Options{})

			_, err := store.Write(t.Context(), appProject(t), domain.Variables{}, domain.NoVersion)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Write() error = %v, want %v", err, tt.want)
			}
		})
	}

	t.Run("other errors are not domain errors", func(t *testing.T) {
		fake, client := newFakeVault(t)
		fake.failWith(http.StatusBadRequest, "something else")

		_, err := vaultstore.New(client, vaultstore.Options{}).Write(t.Context(), appProject(t), domain.Variables{}, domain.NoVersion)
		if err == nil || errors.Is(err, domain.ErrVersionConflict) {
			t.Fatalf("Write() error = %v, want a generic error", err)
		}
	})
}

func TestWithTokenDoesNotMutateReceiver(t *testing.T) {
	_, client := newFakeVault(t)
	client.ClearToken()
	base := vaultstore.New(client, vaultstore.Options{})

	authenticated, err := base.WithToken("caller-token")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := authenticated.List(t.Context()); err != nil {
		t.Fatalf("authenticated List() error: %v", err)
	}
	if _, err := base.List(t.Context()); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("base store must stay anonymous, got error %v", err)
	}
}

func TestPing(t *testing.T) {
	_, client := newFakeVault(t)
	if err := vaultstore.New(client, vaultstore.Options{}).Ping(t.Context()); err != nil {
		t.Fatalf("Ping() error: %v", err)
	}
}

func appProject(t *testing.T) domain.ProjectName {
	t.Helper()
	name, err := domain.NewProjectName("app")
	if err != nil {
		t.Fatal(err)
	}
	return name
}

func TestPrefixIsAppliedToSecretPaths(t *testing.T) {
	fake, client := newFakeVault(t)
	store := vaultstore.New(client, vaultstore.Options{Prefix: "/teams/core/"})

	if _, err := store.Write(t.Context(), appProject(t), domain.Variables{}, domain.NoVersion); err != nil {
		t.Fatal(err)
	}
	if _, ok := fake.secrets["teams/core/app"]; !ok {
		t.Fatalf("secret stored at %v, want teams/core/app", fake.secrets)
	}
}
