package projects

import (
	"authd/internal/auth"
	"encoding/json"
	"fmt"
	"net/http"
)

type DeleteProjectRequest struct {
	ProjectName string `json:"project_name"`
}

func (p *Project) DeleteProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, err := auth.ValidateAuth(r.Header, string(p.jwtKey))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		auth.LogError("delete_project_auth", "Invalid authentication", err)
		return
	}

	var req DeleteProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		auth.LogError("delete_project_decode", "Failed to decode request", err)
		return
	}

	if req.ProjectName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "project_name is required"})
		auth.LogError("delete_project_missing_name", "Missing project_name", nil)
		return
	}

	// Get project ID
	projectID, err := p.repository.FetchProjectId(req.ProjectName, claims.UserID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Project not found or access denied"})
		auth.LogError("delete_project_not_found", "Project not found or access denied", nil)
		return
	}

	// Delete project (this also deletes all project_members entries)
	if err := p.repository.DeleteProject(projectID, claims.UserID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete project"})
		auth.LogError("delete_project_delete", "Failed to delete project", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Project %s and all associated members deleted successfully", req.ProjectName),
	})

	auth.LogSuccess("delete_project_success", fmt.Sprintf("Deleted project %s", req.ProjectName))
}
