package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"backend/utils" // Update with your actual import path

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func CreatePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get userID from context
	userID, err := utils.GetUserIDFromContext(r)
	if err != nil {
		fmt.Printf("DEBUG CreatePost: Auth error: %v\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
		return
	}

	fmt.Printf("DEBUG CreatePost: User authenticated: %s\n", userID)

	// Parse form data
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		fmt.Printf("DEBUG CreatePost: Parse error: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unable to parse form"})
		return
	}

	content := r.FormValue("content")
	if content == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Content required"})
		return
	}

	// Generate post ID
	postID := uuid.New().String()

	fmt.Printf("DEBUG CreatePost: Inserting - PostID: %s, UserID: %s, Content: %s\n",
		postID, userID, content)

	// Check if user exists in database
	var dbUserID string
	err = DB.QueryRow(context.Background(),
		"SELECT id FROM users WHERE id = $1", userID).Scan(&dbUserID)

	if err != nil {
		fmt.Printf("DEBUG CreatePost: User not found: %v\n", err)
		fmt.Printf("DEBUG CreatePost: Looking for user ID: %s\n", userID)

		// List all users
		rows, _ := DB.Query(context.Background(), "SELECT id, email FROM users")
		for rows.Next() {
			var id, email string
			rows.Scan(&id, &email)
			fmt.Printf("DEBUG CreatePost: Existing user - ID: %s, Email: %s\n", id, email)
		}
		rows.Close()

		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "User not found in database. Please login again.",
		})
		return
	}

	fmt.Printf("DEBUG CreatePost: User verified in DB: %s\n", dbUserID)

	// Try simple insert first
	_, err = DB.Exec(context.Background(),
		"INSERT INTO posts (id, user_id, content) VALUES ($1, $2, $3)",
		postID, userID, content)

	if err != nil {
		fmt.Printf("DEBUG CreatePost: Database error: %v\n", err)

		// Check if it's a PostgreSQL error
		if pgErr, ok := err.(interface{ Code() string }); ok {
			fmt.Printf("DEBUG CreatePost: PostgreSQL Error Code: %s\n", pgErr.Code())
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Database error: " + err.Error(),
		})
		return
	}

	fmt.Println("DEBUG CreatePost: Post created successfully!")

	// Success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Post created successfully",
		"post_id": postID,
		"user_id": userID,
		"content": content,
	})
}
func GetPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := DB.Query(r.Context(), "SELECT id, user_id, content, created_at FROM posts ORDER BY created_at DESC")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get posts: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var posts []map[string]interface{}

	for rows.Next() {
		var id, userID, content string
		var createdAt time.Time
		if err := rows.Scan(&id, &userID, &content, &createdAt); err != nil {
			// Log the error but continue processing other rows
			continue
		}

		// Query images
		imgRows, _ := DB.Query(r.Context(), "SELECT image_url FROM post_images WHERE post_id=$1", id)
		var images []string
		for imgRows.Next() {
			var img string
			imgRows.Scan(&img)
			images = append(images, img)
		}
		imgRows.Close()

		posts = append(posts, map[string]interface{}{
			"id":         id,
			"user_id":    userID,
			"content":    content,
			"images":     images,
			"created_at": createdAt,
		})
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Error processing posts: " + err.Error(),
		})
		return
	}

	// Always return an array, even if empty
	json.NewEncoder(w).Encode(posts)
}
