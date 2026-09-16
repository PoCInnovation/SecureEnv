// Package project implements the SecureEnv use cases on top of a Store.
package project

import (
	"context"
	"errors"
	"fmt"

	"github.com/PoCInnovation/SecureEnv/internal/domain"
)

// AnyVersion disables the optimistic concurrency check of ReplaceVariables.
const AnyVersion domain.Version = -1

// maxAttempts bounds the read-modify-write retries of single variable updates.
const maxAttempts = 3

// Service exposes the operations available on projects and their variables.
type Service struct {
	store Store
}

// NewService returns a Service backed by store.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// List returns every project.
func (s *Service) List(ctx context.Context) ([]domain.ProjectName, error) {
	return s.store.List(ctx)
}

// Create registers an empty project.
func (s *Service) Create(ctx context.Context, rawName string) error {
	name, err := domain.NewProjectName(rawName)
	if err != nil {
		return err
	}
	_, err = s.store.Write(ctx, name, domain.Variables{}, domain.NoVersion)
	return existsOnConflict(err)
}

// Info returns the metadata of a project.
func (s *Service) Info(ctx context.Context, rawName string) (domain.ProjectInfo, error) {
	name, err := domain.NewProjectName(rawName)
	if err != nil {
		return domain.ProjectInfo{}, err
	}
	return s.store.Info(ctx, name)
}

// Delete permanently removes a project.
func (s *Service) Delete(ctx context.Context, rawName string) error {
	name, err := domain.NewProjectName(rawName)
	if err != nil {
		return err
	}
	return s.store.Delete(ctx, name)
}

// Rename moves a project and its whole history to a new name.
//
// The store has no atomic move, so versions are copied one by one. The
// destination is created with check-and-set so an existing project is never
// overwritten, and a partial copy is removed before returning the error.
func (s *Service) Rename(ctx context.Context, rawFrom, rawTo string) error {
	from, err := domain.NewProjectName(rawFrom)
	if err != nil {
		return err
	}
	to, err := domain.NewProjectName(rawTo)
	if err != nil {
		return err
	}
	if from == to {
		return fmt.Errorf("%w: %s", domain.ErrProjectExists, to)
	}

	history, err := s.store.History(ctx, from)
	if err != nil {
		return err
	}
	if len(history) == 0 {
		history = []domain.Snapshot{{}}
	}

	expected := domain.NoVersion
	for i, snapshot := range history {
		expected, err = s.store.Write(ctx, to, snapshot.Variables, expected)
		if err != nil {
			if i == 0 {
				return existsOnConflict(err)
			}
			return errors.Join(fmt.Errorf("copy version %d: %w", snapshot.Version, err), s.store.Delete(ctx, to))
		}
	}
	return s.store.Delete(ctx, from)
}

// Variables returns the latest variables of a project.
func (s *Service) Variables(ctx context.Context, rawName string) (domain.Snapshot, error) {
	name, err := domain.NewProjectName(rawName)
	if err != nil {
		return domain.Snapshot{}, err
	}
	return s.store.Read(ctx, name)
}

// Variable returns the value of a single variable.
func (s *Service) Variable(ctx context.Context, rawName, key string) (string, error) {
	snapshot, err := s.Variables(ctx, rawName)
	if err != nil {
		return "", err
	}
	value, ok := snapshot.Variables.Get(key)
	if !ok {
		return "", fmt.Errorf("%w: %s", domain.ErrVariableNotFound, key)
	}
	return value, nil
}

// SetVariable creates or updates a variable and returns the new version.
func (s *Service) SetVariable(ctx context.Context, rawName, rawKey, value string) (domain.Version, error) {
	key, err := domain.NewVariableKey(rawKey)
	if err != nil {
		return 0, err
	}
	return s.update(ctx, rawName, func(vars domain.Variables) (domain.Variables, error) {
		return vars.With(key, value), nil
	})
}

// DeleteVariable removes a variable and returns the new version.
func (s *Service) DeleteVariable(ctx context.Context, rawName, key string) (domain.Version, error) {
	return s.update(ctx, rawName, func(vars domain.Variables) (domain.Variables, error) {
		if _, ok := vars.Get(key); !ok {
			return domain.Variables{}, fmt.Errorf("%w: %s", domain.ErrVariableNotFound, key)
		}
		return vars.Without(key), nil
	})
}

// ReplaceVariables overwrites every variable of a project. When expected is
// not AnyVersion the write fails with domain.ErrVersionConflict if the project
// changed since that version.
func (s *Service) ReplaceVariables(ctx context.Context, rawName string, raw map[string]string, expected domain.Version) (domain.Version, error) {
	name, err := domain.NewProjectName(rawName)
	if err != nil {
		return 0, err
	}
	vars, err := domain.NewVariables(raw)
	if err != nil {
		return 0, err
	}
	if expected == AnyVersion {
		current, err := s.store.Read(ctx, name)
		if err != nil {
			return 0, err
		}
		expected = current.Version
	}
	return s.store.Write(ctx, name, vars, expected)
}

// update applies change to the latest variables with check-and-set, retrying
// when another client wrote in between.
func (s *Service) update(ctx context.Context, rawName string, change func(domain.Variables) (domain.Variables, error)) (domain.Version, error) {
	name, err := domain.NewProjectName(rawName)
	if err != nil {
		return 0, err
	}

	for range maxAttempts - 1 {
		version, err := s.tryUpdate(ctx, name, change)
		if !errors.Is(err, domain.ErrVersionConflict) {
			return version, err
		}
	}
	return s.tryUpdate(ctx, name, change)
}

func (s *Service) tryUpdate(ctx context.Context, name domain.ProjectName, change func(domain.Variables) (domain.Variables, error)) (domain.Version, error) {
	current, err := s.store.Read(ctx, name)
	if err != nil {
		return 0, err
	}
	next, err := change(current.Variables)
	if err != nil {
		return 0, err
	}
	return s.store.Write(ctx, name, next, current.Version)
}

func existsOnConflict(err error) error {
	if errors.Is(err, domain.ErrVersionConflict) {
		return fmt.Errorf("%w: %w", domain.ErrProjectExists, err)
	}
	return err
}
