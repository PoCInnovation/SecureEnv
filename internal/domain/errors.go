package domain

import "errors"

// Sentinel errors shared by every layer. Adapters wrap them so callers can use
// errors.Is regardless of where the failure happened.
var (
	ErrInvalidProjectName  = errors.New("invalid project name")
	ErrInvalidVariableKey  = errors.New("invalid variable key")
	ErrReservedVariableKey = errors.New("reserved variable key")
	ErrProjectNotFound     = errors.New("project not found")
	ErrProjectExists       = errors.New("project already exists")
	ErrVariableNotFound    = errors.New("variable not found")
	ErrVersionConflict     = errors.New("project was modified concurrently")
	ErrUnauthorized        = errors.New("missing or invalid credentials")
	ErrForbidden           = errors.New("permission denied")
)
