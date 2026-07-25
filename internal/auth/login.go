package auth

import (
	"authd/internal/model"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func (c *Config) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loginReq model.LoginRequest

	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		LogError("login_request_decode", "Failed to decode login request", err)
		return
	}

	// Validate request
	var validate = validator.New()
	err = validate.Struct(loginReq)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(FormatValidationErrors(err))
		return
	}

	passwd, err := c.repository.GetUserCredentials(loginReq.UserName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			LogError("login_user_not_found", "User not found", err)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		LogError("login_db_error", "Failed to fetch user credentials", err)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwd), []byte(loginReq.Password))
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		LogError("login_invalid_password", "Invalid password provided", err)
		return
	}

	authToken, err := c.GenerateJWT(loginReq.UserName)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		LogError("login_jwt_generation", "Failed to generate JWT token", err)
		return
	}

	refreshToken, err := c.GenerateRefreshToken(loginReq.UserName)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		LogError("login_refresh_token_generation", "Failed to generate refresh token", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(model.LoginResponse{
		Token:        authToken,
		RefreshToken: refreshToken,
		Message:      "Login successful",
	})

	LogSuccess("login_success", fmt.Sprintf("User %s logged in successfully", loginReq.UserName))

}

func (c *Config) GenerateJWT(userID string) (string, error) {
	expirationTime := time.Now().Add(150 * time.Minute)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(c.jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (c *Config) GenerateRefreshToken(userID string) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(c.jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
