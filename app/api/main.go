package main

import (
	"authd/internal/auth"
	"authd/internal/projects"
	"authd/internal/repository"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	jwtKey := os.Getenv("JWT_KEY")
	dsn := os.Getenv("DSN")

	repo, err := repository.InitDB(dsn)
	if err != nil {
		log.Fatalf("Error opening DB: %v", err)
	}

	defer repo.DB.Close()

	authHandler := auth.New(*repo, jwtKey)
	projectHandler := projects.NewProject(jwtKey, *repo)

	// Register
	http.HandleFunc("/api/v1/authd/register", authHandler.Register)

	// Login
	http.HandleFunc("/api/v1/authd/login", authHandler.Login)

	// Change Password
	http.HandleFunc("/api/v1/authd/user/update", authHandler.Update)

	// Refresh Token
	http.HandleFunc("/api/v1/authd/refresh", authHandler.Refresh)

	// Create Project
	http.HandleFunc("/api/v1/authd/projects/create", projectHandler.CreateProject)

	// Add member
	http.HandleFunc("/api/v1/authd/projects/members/add", projectHandler.AddMembers)

	// Remove member
	http.HandleFunc("/api/v1/authd/projects/members/remove", projectHandler.RemoveMember)

	// List global users
	http.HandleFunc("/api/v1/authd/projects/users/list", projectHandler.ListGlobal)

	// List project members
	http.HandleFunc("/api/v1/authd/projects/members/list", projectHandler.ListByProjectId)

	// List projects
	http.HandleFunc("/api/v1/authd/projects/list", projectHandler.ListProjects)

	// Delete project
	http.HandleFunc("/api/v1/authd/projects/delete", projectHandler.DeleteProject)

	// Delete account
	http.HandleFunc("/api/v1/authd/user/delete", auth.NewDeleteHandler(repo, jwtKey))

	log.Println("Server running on http://localhost:7070...")
	err = http.ListenAndServe(":7070", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
