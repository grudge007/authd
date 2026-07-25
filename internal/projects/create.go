package projects

import (
	"authd/internal/auth"
	"authd/internal/model"
	"authd/internal/repository"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/go-sql-driver/mysql"
)

type Project struct {
	repository *repository.Repository
	db         *sql.DB
	jwtKey     []byte
}

func NewProject(jwtKey string, repo repository.Repository) *Project {
	return &Project{
		repository: &repo,
		db:         repo.DB,
		jwtKey:     []byte(jwtKey),
	}
}

func (p *Project) CreateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, err := auth.ValidateAuth(r.Header, string(p.jwtKey))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		auth.LogError("create_project_auth", "Invalid authentication", err)
		return
	}

	var newProject model.CreateProject
	if err := json.NewDecoder(r.Body).Decode(&newProject); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		auth.LogError("create_project_decode", "Failed to decode project data", err)
		return
	}

	// Validate request
	var validate = validator.New()
	err = validate.Struct(newProject)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(auth.FormatValidationErrors(err))
		return
	}

	// Check if project already exists for this owner
	if err := p.repository.ValidateProjectExistance(newProject.ProjectName, claims.UserID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		auth.LogError("create_project_exists", "Project already exists", err)
		return
	}

	if err := p.repository.InsertProject(newProject, claims.UserID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		// Sanitize database error for response
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1062 {
				if strings.Contains(mysqlErr.Message, "name") {
					json.NewEncoder(w).Encode(map[string]string{"error": "Project name already exists"})
					auth.LogError("create_project_exists", "Project name already exists", nil)
					return
				}
			}
		}

		// Generic error message for all other DB errors
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create project"})
		auth.LogError("create_project_db", "Failed to insert project into database", err)
		return
	}

	// Add the creator as owner in project_members
	projectID, err := p.repository.FetchProjectId(newProject.ProjectName, claims.UserID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to add owner to project members"})
		auth.LogError("create_project_add_owner", "Failed to add owner to project members", err)
		return
	}

	if err := p.repository.InsertProjectMember(projectID, claims.UserID, "owner"); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to add owner to project members"})
		auth.LogError("create_project_add_owner", "Failed to add owner to project members", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Project created successfully",
	})

	auth.LogSuccess("create_project_success", fmt.Sprintf("Project %s created by user %s", newProject.ProjectName, claims.UserID))
}
