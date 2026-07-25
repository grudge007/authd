package projects

import (
	"authd/internal/auth"
	"authd/internal/repository"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

type ListUsersResponse struct {
	Users []string `json:"users"`
}

type ProjectMembersResponse struct {
	ProjectName string                     `json:"project_name"`
	Members     []repository.ProjectMember `json:"members"`
}

type ListProjectsResponse struct {
	Projects []repository.ProjectInfo `json:"projects"`
}

func (p *Project) ListGlobal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, err := auth.ValidateAuth(r.Header, string(p.jwtKey))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		auth.LogError("list_global_auth", "Invalid authentication", err)
		return
	}

	// Fetch all usernames from users table (excluding current user)
	users, err := p.repository.ListUsers(claims.UserID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch users"})
		auth.LogError("list_global_fetch_users", "Failed to fetch users", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ListUsersResponse{Users: users})

	auth.LogSuccess("list_global_success", fmt.Sprintf("Listed global users for user %s", claims.UserID))
}

func (p *Project) ListByProjectId(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, err := auth.ValidateAuth(r.Header, string(p.jwtKey))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		auth.LogError("list_by_project_auth", "Invalid authentication", err)
		return
	}

	// Get project ID from header
	projectID := r.Header.Get("X-Project-Id")
	if projectID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "X-Project-Id header is required"})
		auth.LogError("list_by_project_missing_header", "Missing X-Project-Id header", nil)
		return
	}

	// Verify user owns this project
	projectName, err := p.repository.FetchProjectName(projectID, claims.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Project not found or access denied"})
			auth.LogError("list_by_project_not_found", "Project not found or access denied", nil)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch project"})
		auth.LogError("list_by_project_fetch", "Failed to fetch project", err)
		return
	}

	// Fetch members for this project
	members, err := p.repository.ListProjectMembers(projectID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch project members"})
		auth.LogError("list_by_project_fetch_members", "Failed to fetch project members", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ProjectMembersResponse{
		ProjectName: projectName,
		Members:     members,
	})

	auth.LogSuccess("list_by_project_success", fmt.Sprintf("Listed members for project %s", projectName))
}

func (p *Project) ListProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, err := auth.ValidateAuth(r.Header, string(p.jwtKey))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		auth.LogError("list_projects_auth", "Invalid authentication", err)
		return
	}

	// Fetch all projects owned by the user
	projects, err := p.repository.ListProjects(claims.UserID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch projects"})
		auth.LogError("list_projects_fetch", "Failed to fetch projects", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ListProjectsResponse{Projects: projects})

	auth.LogSuccess("list_projects_success", fmt.Sprintf("Listed projects for user %s", claims.UserID))
}
