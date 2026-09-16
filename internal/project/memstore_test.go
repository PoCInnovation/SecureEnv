package project_test

import (
	"context"
	"maps"
	"slices"
	"sync"

	"github.com/PoCInnovation/SecureEnv/internal/domain"
)

// memStore is an in-memory project.Store that mimics Vault KV v2 versioning
// and check-and-set semantics.
type writeFailure struct {
	after int
	err   error
}

type memStore struct {
	mu       sync.Mutex
	projects map[string][]domain.Variables

	// failures makes Write fail for a project once it accepted `after` writes.
	failures map[string]writeFailure
	// beforeWrite runs once, right before the next write is applied.
	beforeWrite func()
}

func newMemStore() *memStore {
	return &memStore{projects: map[string][]domain.Variables{}, failures: map[string]writeFailure{}}
}

func (s *memStore) List(context.Context) ([]domain.ProjectName, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	names := make([]domain.ProjectName, 0, len(s.projects))
	for _, raw := range slices.Sorted(maps.Keys(s.projects)) {
		name, _ := domain.NewProjectName(raw)
		names = append(names, name)
	}
	return names, nil
}

func (s *memStore) Info(_ context.Context, name domain.ProjectName) (domain.ProjectInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	versions, ok := s.projects[name.String()]
	if !ok {
		return domain.ProjectInfo{}, domain.ErrProjectNotFound
	}
	return domain.ProjectInfo{Name: name, CurrentVersion: domain.Version(len(versions))}, nil
}

func (s *memStore) Read(_ context.Context, name domain.ProjectName) (domain.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	versions, ok := s.projects[name.String()]
	if !ok {
		return domain.Snapshot{}, domain.ErrProjectNotFound
	}
	return domain.Snapshot{Version: domain.Version(len(versions)), Variables: versions[len(versions)-1]}, nil
}

func (s *memStore) History(_ context.Context, name domain.ProjectName) ([]domain.Snapshot, error) {
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

func (s *memStore) Write(_ context.Context, name domain.ProjectName, vars domain.Variables, expected domain.Version) (domain.Version, error) {
	if hook := s.takeHook(); hook != nil {
		hook()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if failure, ok := s.failures[name.String()]; ok {
		if failure.after == 0 {
			return 0, failure.err
		}
		failure.after--
		s.failures[name.String()] = failure
	}
	versions := s.projects[name.String()]
	if domain.Version(len(versions)) != expected {
		return 0, domain.ErrVersionConflict
	}
	s.projects[name.String()] = append(versions, vars)
	return domain.Version(len(versions) + 1), nil
}

func (s *memStore) Delete(_ context.Context, name domain.ProjectName) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.projects[name.String()]; !ok {
		return domain.ErrProjectNotFound
	}
	delete(s.projects, name.String())
	return nil
}

func (s *memStore) takeHook() func() {
	s.mu.Lock()
	defer s.mu.Unlock()
	hook := s.beforeWrite
	s.beforeWrite = nil
	return hook
}

// seed stores each map as a successive version of project.
func (s *memStore) seed(project string, versions ...map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, raw := range versions {
		vars, err := domain.NewVariables(raw)
		if err != nil {
			panic(err)
		}
		s.projects[project] = append(s.projects[project], vars)
	}
}

func (s *memStore) latest(project string) map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	versions := s.projects[project]
	if len(versions) == 0 {
		return nil
	}
	return versions[len(versions)-1].Map()
}

func (s *memStore) has(project string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.projects[project]
	return ok
}
