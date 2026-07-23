package main

import (
	"authd/internal/auth"
	"authd/internal/repository"
	"log"
	"net/http"
)

func main() {
	dsn := "app-admin:passwd@tcp(127.0.0.1:3306)/authd?parseTime=true"

	repo, err := repository.InitDB(dsn)
	if err != nil {
		log.Fatalf("Error opening DB: %v", err)
	}

	defer repo.DB.Close()

	authHandler := auth.New(*repo)

	// Register
	http.HandleFunc("/api/v1/authd/register", authHandler.Register)

	// Login
	http.HandleFunc("/api/v1/authd/login", authHandler.Login)

	log.Println("Server running on http://localhost:7070...")
	err = http.ListenAndServe(":7070", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
