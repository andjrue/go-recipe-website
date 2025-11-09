// Package router stores api routes
package router

import (
	"recipe-website/internal/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRoutes(h *handler.Handlers) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.CleanPath)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/recipes", func(r chi.Router) {
			r.Post("/", h.Recipe.Create)
			r.Get("/{id}", h.Recipe.GetByPK)
		})
	})

	return r
}
