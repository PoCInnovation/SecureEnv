package dotenv_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/PoCInnovation/SecureEnv/internal/dotenv"
)

func TestSaveAndLoad(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), ".env")
	if err := dotenv.Save(path, dotenv.New(map[string]string{"A": "1"})); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("permissions = %o, want 600", perm)
	}

	loaded, err := dotenv.Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.Values()["A"] != "1" {
		t.Fatalf("Load() = %v", loaded.Values())
	}
}

func TestLoadMissingFile(t *testing.T) {
	t.Parallel()

	_, err := dotenv.Load(filepath.Join(t.TempDir(), "missing"))
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Load() error = %v, want fs.ErrNotExist", err)
	}
}
