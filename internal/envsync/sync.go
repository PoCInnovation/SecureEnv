// Package envsync synchronises a local .env file with a remote project.
//
// Local SECURE_ENV_* entries (project name, token, API URL) are never sent
// and are preserved on pull. Operations that would silently lose data refuse
// to run unless forced.
package envsync

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"slices"
	"strings"

	"github.com/PoCInnovation/SecureEnv/internal/domain"
	"github.com/PoCInnovation/SecureEnv/internal/dotenv"
)

// Remote is the API used to synchronise.
type Remote interface {
	Variables(ctx context.Context, project string) (domain.Snapshot, error)
	ReplaceVariables(ctx context.Context, project string, vars map[string]string, expected domain.Version) (domain.Version, error)
}

// Result describes the local and remote state around an operation. Changes
// are always expressed from the local file's point of view.
type Result struct {
	RemoteVersion domain.Version
	Changes       domain.Changes
}

// UnpushedChangesError is returned by Pull when local changes would be lost.
type UnpushedChangesError struct{ Keys []string }

func (e *UnpushedChangesError) Error() string {
	return fmt.Sprintf("local changes would be overwritten: %s (push them first or force the pull)", strings.Join(e.Keys, ", "))
}

// RemoteOnlyVariablesError is returned by Push when remote variables missing
// from the local file would be deleted.
type RemoteOnlyVariablesError struct{ Keys []string }

func (e *RemoteOnlyVariablesError) Error() string {
	return fmt.Sprintf("remote variables missing locally would be deleted: %s (pull first or force the push)", strings.Join(e.Keys, ", "))
}

// Syncer synchronises one .env file.
type Syncer struct {
	remote Remote
	path   string
}

// New returns a Syncer for the .env file at path.
func New(remote Remote, path string) *Syncer {
	return &Syncer{remote: remote, path: path}
}

// Status compares the local file with the remote project.
func (s *Syncer) Status(ctx context.Context, project string) (Result, error) {
	_, local, remote, err := s.load(ctx, project)
	if err != nil {
		return Result{}, err
	}
	return Result{RemoteVersion: remote.Version, Changes: domain.Diff(local, remote.Variables)}, nil
}

// Pull replaces the local project variables with the remote ones.
func (s *Syncer) Pull(ctx context.Context, project string, force bool) (Result, error) {
	file, local, remote, err := s.load(ctx, project)
	if err != nil {
		return Result{}, err
	}

	changes := domain.Diff(local, remote.Variables)
	if lost := slices.Concat(changes.Added, changes.Modified); len(lost) > 0 && !force {
		return Result{}, &UnpushedChangesError{Keys: lost}
	}

	values := file.Reserved()
	maps.Copy(values, remote.Variables.Map())
	if err := dotenv.Save(s.path, dotenv.New(values)); err != nil {
		return Result{}, err
	}
	return Result{RemoteVersion: remote.Version, Changes: changes}, nil
}

// Push replaces the remote variables with the local ones. The write is
// conditioned on the version that was compared, so a concurrent change fails
// with domain.ErrVersionConflict instead of being overwritten.
func (s *Syncer) Push(ctx context.Context, project string, force bool) (Result, error) {
	_, local, remote, err := s.load(ctx, project)
	if err != nil {
		return Result{}, err
	}

	changes := domain.Diff(local, remote.Variables)
	if changes.InSync() {
		return Result{RemoteVersion: remote.Version, Changes: changes}, nil
	}
	if len(changes.Removed) > 0 && !force {
		return Result{}, &RemoteOnlyVariablesError{Keys: changes.Removed}
	}

	version, err := s.remote.ReplaceVariables(ctx, project, local.Map(), remote.Version)
	if err != nil {
		return Result{}, err
	}
	return Result{RemoteVersion: version, Changes: changes}, nil
}

func (s *Syncer) load(ctx context.Context, project string) (dotenv.File, domain.Variables, domain.Snapshot, error) {
	file, err := dotenv.Load(s.path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return dotenv.File{}, domain.Variables{}, domain.Snapshot{}, err
	}

	local, err := domain.NewVariables(file.Unreserved())
	if err != nil {
		return dotenv.File{}, domain.Variables{}, domain.Snapshot{}, fmt.Errorf("%s: %w", s.path, err)
	}

	remote, err := s.remote.Variables(ctx, project)
	if err != nil {
		return dotenv.File{}, domain.Variables{}, domain.Snapshot{}, err
	}
	return file, local, remote, nil
}
