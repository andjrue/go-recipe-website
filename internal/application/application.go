// Package application stores all of our applications dependencies in one central location
package application

import (
	"context"
	"fmt"
	"net/http"
	"recipe-website/internal/database"
	"recipe-website/internal/handler"
	"recipe-website/internal/repository"
	"recipe-website/internal/router"
	"recipe-website/internal/service"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	db     *pgxpool.Pool
	router http.Handler
}

func New() (*App, error) {
	db := database.ConnectToPostgres()
	repos := repository.SetRepositories(db)
	service := service.SetServices(repos)
	handlers := handler.NewHandler(service)
	r := router.SetupRoutes(handlers)
	
	return &App{
		db: db,
		router: r,
	}, nil

}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":3000",
		Handler: a.router,
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	go func() {
		fmt.Printf("Server starting on localhost:3000\n")
		if err := http.ListenAndServe(server.Addr, server.Handler); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server failed to listen: %w", err)
			cancel()
		}
	} ()
	
	<-ctx.Done()
	fmt.Println("\nShutting down the server gracefully")
	
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer shutdownCancel()
	
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	a.db.Close()

	fmt.Println("Server stopped.")

	return nil
}
