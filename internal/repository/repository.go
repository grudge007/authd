package repository

import (
	"authd/internal/model"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type ProjectMember struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type ProjectInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Repository struct {
	DB *sql.DB
}

func InitDB(dsn string) (*Repository, error) {
	// 1. Open database connection handle
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db handle: %w", err)
	}

	// 2. Set Connection Pool Configuration
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 3. Check DB Alive
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Repository{
		DB: db,
	}, nil
}

func (d *Repository) InsertUser(newUser model.USER) error {
	query := `INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)`
	_, err := d.DB.Exec(query, newUser.UserName, newUser.Email, newUser.Password)
	if err != nil {
		return err
	}

	return nil
}

func (d *Repository) GetUserCredentials(username string) (string, error) {
	var passwd string

	query := `SELECT password_hash FROM users WHERE username = ?`

	err := d.DB.QueryRow(query, username).Scan(&passwd)
	if err != nil {
		return "", err
	}

	return passwd, nil
}

func (d *Repository) UpdatePassword(username, password string) error {
	query := "UPDATE users SET password_hash = ? WHERE username = ?"

	_, err := d.DB.Exec(query, password, username)
	if err != nil {
		return err
	}

	return nil
}

func (d *Repository) ValidateProjectExistance(name, owner string) error {
	query := `SELECT name FROM projects WHERE name = ? AND owner_id = ?`
	var projectName string
	err := d.DB.QueryRow(query, name, owner).Scan(&projectName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil // Project doesn't exist, can create
		}
		return err // Database error
	}

	return fmt.Errorf("project with name %s already exists for owner %s", projectName, owner)
}

func (d *Repository) InsertProject(new model.CreateProject, owner string) error {
	query := `INSERT INTO projects (name, description, owner_id) VALUES (?, ?, ?)`
	// owner_id is foreign key (id from  users)

	_, err := d.DB.Exec(query, new.ProjectName, new.ProjectDecs, owner)
	if err != nil {
		return err
	}

	return nil
}

func (d *Repository) FetchProjectId(projectName, user string) (string, error) {
	var projectId string
	query := `SELECT id FROM projects WHERE name = ? AND owner_id = ?`

	row := d.DB.QueryRow(query, projectName, user)

	if err := row.Scan(&projectId); err != nil {
		return "", err
	}

	return projectId, nil
}

func (d *Repository) InsertProjectMember(projectId, user, role string) error {
	query := `INSERT INTO project_members (project_id, user_id, role) VALUES (?, ?, ?)`
	_, err := d.DB.Exec(query, projectId, user, role)
	if err != nil {
		return err
	}

	return nil
}

func (d *Repository) UserExists(username string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE username = ?`
	err := d.DB.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *Repository) FetchProjectName(projectID, owner string) (string, error) {
	var projectName string
	query := `SELECT name FROM projects WHERE id = ? AND owner_id = ?`
	err := d.DB.QueryRow(query, projectID, owner).Scan(&projectName)
	if err != nil {
		return "", err
	}
	return projectName, nil
}

func (d *Repository) ListUsers(excludeUser string) ([]string, error) {
	query := `SELECT username FROM users WHERE username != ?`
	rows, err := d.DB.Query(query, excludeUser)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, err
		}
		users = append(users, username)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (d *Repository) ListProjectMembers(projectID string) ([]ProjectMember, error) {
	query := `SELECT user_id, role FROM project_members WHERE project_id = ?`
	rows, err := d.DB.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []ProjectMember
	for rows.Next() {
		var member ProjectMember
		if err := rows.Scan(&member.UserID, &member.Role); err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

func (d *Repository) ListProjects(owner string) ([]ProjectInfo, error) {
	query := `SELECT id, name, description FROM projects WHERE owner_id = ?`
	rows, err := d.DB.Query(query, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []ProjectInfo
	for rows.Next() {
		var projectID, projectName, projectDesc string
		if err := rows.Scan(&projectID, &projectName, &projectDesc); err != nil {
			return nil, err
		}
		projects = append(projects, ProjectInfo{
			ID:          projectID,
			Name:        projectName,
			Description: projectDesc,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}

func (d *Repository) RemoveProjectMember(projectID, userID string) error {
	query := `DELETE FROM project_members WHERE project_id = ? AND user_id = ?`
	_, err := d.DB.Exec(query, projectID, userID)
	if err != nil {
		return err
	}
	return nil
}

func (d *Repository) DeleteUser(username string) error {
	// First delete from project_members
	query1 := `DELETE FROM project_members WHERE user_id = ?`
	_, err := d.DB.Exec(query1, username)
	if err != nil {
		return err
	}

	// Then delete from projects (as owner)
	query2 := `DELETE FROM projects WHERE owner_id = ?`
	_, err = d.DB.Exec(query2, username)
	if err != nil {
		return err
	}

	// Finally delete from users
	query3 := `DELETE FROM users WHERE username = ?`
	_, err = d.DB.Exec(query3, username)
	if err != nil {
		return err
	}

	return nil
}

func (d *Repository) HasProjectAccess(projectID, userID string) (bool, error) {
	var role string
	query := `SELECT role FROM project_members WHERE project_id = ? AND user_id = ?`
	err := d.DB.QueryRow(query, projectID, userID).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // User not in project
		}
		return false, err
	}

	// Allow access if user is owner or admin
	return role == "owner" || role == "admin", nil
}

func (d *Repository) GetProjectIdByNameAndUser(projectName, userID string) (string, error) {
	var projectID string
	// First try as owner
	query := `SELECT id FROM projects WHERE name = ? AND owner_id = ?`
	err := d.DB.QueryRow(query, projectName, userID).Scan(&projectID)
	if err == nil {
		return projectID, nil
	}

	// Try to find project by name and check if user has admin/owner access
	query = `SELECT p.id FROM projects p JOIN project_members pm ON p.id = pm.project_id WHERE p.name = ? AND pm.user_id = ? AND (pm.role = 'owner' OR pm.role = 'admin')`
	err = d.DB.QueryRow(query, projectName, userID).Scan(&projectID)
	if err == nil {
		return projectID, nil
	}

	return "", sql.ErrNoRows
}

func (d *Repository) DeleteProject(projectID, userID string) error {
	// Verify user owns the project
	var count int
	query := `SELECT COUNT(*) FROM projects WHERE id = ? AND owner_id = ?`
	err := d.DB.QueryRow(query, projectID, userID).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("project not found or access denied")
	}

	// Delete all project members first (cascade)
	query = `DELETE FROM project_members WHERE project_id = ?`
	_, err = d.DB.Exec(query, projectID)
	if err != nil {
		return err
	}

	// Then delete the project
	query = `DELETE FROM projects WHERE id = ?`
	_, err = d.DB.Exec(query, projectID)
	if err != nil {
		return err
	}

	return nil
}
