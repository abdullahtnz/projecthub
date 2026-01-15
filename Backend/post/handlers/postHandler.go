package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"backend/utils" // Update with your actual import path

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func CreatePost(w http.ResponseWriter, r *http.Request) {
	// Set response content type
	w.Header().Set("Content-Type", "application/json")

	// Get userID from context using utility function
	userID, err := utils.GetUserIDFromContext(r)
	if err != nil {
		fmt.Printf("DEBUG: Auth error: %v\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Authentication required. Please login first.",
		})
		return
	}

	fmt.Printf("DEBUG: Creating post for userID: %s\n", userID)

	// Parse multipart form (for file uploads)
	maxUploadSize := int64(10 << 20) // 10 MB
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		fmt.Printf("DEBUG: ParseMultipartForm error: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Unable to parse form data. Max size: 10MB",
		})
		return
	}

	// Get post content
	content := r.FormValue("content")
	if content == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Post content cannot be empty",
		})
		return
	}

	// Validate content length
	if len(content) > 5000 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Post content too long. Maximum 5000 characters.",
		})
		return
	}

	// Generate post ID
	postID := uuid.New().String()
	createdAt := time.Now()

	// Start transaction for atomic operations
	tx, err := DB.Begin(r.Context())
	if err != nil {
		fmt.Printf("DEBUG: Transaction begin error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to start transaction",
		})
		return
	}
	defer tx.Rollback(r.Context())

	// Insert post into database
	_, err = tx.Exec(
		r.Context(),
		`INSERT INTO posts (id, user_id, content, created_at) 
		 VALUES ($1, $2, $3, $4)`,
		postID, userID, content, createdAt,
	)
	if err != nil {
		fmt.Printf("DEBUG: Insert post error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to create post in database",
		})
		return
	}

	// Handle image uploads
	var imagePaths []string
	files := r.MultipartForm.File["images"]

	// Create uploads directory if it doesn't exist
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		fmt.Printf("DEBUG: Create directory error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to create upload directory",
		})
		return
	}

	// Limit number of images
	if len(files) > 5 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Maximum 5 images allowed per post",
		})
		return
	}

	// Process each image
	for _, fileHeader := range files {
		// Open uploaded file
		file, err := fileHeader.Open()
		if err != nil {
			fmt.Printf("DEBUG: File open error: %v\n", err)
			continue // Skip this file but continue with others
		}

		// Validate file size
		if fileHeader.Size > 5<<20 { // 5 MB per image
			file.Close()
			fmt.Printf("DEBUG: File too large: %s\n", fileHeader.Filename)
			continue
		}

		// Validate file type
		allowedTypes := map[string]bool{
			"image/jpeg": true,
			"image/jpg":  true,
			"image/png":  true,
			"image/gif":  true,
			"image/webp": true,
		}

		buffer := make([]byte, 512)
		_, err = file.Read(buffer)
		if err != nil && err != io.EOF {
			file.Close()
			fmt.Printf("DEBUG: File read error: %v\n", err)
			continue
		}

		file.Seek(0, 0) // Reset file pointer
		contentType := http.DetectContentType(buffer)
		if !allowedTypes[contentType] {
			file.Close()
			fmt.Printf("DEBUG: Invalid file type: %s\n", contentType)
			continue
		}

		// Generate unique filename
		fileExt := filepath.Ext(fileHeader.Filename)
		if fileExt == "" {
			// Default extension based on content type
			switch contentType {
			case "image/jpeg", "image/jpg":
				fileExt = ".jpg"
			case "image/png":
				fileExt = ".png"
			case "image/gif":
				fileExt = ".gif"
			case "image/webp":
				fileExt = ".webp"
			default:
				fileExt = ".jpg"
			}
		}

		imageID := uuid.New().String()
		filename := imageID + fileExt
		imagePath := filepath.Join(uploadDir, filename)

		// Save file to disk
		dst, err := os.Create(imagePath)
		if err != nil {
			file.Close()
			fmt.Printf("DEBUG: Create file error: %v\n", err)
			continue
		}

		// Copy file content
		if _, err := io.Copy(dst, file); err != nil {
			file.Close()
			dst.Close()
			fmt.Printf("DEBUG: File copy error: %v\n", err)
			os.Remove(imagePath) // Clean up failed file
			continue
		}

		// Close files
		file.Close()
		dst.Close()

		// Insert image record into database
		imageRecordID := uuid.New().String()
		_, err = tx.Exec(
			r.Context(),
			`INSERT INTO post_images (id, post_id, image_url, created_at) 
			 VALUES ($1, $2, $3, $4)`,
			imageRecordID, postID, filename, createdAt,
		)
		if err != nil {
			fmt.Printf("DEBUG: Insert image error: %v\n", err)
			// Continue with other images
		} else {
			imagePaths = append(imagePaths, filename)
		}
	}

	// Commit transaction
	if err := tx.Commit(r.Context()); err != nil {
		fmt.Printf("DEBUG: Transaction commit error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to save post",
		})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Post created successfully",
		"post": map[string]interface{}{
			"id":         postID,
			"user_id":    userID,
			"content":    content,
			"images":     imagePaths,
			"created_at": createdAt.Format(time.RFC3339),
		},
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
