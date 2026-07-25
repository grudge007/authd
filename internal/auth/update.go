package auth

import (
	"authd/internal/model"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const bearer = "Bearer "

func (c *Config) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, bearer) {
		http.Error(w, "Missing bearer token", http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, bearer)

	token, err := jwt.ParseWithClaims(tokenString,
		&Claims{},
		func(t *jwt.Token) (any, error) {
			return c.jwtKey, nil
		},
	)

	if err != nil || token == nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		LogError("update_invalid_token", "Invalid or expired JWT token", err)
		return
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	var req model.ChangePassword
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		LogError("update_request_decode", "Failed to decode change password request", err)
		return
	}

	// Validate request
	var validate = validator.New()
	err = validate.Struct(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(formatValidationErrors(err))
		return
	}

	passwdHash, err := c.repository.GetUserCredentials(claims.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Invalid user", http.StatusUnauthorized)
			LogError("update_user_not_found", "User not found", err)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		LogError("update_db_error", "Failed to fetch user credentials", err)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(passwdHash), []byte(req.OldPassword)); err != nil {
		http.Error(w, "Invalid Password", http.StatusUnauthorized)
		LogError("update_invalid_old_password", "Invalid old password provided", err)
		return
	}

	newPasswdHash, err := Hash(req.NewPassword)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		LogError("update_hash_password", "Failed to hash new password", err)
		return
	}

	if err = c.repository.UpdatePassword(claims.UserID, newPasswdHash); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		LogError("update_db_error", "Failed to update password in database", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]string{
		"message": "password changed successfully",
	})

	LogSuccess("update_password_success", fmt.Sprintf("Password updated for user %s", claims.UserID))

}
