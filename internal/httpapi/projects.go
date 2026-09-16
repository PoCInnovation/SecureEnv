package httpapi

import (
	"net/http"
	"net/url"

	"github.com/PoCInnovation/SecureEnv/internal/apiv1"
)

func (s *server) listProjects(w http.ResponseWriter, r *http.Request, svc ProjectService) error {
	names, err := svc.List(r.Context())
	if err != nil {
		return err
	}
	body := apiv1.ProjectList{Projects: make([]string, len(names))}
	for i, name := range names {
		body.Projects[i] = name.String()
	}
	writeJSON(w, http.StatusOK, body)
	return nil
}

func (s *server) createProject(w http.ResponseWriter, r *http.Request, svc ProjectService) error {
	var req apiv1.ProjectNameRequest
	if err := decodeJSON(r, &req); err != nil {
		return err
	}
	if err := svc.Create(r.Context(), req.Name); err != nil {
		return err
	}
	w.Header().Set("Location", apiv1.BasePath+"/projects/"+url.PathEscape(req.Name))
	w.WriteHeader(http.StatusCreated)
	return nil
}

func (s *server) projectInfo(w http.ResponseWriter, r *http.Request, svc ProjectService) error {
	info, err := svc.Info(r.Context(), r.PathValue("project"))
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, apiv1.Project{
		Name:           info.Name.String(),
		CurrentVersion: int(info.CurrentVersion),
		CreatedAt:      info.CreatedAt,
		UpdatedAt:      info.UpdatedAt,
	})
	return nil
}

func (s *server) renameProject(w http.ResponseWriter, r *http.Request, svc ProjectService) error {
	var req apiv1.ProjectNameRequest
	if err := decodeJSON(r, &req); err != nil {
		return err
	}
	if err := svc.Rename(r.Context(), r.PathValue("project"), req.Name); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *server) deleteProject(w http.ResponseWriter, r *http.Request, svc ProjectService) error {
	if err := svc.Delete(r.Context(), r.PathValue("project")); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
