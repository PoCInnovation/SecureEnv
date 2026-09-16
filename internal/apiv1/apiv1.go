// Package apiv1 defines the wire contract of the SecureEnv HTTP API v1. It is
// shared by the server and the client so both sides always agree.
package apiv1

import "time"

// BasePath prefixes every versioned route.
const BasePath = "/v1"

// ProjectList is returned by GET /v1/projects.
type ProjectList struct {
	Projects []string `json:"projects"`
}

// ProjectNameRequest is the body of POST /v1/projects and
// PATCH /v1/projects/{project}.
type ProjectNameRequest struct {
	Name string `json:"name"`
}

// Project is returned by GET /v1/projects/{project}.
type Project struct {
	Name           string    `json:"name"`
	CurrentVersion int       `json:"current_version"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Variables is returned by GET /v1/projects/{project}/variables.
type Variables struct {
	Version   int               `json:"version"`
	Variables map[string]string `json:"variables"`
}

// ReplaceVariablesRequest is the body of PUT /v1/projects/{project}/variables.
// Send the version you based your changes on in the If-Match header to
// reject concurrent updates.
type ReplaceVariablesRequest struct {
	Variables map[string]string `json:"variables"`
}

// Variable is returned by GET /v1/projects/{project}/variables/{key}.
type Variable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// SetVariableRequest is the body of PUT /v1/projects/{project}/variables/{key}.
type SetVariableRequest struct {
	Value string `json:"value"`
}

// VersionResponse is returned by every write on variables.
type VersionResponse struct {
	Version int `json:"version"`
}

// ErrorResponse is the body of every non 2xx response.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody describes a failure with a stable, machine readable code.
type ErrorBody struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
}
