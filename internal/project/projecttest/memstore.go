// Package projecttest provides test helpers for code built on project.Store.
package projecttest

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/PoCInnovation/SecureEnv/internal/domain"
)

// MemStore is an in-memory project.Store that mimics Vault KV v2 versioning
// and check-and-set semantics. It is safe for concurrent use.
// WriteFailure makes writes to a project fail with Err once After writes
// succeeded.
type WriteFailure struct {
	After int
	Err   error
}

type MemStore struct {
	mu       sync.Mutex
	projects map[string][]domain.Variables

	failures    map[string]WriteFailure
	beforeWrite func()
}

// NewMemStore returns an empty store.
func NewMemStore() *MemStore {
	return &MemStore{projects: map[string][]domain.Variables{}, failures: map[string]WriteFailure{}}
}

func (s *MemStore) List(context.Context) ([]domain.ProjectName, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	names := make([]domain.ProjectName, 0, len(s.projects))
	for _, raw := range slices.Sorted(maps.Keys(s.projects)) {
		name, _ := domain.NewProjectName(raw)
		names = append(names, name)
	}
	return names, nil
}

func (s *MemStore) Info(_ context.Context, name domain.ProjectName) (domain.ProjectInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	versions, ok := s.projects[name.String()]
	if !ok {
		return domain.ProjectInfo{}, domain.ErrProjectNotFound
	}
	return domain.ProjectInfo{Name: name, CurrentVersion: domain.Version(len(versions))}, nil
}

func (s *MemStore) Read(_ context.Context, name domain.ProjectName) (domain.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	versions, ok := s.projects[name.String()]
	if !ok {
		return domain.Snapshot{}, domain.ErrProjectNotFound
	}
	return domain.Snapshot{Version: domain.Version(len(versions)), Variables: versions[len(versions)-1]}, nil
}

func (s *MemStore) History(_ context.Context, name domain.ProjectName) ([]domain.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	versions, ok := s.projects[name.String()]
	if !ok {
		return nil, domain.ErrProjectNotFound
	}
	history := make([]domain.Snapshot, len(versions))
	for i, vars := range versions {
		history[i] = domain.Snapshot{Version: domain.Version(i + 1), Variables: vars}
	}
	return history, nil
}

func (s *MemStore) Write(_ context.Context, name domain.ProjectName, vars domain.Variables, expected domain.Version) (domain.Version, error) {
	if hook := s.takeHook(); hook != nil {
		hook()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if failure, ok := s.failures[name.String()]; ok {
		if failure.After == 0 {
			return 0, failure.Err
		}
		failure.After--
		s.failures[name.String()] = failure
	}
	versions := s.projects[name.String()]
	if domain.Version(len(versions)) != expected {
		return 0, domain.ErrVersionConflict
	}
	s.projects[name.String()] = append(versions, vars)
	return domain.Version(len(versions) + 1), nil
}

func (s *MemStore) Delete(_ context.Context, name domain.ProjectName) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.projects[name.String()]; !ok {
		return domain.ErrProjectNotFound
	}
	delete(s.projects, name.String())
	return nil
}

// FailWrites injects a write failure for project.
func (s *MemStore) FailWrites(project string, failure WriteFailure) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failures[project] = failure
}

// BeforeNextWrite runs hook once, right before the next write is applied.
func (s *MemStore) BeforeNextWrite(hook func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.beforeWrite = hook
}

func (s *MemStore) takeHook() func() {
	s.mu.Lock()
	defer s.mu.Unlock()
	hook := s.beforeWrite
	s.beforeWrite = nil
	return hook
}

// Seed stores each map as a successive version of project.
func (s *MemStore) Seed(project string, versions ...map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, raw := range versions {
		vars, err := domain.NewVariables(raw)
		if err != nil {
			panic(fmt.Sprintf("projecttest: invalid seed: %v", err))
		}
		s.projects[project] = append(s.projects[project], vars)
	}
}

// Latest returns the newest variables of project, or nil.
func (s *MemStore) Latest(project string) map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	versions := s.projects[project]
	if len(versions) == 0 {
		return nil
	}
	return versions[len(versions)-1].Map()
}

// Has reports whether project exists.
func (s *MemStore) Has(project string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.projects[project]
	return ok
}
