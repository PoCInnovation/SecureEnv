// Package vaultstore implements project.Store on the HashiCorp Vault KV v2
// secrets engine.
//
// Each project is one secret at <mount>/data/<prefix>/<project>; its
// variables are the key/value pairs of that secret. Every write goes through
// check-and-set so Vault itself arbitrates concurrent updates.
package vaultstore

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	vault "github.com/hashicorp/vault/api"

	"github.com/PoCInnovation/SecureEnv/internal/domain"
)

// Options selects where projects live in Vault.
type Options struct {
	// Mount is the KV v2 mount path. Defaults to "secret".
	Mount string
	// Prefix is an optional folder under the mount holding every project.
	Prefix string
}

// Store is a project.Store backed by Vault.
type Store struct {
	client *vault.Client
	mount  string
	prefix string
}

// New returns a Store using client. The client token, if any, is used for
// every call; see WithToken to act on behalf of a caller.
func New(client *vault.Client, opts Options) *Store {
	mount := strings.Trim(opts.Mount, "/")
	if mount == "" {
		mount = "secret"
	}
	return &Store{client: client, mount: mount, prefix: strings.Trim(opts.Prefix, "/")}
}

// WithToken returns a copy of the store that authenticates with token. The
// receiver is left untouched, so it is safe to call concurrently.
func (s *Store) WithToken(token string) (*Store, error) {
	client, err := s.client.Clone()
	if err != nil {
		return nil, fmt.Errorf("vault: clone client: %w", err)
	}
	client.SetToken(token)
	return &Store{client: client, mount: s.mount, prefix: s.prefix}, nil
}

// Ping reports whether Vault is reachable, initialized and unsealed.
func (s *Store) Ping(ctx context.Context) error {
	health, err := s.client.Sys().HealthWithContext(ctx)
	if err != nil {
		return fmt.Errorf("vault: health: %w", err)
	}
	switch {
	case !health.Initialized:
		return errors.New("vault: not initialized")
	case health.Sealed:
		return errors.New("vault: sealed")
	}
	return nil
}

// List implements project.Store.
func (s *Store) List(ctx context.Context) ([]domain.ProjectName, error) {
	secret, err := s.client.Logical().ListWithContext(ctx, path.Join(s.mount, "metadata", s.prefix)+"/")
	if err != nil {
		return nil, mapError(err)
	}
	if secret == nil || secret.Data == nil {
		return []domain.ProjectName{}, nil
	}

	keys, _ := secret.Data["keys"].([]any)
	names := make([]domain.ProjectName, 0, len(keys))
	for _, key := range keys {
		raw, _ := key.(string)
		// Folders end with "/" and foreign secrets may not be valid names.
		if name, err := domain.NewProjectName(raw); err == nil {
			names = append(names, name)
		}
	}
	return names, nil
}

// Info implements project.Store.
func (s *Store) Info(ctx context.Context, name domain.ProjectName) (domain.ProjectInfo, error) {
	metadata, err := s.kv().GetMetadata(ctx, s.secretPath(name))
	if err != nil {
		return domain.ProjectInfo{}, mapError(err)
	}
	return domain.ProjectInfo{
		Name:           name,
		CurrentVersion: domain.Version(metadata.CurrentVersion),
		CreatedAt:      metadata.CreatedTime,
		UpdatedAt:      metadata.UpdatedTime,
	}, nil
}

// Read implements project.Store. A soft-deleted latest version reads as an
// empty set of variables.
func (s *Store) Read(ctx context.Context, name domain.ProjectName) (domain.Snapshot, error) {
	secret, err := s.kv().Get(ctx, s.secretPath(name))
	if err != nil {
		return domain.Snapshot{}, mapError(err)
	}
	return toSnapshot(secret)
}

// History implements project.Store. Deleted and destroyed versions are skipped.
func (s *Store) History(ctx context.Context, name domain.ProjectName) ([]domain.Snapshot, error) {
	versions, err := s.kv().GetVersionsAsList(ctx, s.secretPath(name))
	if err != nil {
		return nil, mapError(err)
	}

	history := make([]domain.Snapshot, 0, len(versions))
	for _, version := range versions {
		if version.Destroyed || !version.DeletionTime.IsZero() {
			continue
		}
		secret, err := s.kv().GetVersion(ctx, s.secretPath(name), version.Version)
		if err != nil {
			return nil, mapError(err)
		}
		snapshot, err := toSnapshot(secret)
		if err != nil {
			return nil, err
		}
		history = append(history, snapshot)
	}
	return history, nil
}

// Write implements project.Store.
func (s *Store) Write(ctx context.Context, name domain.ProjectName, vars domain.Variables, expected domain.Version) (domain.Version, error) {
	data := make(map[string]any, vars.Len())
	for key, value := range vars.Map() {
		data[key] = value
	}

	secret, err := s.kv().Put(ctx, s.secretPath(name), data, vault.WithCheckAndSet(int(expected)))
	if err != nil {
		return 0, mapError(err)
	}
	return domain.Version(secret.VersionMetadata.Version), nil
}

// Delete implements project.Store.
func (s *Store) Delete(ctx context.Context, name domain.ProjectName) error {
	// Vault answers 204 when deleting unknown metadata; check first so
	// callers get a meaningful error.
	if _, err := s.kv().GetMetadata(ctx, s.secretPath(name)); err != nil {
		return mapError(err)
	}
	return mapError(s.kv().DeleteMetadata(ctx, s.secretPath(name)))
}

func (s *Store) kv() *vault.KVv2 { return s.client.KVv2(s.mount) }

func (s *Store) secretPath(name domain.ProjectName) string {
	return path.Join(s.prefix, name.String())
}

func toSnapshot(secret *vault.KVSecret) (domain.Snapshot, error) {
	raw := make(map[string]string, len(secret.Data))
	for key, value := range secret.Data {
		str, ok := value.(string)
		if !ok {
			return domain.Snapshot{}, fmt.Errorf("vault: variable %q is a %T, only strings are supported", key, value)
		}
		raw[key] = str
	}

	vars, err := domain.NewVariables(raw)
	if err != nil {
		return domain.Snapshot{}, fmt.Errorf("vault: stored data is not a valid project: %w", err)
	}

	var version domain.Version
	if secret.VersionMetadata != nil {
		version = domain.Version(secret.VersionMetadata.Version)
	}
	return domain.Snapshot{Version: version, Variables: vars}, nil
}
