package auth

import (
	"authd/internal/model"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
)

func (c *Config) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var refreshReq model.RefreshRequest

	err := json.NewDecoder(r.Body).Decode(&refreshReq)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		LogError("refresh_request_decode", "Failed to decode refresh request", err)
		return
	}

	// Validate refresh token
	var validate = validator.New()
	err = validate.Struct(refreshReq)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(FormatValidationErrors(err))
		return
	}

	tokenString := strings.TrimSpace(refreshReq.RefreshToken)

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return c.jwtKey, nil
	})

	if err != nil || token == nil || !token.Valid {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		LogError("refresh_invalid_token", "Invalid or expired refresh token", err)
		return
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		LogError("refresh_invalid_claims", "Invalid token claims", nil)
		return
	}

	// Generate new access token (15 minutes expiry)
	authToken, err := c.GenerateJWT(claims.UserID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		LogError("refresh_jwt_generation", "Failed to generate new JWT token", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(model.LoginResponse{
		Token:   authToken,
		Message: "Token refreshed successfully",
	})

	LogSuccess("refresh_success", "Access token refreshed successfully")
}
