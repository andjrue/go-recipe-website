// Package api contains the HTTP router and handlers.
package api

import (
	"net/http"

	"recipe-website/internal/repository"
)

func NewRouter(users repository.UserRepository, recipes repository.RecipeRepository) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", handleHealth)

	authHandler := NewAuthHandlerFromEnv(users)
	usersHandler := NewUserHandler(users)
	mux.Handle("GET /api/users/{id}", authHandler.RequireAuth(http.HandlerFunc(usersHandler.GetByID)))
	recipesHandler := NewRecipeHandler(recipes)
	mux.Handle("GET /api/recipes", authHandler.RequireAuth(http.HandlerFunc(recipesHandler.List)))
	mux.Handle("POST /api/recipes", authHandler.RequireAuth(http.HandlerFunc(recipesHandler.Create)))
	mux.Handle("GET /api/recipes/{id}", authHandler.RequireAuth(http.HandlerFunc(recipesHandler.GetByID)))
	mux.Handle("PUT /api/recipes/{id}", authHandler.RequireAuth(http.HandlerFunc(recipesHandler.Update)))
	mux.Handle("DELETE /api/recipes/{id}", authHandler.RequireAuth(http.HandlerFunc(recipesHandler.Delete)))

	mux.HandleFunc("GET /api/auth/google/login", authHandler.Login)
	mux.HandleFunc("GET /api/auth/google/callback", authHandler.Callback)
	mux.HandleFunc("GET /api/me", authHandler.Me)
	mux.HandleFunc("POST /api/auth/logout", authHandler.Logout)

	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
