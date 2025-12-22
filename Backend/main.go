package main

import (
	database "backend/authdatabase"
	"log"
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
	}
	defer pool.Close()

}
