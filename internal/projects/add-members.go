package projects

import (
	"authd/internal/auth"
	"authd/internal/model"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

func (p *Project) AddMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, err := auth.ValidateAuth(r.Header, string(p.jwtKey))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		auth.LogError("add_members_auth", "Invalid authentication", err)
		return
	}

	var addMembersReq model.AddMembers
	if err := json.NewDecoder(r.Body).Decode(&addMembersReq); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		auth.LogError("add_members_decode", "Failed to decode add members request", err)
		return
	}

	// Validate request
	var validate = validator.New()
	err = validate.Struct(addMembersReq)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(auth.FormatValidationErrors(err))
		return
	}

	// Get project ID - allows owner or admin/access users
	projectID, err := p.repository.GetProjectIdByNameAndUser(addMembersReq.ProjectName, claims.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "Access denied. Only project owner or admin can add members."})
			auth.LogError("add_members_access_denied", "Access denied - not owner or admin", nil)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch project"})
		auth.LogError("add_members_fetch_project", "Failed to fetch project ID", err)
		return
	}

	// Validate that all members exist before adding any
	var invalidMembers []string
	for _, member := range addMembersReq.Members {
		exists, err := p.repository.UserExists(member)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to validate member"})
			auth.LogError("add_members_validate", "Failed to check if user exists", err)
			return
		}
		if !exists {
			invalidMembers = append(invalidMembers, member)
		}
	}

	if len(invalidMembers) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Invalid members: %s", strings.Join(invalidMembers, ", ")),
		})
		auth.LogError("add_members_invalid", "Invalid members provided", nil)
		return
	}

	// Add each member to the project
	for _, member := range addMembersReq.Members {
		err = p.repository.InsertProjectMember(projectID, member, "member")
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to add member"})
			auth.LogError("add_members_insert", "Failed to insert project member", err)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("%d members added successfully", len(addMembersReq.Members)),
	})

	auth.LogSuccess("add_members_success", fmt.Sprintf("Added %d members to project %s", len(addMembersReq.Members), addMembersReq.ProjectName))
}
