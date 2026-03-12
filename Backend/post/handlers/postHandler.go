package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"backend/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	createdAt := time.Now()

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

	// Simple insert
	tx, err := DB.Begin(context.Background())
	if err != nil {
		fmt.Printf("DEBUG: Transaction begin error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Database error",
		})
		return
	}
	defer tx.Rollback(context.Background())

	// Insert post using transaction
	_, err = tx.Exec(context.Background(),
		"INSERT INTO posts (id, user_id, content, created_at) VALUES ($1, $2, $3, $4)",
		postID, userID, content, createdAt)

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

	var imageUrls []string
	files := r.MultipartForm.File["images"]

	if files != nil && len(files) > 0 {
		// Create uploads directory if it doesn't exist
		uploadDir := "uploads"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			fmt.Printf("DEBUG: Failed to create uploads directory: %v\n", err)
		} else {
			fmt.Printf("DEBUG: Uploads directory ready: %s\n", uploadDir)
		}

		// Process each uploaded image
		for _, fileHeader := range files {
			// Open the file
			file, err := fileHeader.Open()
			if err != nil {
				fmt.Printf("DEBUG: Failed to open file %s: %v\n", fileHeader.Filename, err)
				continue
			}
			defer file.Close()

			// Validate file size (max 5MB per image)
			if fileHeader.Size > 5<<20 {
				fmt.Printf("DEBUG: File %s too large: %d bytes\n", fileHeader.Filename, fileHeader.Size)
				continue
			}

			// Validate file type
			buff := make([]byte, 512)
			_, err = file.Read(buff)
			if err != nil {
				fmt.Printf("DEBUG: Failed to read file for validation: %v\n", err)
				continue
			}
			file.Seek(0, 0) // Reset file pointer

			fileType := http.DetectContentType(buff)
			if !strings.HasPrefix(fileType, "image/") {
				fmt.Printf("DEBUG: File %s is not an image: %s\n", fileHeader.Filename, fileType)
				continue
			}

			// Generate unique filename
			fileExt := filepath.Ext(fileHeader.Filename)
			if fileExt == "" {
				// Default to .jpg if no extension
				fileExt = ".jpg"
			}

			imageID := uuid.New().String()
			newFilename := imageID + fileExt
			filePath := filepath.Join(uploadDir, newFilename)

			// Save file to disk
			dst, err := os.Create(filePath)
			if err != nil {
				fmt.Printf("DEBUG: Failed to create file %s: %v\n", filePath, err)
				continue
			}
			defer dst.Close()

			// Copy file content
			if _, err := io.Copy(dst, file); err != nil {
				fmt.Printf("DEBUG: Failed to save file %s: %v\n", filePath, err)
				os.Remove(filePath) // Clean up failed file
				continue
			}

			// Insert image record into database
			_, err = tx.Exec(context.Background(),
				"INSERT INTO post_images (post_id, image_url) VALUES ($1, $2)",
				postID, newFilename) 

			if err != nil {
				fmt.Printf("DEBUG: Failed to insert image record: %v\n", err)
				continue
			}

			// Add to response
			imageUrls = append(imageUrls, newFilename)
			fmt.Printf("DEBUG: Image saved: %s\n", newFilename)
		}

	}

	if err := tx.Commit(context.Background()); err != nil {
		fmt.Printf("DEBUG CreatePost: Transaction commit error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to save post",
		})
		return
	}

	if files == nil || len(files) == 0 {
		fmt.Println("DEBUG: No images uploaded with this post")
	}

	// Success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Post created successfully",
		"post_id": postID,
		"user_id": userID,
		"content": content,
		"images":  imageUrls,
	})
}

func GetPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from context if authenticated
	userID, _ := utils.GetUserIDFromContext(r)

	var query string
	var rows pgx.Rows
	var err error

	if userID == "" {
		// Public feed query (no liked_by_me field)
		query = `
            SELECT 
                p.id, 
                p.user_id, 
                p.content, 
                p.created_at,
                COALESCE(
                    json_agg(DISTINCT pi.image_url) FILTER (WHERE pi.image_url IS NOT NULL),
                    '[]'::json
                ) as images,
                COUNT(DISTINCT l.id) as like_count
            FROM posts p
            LEFT JOIN post_images pi ON p.id = pi.post_id
            LEFT JOIN likes l ON p.id = l.post_id
            GROUP BY p.id, p.user_id, p.content, p.created_at
            ORDER BY p.created_at DESC
        `
		rows, err = DB.Query(r.Context(), query)
	} else {
		// Authenticated feed query with liked_by_me
		query = `
            SELECT 
                p.id, 
                p.user_id as post_user_id, 
                p.content, 
                p.created_at,
                COALESCE(
                    json_agg(DISTINCT pi.image_url) FILTER (WHERE pi.image_url IS NOT NULL),
                    '[]'::json
                ) as images,
                COUNT(DISTINCT l.id) as like_count,
                EXISTS(SELECT 1 FROM likes WHERE post_id = p.id AND user_id = $1) as liked_by_me
            FROM posts p
            LEFT JOIN post_images pi ON p.id = pi.post_id
            LEFT JOIN likes l ON p.id = l.post_id
            GROUP BY p.id, p.user_id, p.content, p.created_at
            ORDER BY p.created_at DESC
        `
		rows, err = DB.Query(r.Context(), query, userID)
	}

	if err != nil {
		fmt.Printf("Database error in GetPosts: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get posts: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var posts []map[string]interface{}

	for rows.Next() {
		if userID == "" {
			// Scan for public feed (6 columns)
			var id, postUserID, content string
			var createdAt time.Time
			var images []string
			var likeCount int

			err := rows.Scan(&id, &postUserID, &content, &createdAt, &images, &likeCount)
			if err != nil {
				fmt.Printf("Error scanning row: %v\n", err)
				continue
			}

			posts = append(posts, map[string]interface{}{
				"id":          id,
				"user_id":     postUserID,
				"content":     content,
				"images":      images,
				"created_at":  createdAt,
				"like_count":  likeCount,
				"liked_by_me": false,
			})
		} else {
			// Scan for authenticated feed (7 columns)
			var id, postUserID, content string
			var createdAt time.Time
			var images []string
			var likeCount int
			var likedByMe bool

			err := rows.Scan(&id, &postUserID, &content, &createdAt, &images, &likeCount, &likedByMe)
			if err != nil {
				fmt.Printf("Error scanning row: %v\n", err)
				continue
			}

			posts = append(posts, map[string]interface{}{
				"id":          id,
				"user_id":     postUserID,
				"content":     content,
				"images":      images,
				"created_at":  createdAt,
				"like_count":  likeCount,
				"liked_by_me": likedByMe,
			})
		}
	}

	if err = rows.Err(); err != nil {
		fmt.Printf("Rows iteration error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Error processing results",
		})
		return
	}

	json.NewEncoder(w).Encode(posts)
}

