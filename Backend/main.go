package main

import (
	database "backend/AuthDatabase"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

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
		fmt.Print("Succes")
	}
	defer pool.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/signup", database.SignupHandler(pool))

}
