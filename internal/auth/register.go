package auth

import (
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
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	repository *repository.Repository
	db         *sql.DB
	jwtKey     []byte
}

func New(repo repository.Repository, jwtKey string) *Config {
	return &Config{
		repository: &repo,
		db:         repo.DB,
		jwtKey:     []byte(jwtKey),
	}
}

func (c *Config) Register(w http.ResponseWriter, r *http.Request) {
	// 1. verify method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Decode the incoming JSON request body into a User struct
	var newUser model.USER
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 3. validate request
	var validate = validator.New()
	err = validate.Struct(newUser)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(FormatValidationErrors(err))
		return
	}

	// 4. Hash password
	newUser.Password, err = Hash(newUser.Password)
	if err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	// store it in database
	err = c.repository.InsertUser(newUser)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1062 {

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)

				if strings.Contains(mysqlErr.Message, "username") {
					LogError("register_username_taken", "Username is already taken", nil)
					json.NewEncoder(w).Encode(map[string]string{"error": "Username is already taken"})
					return
				}

				if strings.Contains(mysqlErr.Message, "email") {
					LogError("register_email_taken", "Email is already taken", nil)
					json.NewEncoder(w).Encode(map[string]string{"error": "email is already taken"})
				}
				return
			}

			http.Error(w, "Failed to register user", http.StatusInternalServerError)
			LogError("register_db_error", "Failed to insert user into database", err)
			return
		}
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User created successfully",
	})

	LogSuccess("register_success", fmt.Sprintf("User %s registered successfully", newUser.UserName))
}

func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
