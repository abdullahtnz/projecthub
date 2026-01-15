package database

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// UserSignupRequest is what we parse from the POST body
type UserSignupRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Response struct {
	Message string `json:"message"`
	UserID  string `json:"user_id,omitempty"`
}

func SignupHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Method check
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(Response{Message: "Method not allowed"})
			return
		}

		// Parse JSON request
		var req UserSignupRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{Message: "Invalid JSON"})
			return
		}

		// Validate required fields
		if req.Email == "" || req.Username == "" || req.Password == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{Message: "All fields are required"})
			return
		}

		// Password strength
		if len(req.Password) < 6 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{Message: "Password must be at least 6 characters long"})
			return
		}

		fmt.Printf("DEBUG Signup: Checking for existing user - Email: %s, Username: %s\n",
			req.Email, req.Username)

		// Check if user already exists
		var exists bool
		err = pool.QueryRow(
			context.Background(),
			`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 OR username = $2)`,
			req.Email, req.Username,
		).Scan(&exists)

		if err != nil {
			fmt.Printf("DEBUG Signup: Check existence error: %v\n", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{Message: "Database error checking user"})
			return
		}

		if exists {
			fmt.Printf("DEBUG Signup: User already exists - Email: %s, Username: %s\n",
				req.Email, req.Username)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(Response{Message: "Email or username already exists"})
			return
		}

		// Hash password
		fmt.Printf("DEBUG Signup: Hashing password for user: %s\n", req.Email)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Printf("DEBUG Signup: Password hash error: %v\n", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{Message: "Server error"})
			return
		}

		// Generate UUID for user
		userID := uuid.New().String()
		fmt.Printf("DEBUG Signup: Generated user ID: %s\n", userID)

		// Insert user into database
		fmt.Printf("DEBUG Signup: Inserting user into database - ID: %s, Email: %s\n",
			userID, req.Email)

		// Option 1: Let database generate UUID (if you have DEFAULT gen_random_uuid())
		var returnedID string
		err = pool.QueryRow(
			context.Background(),
			`INSERT INTO users (id, email, username, password_hash, created_at)
		     VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			userID, req.Email, req.Username, string(hashedPassword), time.Now(),
		).Scan(&returnedID)

		if err != nil {
			fmt.Printf("DEBUG Signup: Database insert error: %v\n", err)

			// Try without specifying ID (let DB generate it)
			fmt.Printf("DEBUG Signup: Trying without specifying ID...\n")
			err = pool.QueryRow(
				context.Background(),
				`INSERT INTO users (email, username, password_hash, created_at)
				 VALUES ($1, $2, $3, $4) RETURNING id`,
				req.Email, req.Username, string(hashedPassword), time.Now(),
			).Scan(&returnedID)

			if err != nil {
				fmt.Printf("DEBUG Signup: Second insert attempt also failed: %v\n", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(Response{Message: "Failed to create user: " + err.Error()})
				return
			}
			userID = returnedID
		} else {
			userID = returnedID
		}

		fmt.Printf("DEBUG Signup: User created successfully - ID: %s\n", userID)

		// Verify user was inserted
		var verifyID string
		err = pool.QueryRow(
			context.Background(),
			"SELECT id FROM users WHERE email = $1",
			req.Email,
		).Scan(&verifyID)

		if err != nil {
			fmt.Printf("DEBUG Signup: Verification failed: %v\n", err)
		} else {
			fmt.Printf("DEBUG Signup: Verified user exists in DB - ID: %s\n", verifyID)
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":  "Signup successful",
			"user_id":  userID,
			"email":    req.Email,
			"username": req.Username,
		})
	}
}
