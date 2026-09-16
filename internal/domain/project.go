package domain

import (
	"fmt"
	"regexp"
	"time"
)

var projectNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

// ProjectName is a validated project identifier. It is safe to use as a single
// Vault path segment.
type ProjectName struct{ value string }

// NewProjectName validates raw and returns a ProjectName.
func NewProjectName(raw string) (ProjectName, error) {
	if !projectNamePattern.MatchString(raw) {
		return ProjectName{}, fmt.Errorf("%w: %q must match %s", ErrInvalidProjectName, raw, projectNamePattern)
	}
	return ProjectName{value: raw}, nil
}

func (n ProjectName) String() string { return n.value }

// Version identifies a revision of a project's variables. Versions start at 1.
type Version int

// NoVersion is used for check-and-set writes that require the project to not
// exist yet.
const NoVersion Version = 0

// Snapshot is the set of variables of a project at a given version.
type Snapshot struct {
	Version   Version
	Variables Variables
}

// ProjectInfo describes a project without exposing its secret values.
type ProjectInfo struct {
	Name           ProjectName
	CurrentVersion Version
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
