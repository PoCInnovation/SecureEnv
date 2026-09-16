package apiv1

import (
	"errors"
	"net/http"

	"github.com/PoCInnovation/SecureEnv/internal/domain"
)

// Code is a stable error identifier.
type Code string

// Error codes returned by the API.
const (
	CodeInvalidRequest      Code = "invalid_request"
	CodeInvalidProjectName  Code = "invalid_project_name"
	CodeInvalidVariableKey  Code = "invalid_variable_key"
	CodeReservedVariableKey Code = "reserved_variable_key"
	CodeProjectNotFound     Code = "project_not_found"
	CodeProjectExists       Code = "project_exists"
	CodeVariableNotFound    Code = "variable_not_found"
	CodeVersionConflict     Code = "version_conflict"
	CodeUnauthorized        Code = "unauthorized"
	CodeForbidden           Code = "forbidden"
	CodeInternal            Code = "internal"
)

type mapping struct {
	err    error
	code   Code
	status int
}

// mappings is the single source of truth between domain errors, codes and
// HTTP statuses. Order matters: an error may wrap several sentinels (a failed
// create wraps both ErrProjectExists and ErrVersionConflict) and the first
// match wins.
var mappings = []mapping{
	{domain.ErrInvalidProjectName, CodeInvalidProjectName, http.StatusBadRequest},
	{domain.ErrInvalidVariableKey, CodeInvalidVariableKey, http.StatusBadRequest},
	{domain.ErrReservedVariableKey, CodeReservedVariableKey, http.StatusBadRequest},
	{domain.ErrProjectNotFound, CodeProjectNotFound, http.StatusNotFound},
	{domain.ErrProjectExists, CodeProjectExists, http.StatusConflict},
	{domain.ErrVariableNotFound, CodeVariableNotFound, http.StatusNotFound},
	{domain.ErrVersionConflict, CodeVersionConflict, http.StatusPreconditionFailed},
	{domain.ErrUnauthorized, CodeUnauthorized, http.StatusUnauthorized},
	{domain.ErrForbidden, CodeForbidden, http.StatusForbidden},
}

// FromError returns the code and HTTP status describing err. Unknown errors
// are internal.
func FromError(err error) (Code, int) {
	for _, m := range mappings {
		if errors.Is(err, m.err) {
			return m.code, m.status
		}
	}
	return CodeInternal, http.StatusInternalServerError
}

// ToError returns the domain error matching code, or nil when there is none.
func ToError(code Code) error {
	for _, m := range mappings {
		if m.code == code {
			return m.err
		}
	}
	return nil
}
