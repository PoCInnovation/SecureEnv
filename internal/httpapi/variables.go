package httpapi

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/PoCInnovation/SecureEnv/internal/apiv1"
	"github.com/PoCInnovation/SecureEnv/internal/domain"
)

func (s *server) getVariables(w http.ResponseWriter, r *http.Request, svc ProjectService) error {
	snapshot, err := svc.Variables(r.Context(), r.PathValue("project"))
	if err != nil {
		return err
	}
	setETag(w, snapshot.Version)
	writeJSON(w, http.StatusOK, apiv1.Variables{
		Version:   int(snapshot.Version),
		Variables: snapshot.Variables.Map(),
	})
	return nil
}

func (s *server) replaceVariables(w http.ResponseWriter, r *http.Request, svc ProjectService) error {
	expected, err := ifMatchVersion(r)
	if err != nil {
		return err
	}
	var req apiv1.ReplaceVariablesRequest
	if err := decodeJSON(r, &req); err != nil {
		return err
	}
	if req.Variables == nil {
		return fmt.Errorf("%w: variables is required", errInvalidRequest)
	}

	version, err := svc.ReplaceVariables(r.Context(), r.PathValue("project"), req.Variables, expected)
	if err != nil {
		return err
	}
	return writeVersion(w, version)
}

func (s *server) getVariable(w http.ResponseWriter, r *http.Request, svc ProjectService) error {
	key := r.PathValue("key")
	value, err := svc.Variable(r.Context(), r.PathValue("project"), key)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, apiv1.Variable{Key: key, Value: value})
	return nil
}

func (s *server) setVariable(w http.ResponseWriter, r *http.Request, svc ProjectService) error {
	var req apiv1.SetVariableRequest
	if err := decodeJSON(r, &req); err != nil {
		return err
	}
	version, err := svc.SetVariable(r.Context(), r.PathValue("project"), r.PathValue("key"), req.Value)
	if err != nil {
		return err
	}
	return writeVersion(w, version)
}

func (s *server) deleteVariable(w http.ResponseWriter, r *http.Request, svc ProjectService) error {
	version, err := svc.DeleteVariable(r.Context(), r.PathValue("project"), r.PathValue("key"))
	if err != nil {
		return err
	}
	return writeVersion(w, version)
}

func writeVersion(w http.ResponseWriter, version domain.Version) error {
	setETag(w, version)
	writeJSON(w, http.StatusOK, apiv1.VersionResponse{Version: int(version)})
	return nil
}

func setETag(w http.ResponseWriter, version domain.Version) {
	w.Header().Set("ETag", strconv.Quote(strconv.Itoa(int(version))))
}

// ifMatchVersion reads the optional If-Match header holding a strong ETag
// produced by setETag.
func ifMatchVersion(r *http.Request) (domain.Version, error) {
	header := r.Header.Get("If-Match")
	if header == "" {
		return domain.AnyVersion, nil
	}
	unquoted, err := strconv.Unquote(header)
	if err == nil {
		var version int
		if version, err = strconv.Atoi(unquoted); err == nil && version >= 0 {
			return domain.Version(version), nil
		}
	}
	return 0, fmt.Errorf("%w: If-Match must be a version ETag such as \"3\"", errInvalidRequest)
}
