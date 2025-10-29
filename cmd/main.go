package main

import (
	"context"
	"log"
	"recipe-website/internal/database"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("error loading env file")
	}

	db := database.ConnectToPostgres()
	defer db.Close(context.Background())

	storage := database.NewPostgresDatabase(db)
	_ = storage
}
