package projects

import (
	"authd/internal/auth"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

type RemoveMemberRequest struct {
	ProjectName string `json:"project_name"`
	UserID      string `json:"user_id"`
}

func (p *Project) RemoveMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, err := auth.ValidateAuth(r.Header, string(p.jwtKey))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		auth.LogError("remove_member_auth", "Invalid authentication", err)
		return
	}

	var req RemoveMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		auth.LogError("remove_member_decode", "Failed to decode request", err)
		return
	}

	// Validate required fields
	if req.ProjectName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "project_name is required"})
		auth.LogError("remove_member_missing_project", "Missing project_name", nil)
		return
	}

	if req.UserID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "user_id is required"})
		auth.LogError("remove_member_missing_user", "Missing user_id", nil)
		return
	}

	// Check if current user owns the project
	projectID, err := p.repository.FetchProjectId(req.ProjectName, claims.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Project not found or access denied"})
			auth.LogError("remove_member_project_not_found", "Project not found or access denied", nil)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch project"})
		auth.LogError("remove_member_fetch_project", "Failed to fetch project", err)
		return
	}

	// Prevent owner from removing themselves
	if req.UserID == claims.UserID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Cannot remove yourself from the project"})
		auth.LogError("remove_member_self", "Owner cannot remove themselves", nil)
		return
	}

	// Remove the member from the project
	if err := p.repository.RemoveProjectMember(projectID, req.UserID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to remove member"})
		auth.LogError("remove_member_delete", "Failed to remove member from project", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("User %s removed from project %s", req.UserID, req.ProjectName),
	})

	auth.LogSuccess("remove_member_success", fmt.Sprintf("Removed user %s from project %s", req.UserID, req.ProjectName))
}
