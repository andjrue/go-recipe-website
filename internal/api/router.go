// Package api contains the HTTP router and handlers.
package api

import (
	"net/http"

	"recipe-website/internal/repository"
)

func NewRouter(users repository.UserRepository) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", handleHealth)

	usersHandler := NewUserHandler(users)
	mux.HandleFunc("GET /api/users/{id}", usersHandler.GetByID)

	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
