// Package handler provides routes for requests
package handler

import (
	"encoding/json"
	"net/http"
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

// /recipes/{id}
func (h *RecipeHandler) GetByPK(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid recipe id")
		return
	}

	recipe, err := h.RecipeService.GetByPK(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, recipe)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
