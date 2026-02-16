package main

import (
	database "backend/AuthDatabase"
	handlers "backend/post/handlers"
	utils "backend/utils"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Could not load .env file")
	}

	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		log.Fatal("DB_URL not set in environment")
	}

	pool, err := database.Connect(connStr)
	if err != nil {
		log.Fatal("Connection failed:", err)
	} else {
		fmt.Println("Succes")
	}
	defer pool.Close()
	handlers.DB = pool

	// Like Handling
	ctx := context.Background()
	err = createLikesTable(ctx, pool)
	if err != nil {
		log.Fatal("Failed to create likes table:", err)
	}
	fmt.Println("✅ Likes table ready")

	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("/signup", database.SignupHandler(pool))
	mux.HandleFunc("/login", database.LoginHandler(pool))
	mux.HandleFunc("/posts/feed", handlers.GetPosts) // GET only, public
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	// Protected routes (with middleware)
	mux.Handle("/dashboard", utils.JWTMiddleware(http.HandlerFunc(database.DashboardHandler)))
	mux.Handle("/profile", utils.JWTMiddleware(http.HandlerFunc(database.DashboardHandler)))
	mux.Handle("/posts", utils.JWTMiddleware(http.HandlerFunc(handlers.CreatePost))) // POST only, protected

	handler := enableCORS(mux)

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func createLikesTable(ctx context.Context, db *pgxpool.Pool) error {
	query := `
	CREATE TABLE IF NOT EXISTS post_likes (
		id BIGSERIAL PRIMARY KEY,
		post_id BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
		user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE(post_id, user_id)
	);
	
	CREATE INDEX IF NOT EXISTS idx_post_likes_post_id ON post_likes(post_id);
	CREATE INDEX IF NOT EXISTS idx_post_likes_user_id ON post_likes(user_id);
	`
	
	_, err := db.Exec(ctx, query)
	return err
}
