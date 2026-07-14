package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"recipe-website/internal/api"
	"recipe-website/internal/database"
	"recipe-website/internal/repository"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file loaded; using process environment")
	}

	ctx := context.Background()

	pool, err := database.ConnectToPostgres(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// repositories wire onto the pool; handlers will depend on these interfaces
	users := repository.NewUserPostgres(pool)
	recipes := repository.NewRecipePostgres(pool)

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           api.NewRouter(users, recipes),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
