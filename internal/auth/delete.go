package auth

import (
	"authd/internal/repository"
	"encoding/json"
	"net/http"
)

type DeleteAccountRequest struct {
	Password string `json:"password"`
}

type DeleteAccountResponse struct {
	Message string `json:"message"`
}

func NewDeleteHandler(repo *repository.Repository, jwtKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		claims, err := ValidateAuth(r.Header, jwtKey)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			LogError("delete_account_auth", "Invalid authentication", err)
			return
		}

		var req DeleteAccountRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
			LogError("delete_account_decode", "Failed to decode request", err)
			return
		}

		if req.Password == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "password is required"})
			LogError("delete_account_missing_password", "Missing password", nil)
			return
		}

		// Verify password
		storedPasswd, err := repo.GetUserCredentials(claims.UserID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to verify credentials"})
			LogError("delete_account_verify", "Failed to get stored password", err)
			return
		}

		if req.Password != storedPasswd {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid password"})
			LogError("delete_account_invalid_password", "Invalid password provided", nil)
			return
		}

		// Delete user account
		if err := repo.DeleteUser(claims.UserID); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete account"})
			LogError("delete_account_delete", "Failed to delete user account", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(DeleteAccountResponse{
			Message: "Account deleted successfully",
		})

		LogSuccess("delete_account_success", "Account deleted for user "+claims.UserID)
	}
}