func GetMyPosts(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    // Get user ID from context (authentication required)
    userID, err := utils.GetUserIDFromContext(r)
    if err != nil {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Authentication required",
        })
        return
    }

    // Query that filters by the authenticated user's ID
    query := `
        SELECT 
            p.id, 
            p.user_id, 
            p.content, 
            p.created_at,
            COALESCE(
                json_agg(DISTINCT pi.image_url) FILTER (WHERE pi.image_url IS NOT NULL),
                '[]'::json
            ) as images,
            COUNT(DISTINCT l.id) as like_count,
            EXISTS(SELECT 1 FROM likes WHERE post_id = p.id AND user_id = $1) as liked_by_me
        FROM posts p
        LEFT JOIN post_images pi ON p.id = pi.post_id
        LEFT JOIN likes l ON p.id = l.post_id
        WHERE p.user_id = $1  -- This filters by user ID
        GROUP BY p.id, p.user_id, p.content, p.created_at
        ORDER BY p.created_at DESC
    `

    rows, err := DB.Query(r.Context(), query, userID)
    if err != nil {
        fmt.Printf("Database error in GetMyPosts: %v\n", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Failed to get your posts",
        })
        return
    }
    defer rows.Close()

    var posts []map[string]interface{}

    for rows.Next() {
        var id, postUserID, content string
        var createdAt time.Time
        var images []string
        var likeCount int
        var likedByMe bool

        err := rows.Scan(&id, &postUserID, &content, &createdAt, &images, &likeCount, &likedByMe)
        if err != nil {
            fmt.Printf("Error scanning row: %v\n", err)
            continue
        }

        posts = append(posts, map[string]interface{}{
            "id":          id,
            "user_id":     postUserID,
            "content":     content,
            "images":      images,
            "created_at":  createdAt,
            "like_count":  likeCount,
            "liked_by_me": likedByMe, // Will always be true for your own posts?
        })
    }

    if err = rows.Err(); err != nil {
        fmt.Printf("Rows iteration error: %v\n", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Error processing results",
        })
        return
    }

    json.NewEncoder(w).Encode(posts)
}

func GetUserPosts(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    targetUserID := r.URL.Query().Get("user_id")
    if targetUserID == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "User ID required"})
        return
    }

    currentUserID, _ := utils.GetUserIDFromContext(r)

    // Use NULLIF to handle empty string
    query := `
        SELECT 
            p.id, 
            p.user_id, 
            p.content, 
            p.created_at,
            COALESCE(
                json_agg(DISTINCT pi.image_url) FILTER (WHERE pi.image_url IS NOT NULL),
                '[]'::json
            ) as images,
            COUNT(DISTINCT l.id) as like_count,
            $2::uuid IS NOT NULL AND EXISTS(
                SELECT 1 FROM likes WHERE post_id = p.id AND user_id = $2
            ) as liked_by_me
        FROM posts p
        LEFT JOIN post_images pi ON p.id = pi.post_id
        LEFT JOIN likes l ON p.id = l.post_id
        WHERE p.user_id = $1
        GROUP BY p.id, p.user_id, p.content, p.created_at
        ORDER BY p.created_at DESC
    `

    var currentUserIDParam interface{}
    if currentUserID == "" {
        currentUserIDParam = nil
    } else {
        currentUserIDParam = currentUserID
    }

    rows, err := DB.Query(r.Context(), query, targetUserID, currentUserIDParam)
    if err != nil {
        fmt.Printf("Database error in GetUserPosts: %v\n", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get user posts"})
        return
    }
    defer rows.Close()

    var posts []map[string]interface{}

    for rows.Next() {
        var id, postUserID, content string
        var createdAt time.Time
        var images []string
        var likeCount int
        var likedByMe bool

        err := rows.Scan(&id, &postUserID, &content, &createdAt, &images, &likeCount, &likedByMe)
        if err != nil {
            fmt.Printf("Error scanning row: %v\n", err)
            continue
        }

        posts = append(posts, map[string]interface{}{
            "id":          id,
            "user_id":     postUserID,
            "content":     content,
            "images":      images,
            "created_at":  createdAt,
            "like_count":  likeCount,
            "liked_by_me": likedByMe,
        })
    }

    json.NewEncoder(w).Encode(posts)
}

func GetUsername(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    // Get user_id from query parameter
    userID := r.URL.Query().Get("user_id")
    if userID == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "User ID required",
        })
        return
    }

    // Query database for username
    var username string
    err := DB.QueryRow(r.Context(),
        "SELECT username FROM users WHERE id = $1", userID).Scan(&username)

    if err != nil {
        if err.Error() == "no rows in result set" {
            w.WriteHeader(http.StatusNotFound)
            json.NewEncoder(w).Encode(map[string]string{
                "error": "User not found",
            })
            return
        }
        
        fmt.Printf("Database error in GetUsername: %v\n", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Database error",
        })
        return
    }

    // Success response
    json.NewEncoder(w).Encode(map[string]string{
        "user_id": userID,
        "username": username,
    })
}