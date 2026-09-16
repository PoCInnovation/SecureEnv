package dotenv

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// Load parses the .env file at path. A missing file yields an error matching
// fs.ErrNotExist.
func Load(path string) (File, error) {
	f, err := os.Open(path) //nolint:gosec // reading a user-chosen env file is the purpose
	if err != nil {
		return File{}, err
	}
	defer f.Close()

	file, err := Parse(f)
	if err != nil {
		return File{}, fmt.Errorf("%s: %w", path, err)
	}
	return file, nil
}

// Save atomically replaces the file at path with f. The file is only readable
// by its owner because it holds secrets.
func Save(path string, f File) error {
	var buf bytes.Buffer
	if err := f.Format(&buf); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".env-*.tmp")
	if err != nil {
		return fmt.Errorf("dotenv: create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("dotenv: chmod: %w", err)
	}
	if _, err := tmp.Write(buf.Bytes()); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("dotenv: write: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("dotenv: close: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("dotenv: replace %s: %w", path, err)
	}
	return nil
}
