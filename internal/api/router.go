// Package api contains the HTTP router and handlers.
package api

import (
	"net/http"

	"recipe-website/internal/repository"
)

func NewRouter(users repository.UserRepository) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", handleHealth)

	authHandler := NewAuthHandlerFromEnv(users)
	usersHandler := NewUserHandler(users)
	mux.Handle("GET /api/users/{id}", authHandler.RequireAuth(http.HandlerFunc(usersHandler.GetByID)))

	mux.HandleFunc("GET /api/auth/google/login", authHandler.Login)
	mux.HandleFunc("GET /api/auth/google/callback", authHandler.Callback)
	mux.HandleFunc("GET /api/me", authHandler.Me)
	mux.HandleFunc("POST /api/auth/logout", authHandler.Logout)

	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
