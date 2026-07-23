package repository

import (
	"authd/internal/model"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

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

func (d *Repository) GetUserCredentials(username string) (string, string, error) {
	var id, passwd string

	query := `SELECT id, password_hash FROM users WHERE username = ?`

	err := d.DB.QueryRow(query, username).Scan(&id, &passwd)
	if err != nil {
		return "", "", err
	}

	return id, passwd, nil
}
