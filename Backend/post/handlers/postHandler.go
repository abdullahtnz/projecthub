package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func CreatePost(w http.ResponseWriter, r *http.Request) {

	// Parse form data
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")
	userID := r.Context().Value("userID").(string) // from auth
	postID := uuid.New().String()

	// Use context.Background() with pgxpool
	_, err := DB.Exec(
		context.Background(),
		"INSERT INTO posts (id, user_id, content) VALUES ($1, $2, $3)",
		postID, userID, content,
	)
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	// Handle images
	files := r.MultipartForm.File["images"]
	var images []string

	for _, f := range files {
		file, _ := f.Open()
		defer file.Close()

		imageID := uuid.New().String()
		imagePath := filepath.Join("uploads", imageID+filepath.Ext(f.Filename))

		dst, _ := os.Create(imagePath)
		defer dst.Close()
		dst.ReadFrom(file)

		// Insert image with pgxpool
		_, err = DB.Exec(
			context.Background(),
			"INSERT INTO post_images (id, post_id, image_url) VALUES ($1, $2, $3)",
			uuid.New().String(), postID, imagePath,
		)
		if err != nil {
			http.Error(w, "Failed to save image", http.StatusInternalServerError)
			return
		}

		images = append(images, imagePath)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Post created",
		"post": map[string]interface{}{
			"id":      postID,
			"content": content,
			"images":  images,
		},
	})
}
