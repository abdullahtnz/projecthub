package main

import (
	database "backend/AuthDatabase"
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

		// handle preflight
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

	mux := http.NewServeMux()

	mux.HandleFunc("/signup", database.SignupHandler(pool))
	mux.HandleFunc("/login", database.LoginHandler(pool))

	handler := enableCORS(mux)

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))

	// 2026 commit :)

}
