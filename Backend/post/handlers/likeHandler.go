package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"backend/utils"
)

// LikePost handles POST /posts/{id}/like
func LikePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get post ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	postID := pathParts[2]

	// Get userID from context
	userID, err := utils.GetUserIDFromContext(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
		return
	}

	// Check if post exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)`
	err = DB.QueryRow(r.Context(), checkQuery, postID).Scan(&exists)
	if err != nil || !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Post not found"})
		return
	}

	// Add like
	insertQuery := `INSERT INTO likes (post_id, user_id) VALUES ($1, $2) ON CONFLICT (post_id, user_id) DO NOTHING`
	_, err = DB.Exec(r.Context(), insertQuery, postID, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to like post"})
		return
	}

	// Get updated like count
	var likeCount int
	countQuery := `SELECT COUNT(*) FROM likes WHERE post_id = $1`
	err = DB.QueryRow(r.Context(), countQuery, postID).Scan(&likeCount)
	if err != nil {
		likeCount = 0
	}

	// Return response
	json.NewEncoder(w).Encode(map[string]interface{}{
		"liked":      true,
		"like_count": likeCount,
		"post_id":    postID,
	})
}

// UnlikePost handles POST /posts/{id}/unlike
func UnlikePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get post ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	postID := pathParts[2]

	// Get userID from context
	userID, err := utils.GetUserIDFromContext(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
		return
	}

	// Remove like
	deleteQuery := `DELETE FROM likes WHERE post_id = $1 AND user_id = $2`
	_, err = DB.Exec(r.Context(), deleteQuery, postID, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to unlike post"})
		return
	}

	// Get updated like count
	var likeCount int
	countQuery := `SELECT COUNT(*) FROM likes WHERE post_id = $1`
	err = DB.QueryRow(r.Context(), countQuery, postID).Scan(&likeCount)
	if err != nil {
		likeCount = 0
	}

	// Return response
	json.NewEncoder(w).Encode(map[string]interface{}{
		"liked":      false,
		"like_count": likeCount,
		"post_id":    postID,
	})
}

// GetLikeStatus handles GET /posts/{id}/like-status
func GetLikeStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get post ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	postID := pathParts[2]

	// Get userID from context
	userID, err := utils.GetUserIDFromContext(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
		return
	}

	// Check if user liked the post
	var liked bool
	likeQuery := `SELECT EXISTS(SELECT 1 FROM likes WHERE post_id = $1 AND user_id = $2)`
	err = DB.QueryRow(r.Context(), likeQuery, postID, userID).Scan(&liked)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to check like status"})
		return
	}

	// Get total like count
	var likeCount int
	countQuery := `SELECT COUNT(*) FROM likes WHERE post_id = $1`
	err = DB.QueryRow(r.Context(), countQuery, postID).Scan(&likeCount)
	if err != nil {
		likeCount = 0
	}

	// Return response
	json.NewEncoder(w).Encode(map[string]interface{}{
		"liked":      liked,
		"like_count": likeCount,
		"post_id":    postID,
	})
}

// GetPostLikes handles GET /posts/{id}/likes (optional - shows who liked)
func GetPostLikes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get post ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	postID := pathParts[2]

	// Get users who liked this post
	query := `
		SELECT u.id, u.username, u.email, l.created_at
		FROM likes l
		JOIN users u ON l.user_id = u.id
		WHERE l.post_id = $1
		ORDER BY l.created_at DESC
	`

	rows, err := DB.Query(r.Context(), query, postID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get likes"})
		return
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id, username, email string
		var likedAt interface{}

		err := rows.Scan(&id, &username, &email, &likedAt)
		if err != nil {
			continue
		}

		users = append(users, map[string]interface{}{
			"user_id":  id,
			"username": username,
			"email":    email,
			"liked_at": likedAt,
		})
	}

	json.NewEncoder(w).Encode(users)
}
