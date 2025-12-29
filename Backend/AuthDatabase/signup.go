package database

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// UserSignupRequest is what we parse from the POST body
type UserSignupRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// User is the actual record in your database
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

func SignupHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// MEthod check
		if r.Method != http.MethodPost {
			http.Error(w, "Method is not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Converting request to json
		var req UserSignupRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Checking everything exist
		if req.Email == "" || req.Username == "" || req.Password == "" {
			http.Error(w, "All fields are required", http.StatusBadRequest)
			return
		}

		// Password strength
		if len(req.Password) < 6 {
			http.Error(w, "Password must be at least 8 characters long", http.StatusBadRequest)
			return
		}

		// Repeatedness check
		var exist bool
		errors := pool.QueryRow(
			context.Background(),
			`SELECT EXISTS(
	    	    SELECT 1 FROM users WHERE email=$1 OR username=$2
	    	)`,
			req.Email, req.Username,
		).Scan(&exist)

		if errors != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		if exist {
			http.Error(w, "Email or username already exists", http.StatusConflict)
			return
		}

		// Hashing pw
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		// Add user into database
		var id int
		err = pool.QueryRow(
			context.Background(),
			`INSERT INTO users (email, username, password_hash, created_at)
		     VALUES ($1, $2, $3, $4) RETURNING id`,
			req.Email, req.Username, string(hashedPassword), time.Now(),
		).Scan(&id)

	}
}
