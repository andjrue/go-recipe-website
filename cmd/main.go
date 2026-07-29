package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"recipe-website/internal/api"
	"recipe-website/internal/database"
	"recipe-website/internal/repository"
	"recipe-website/internal/storage"

	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file loaded; using process environment")
	}

	ctx := context.Background()

	pool, err := database.ConnectToPostgres(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	migrationContext, cancelMigrations := context.WithTimeout(ctx, 60*time.Second)
	defer cancelMigrations()
	if err := database.Migrate(migrationContext, pool); err != nil {
		return err
	}

	// repositories wire onto the pool; handlers will depend on these interfaces
	users := repository.NewUserPostgres(pool)
	recipes := repository.NewRecipePostgres(pool)
	imageDirectory := os.Getenv("IMAGE_STORAGE_DIR")
	if imageDirectory == "" {
		imageDirectory = "data/images"
	}
	images, err := storage.NewFilesystemImageStore(imageDirectory)
	if err != nil {
		return err
	}

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           api.NewRouter(users, recipes, pool, images),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Listening on %s", addr)
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-shutdownSignal.Done():
		log.Println("Shutting down HTTP server")
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			return err
		}
	}
	return nil
}
