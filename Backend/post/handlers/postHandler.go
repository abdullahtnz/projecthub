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

func GetPosts(w http.ResponseWriter, r *http.Request) {

	rows, err := DB.Query(
		context.Background(),
		`SELECT p.id, p.user_id, p.content, p.created_at,
                COALESCE(array_agg(pi.image_url) FILTER (WHERE pi.image_url IS NOT NULL), '{}') AS images
         FROM posts p
         LEFT JOIN post_images pi ON pi.post_id = p.id
         GROUP BY p.id
         ORDER BY p.created_at DESC`,
	)
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type PostResp struct {
		ID        string   `json:"id"`
		UserID    string   `json:"user_id"`
		Content   string   `json:"content"`
		Images    []string `json:"images"`
		CreatedAt string   `json:"created_at"`
	}

	var posts []PostResp

	for rows.Next() {
		var p PostResp
		var images []string

		err := rows.Scan(&p.ID, &p.UserID, &p.Content, &p.CreatedAt, &images)
		if err != nil {
			http.Error(w, "Failed to read post", http.StatusInternalServerError)
			return
		}
		p.Images = images
		posts = append(posts, p)
	}

	json.NewEncoder(w).Encode(posts)
}
