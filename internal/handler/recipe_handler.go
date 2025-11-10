// Package handler provides routes for requests
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"recipe-website/internal/domain"
	"recipe-website/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

/*
 *
 * Routes we need:
 * GetByPK - id url param
 * GetAllByUserID - userID url param
 * create - Name, TimeToCook (optional), Description (optional), UserID body
 * update - Name (optional), TimeToCook (optional), Description (optional) body, recipeID url param
 * delete - id url param
 *
 */

type RecipeHandler struct {
	RecipeService *service.RecipeService
}

func NewRecipeHandler(recipeService *service.RecipeService) *RecipeHandler {
	return &RecipeHandler{
		RecipeService: recipeService,
	}
}

// GetByPK - GET /recipes/{id}
func (h *RecipeHandler) GetByPK(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, fmt.Errorf("recipe id is malformed: %w", domain.ErrValidation))
		return
	}

	recipe, err := h.RecipeService.GetByPK(r.Context(), id)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, recipe)
}

// Create - POST /recipes
func (h *RecipeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateRecipeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, fmt.Errorf("invalid request body: %w", domain.ErrValidation))
		return
	}

	recipe, err := h.RecipeService.Create(r.Context(), &req)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, recipe)
}

