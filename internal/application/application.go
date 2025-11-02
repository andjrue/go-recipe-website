// Package application stores all of our applications depedencies in one central location
package application

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
)

type App struct {
	db     *pgx.Conn
	router http.Handler
}

func New() (*App, error) {
	return nil, nil
	// Eventually this will house all of our dependencies

}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":3000",
		Handler: a.router,
	}

	err := server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("unable to start server: %w", err)
	}

	return nil
}
