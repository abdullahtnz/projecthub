package database

import (
	utils "backend/utils"
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(utils.UserContextKey).(jwt.MapClaims)
	userID := claims["user_id"]
	email := claims["email"]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Welcome!",
		"user_id": userID,
		"email":   email,
	})
}
