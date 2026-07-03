package main

import (
	"context"
	"log"

	"recipe-website/internal/database"
	"recipe-website/internal/repository"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("error loading env file")
	}

	ctx := context.Background()

	pool, err := database.ConnectToPostgres(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// repositories wire onto the pool; handlers will depend on these interfaces
	users := repository.NewUserPostgres(pool)
	_ = users

	log.Println("Successfully connected to postgres")
}
