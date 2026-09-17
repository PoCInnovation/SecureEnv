package project

import (
	"context"

	"github.com/PoCInnovation/SecureEnv/internal/domain"
)

// Store persists versioned project variables. Implementations must return the
// domain sentinel errors (ErrProjectNotFound, ErrVersionConflict, ...) so the
// service stays storage agnostic.
type Store interface {
	// List returns every project name in lexical order.
	List(ctx context.Context) ([]domain.ProjectName, error)
	// Info returns the metadata of a project.
	Info(ctx context.Context, name domain.ProjectName) (domain.ProjectInfo, error)
	// Read returns the latest variables of a project.
	Read(ctx context.Context, name domain.ProjectName) (domain.Snapshot, error)
	// History returns every readable version of a project, oldest first.
	History(ctx context.Context, name domain.ProjectName) ([]domain.Snapshot, error)
	// Write stores vars as a new version only if the current version equals
	// expected (check-and-set). domain.NoVersion requires the project to not
	// exist. It returns the new version.
	Write(ctx context.Context, name domain.ProjectName, vars domain.Variables, expected domain.Version) (domain.Version, error)
	// Delete permanently removes a project and its whole history.
	Delete(ctx context.Context, name domain.ProjectName) error
}
