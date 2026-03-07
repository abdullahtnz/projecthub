package main

import (
	database "backend/AuthDatabase"
	handlers "backend/post/handlers"
	utils "backend/utils"
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

	// Like routes
	mux.Handle("/posts/{id}/like", utils.JWTMiddleware(http.HandlerFunc(handlers.LikePost)))
	mux.Handle("/posts/{id}/unlike", utils.JWTMiddleware(http.HandlerFunc(handlers.UnlikePost)))
	mux.Handle("/posts/{id}/like-status", utils.JWTMiddleware(http.HandlerFunc(handlers.GetLikeStatus)))
	mux.Handle("/posts/{id}/likes", utils.JWTMiddleware(http.HandlerFunc(handlers.GetPostLikes)))

	handler := enableCORS(mux)

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

