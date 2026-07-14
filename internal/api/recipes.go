package api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"recipe-website/internal/apperror"
	"recipe-website/internal/repository"
)

const maxRecipeRequestBytes = 1 << 20

type RecipeHandler struct {
	recipes repository.RecipeRepository
}

type recipeRequest struct {
	Name        string              `json:"name"`
	RecipeType  string              `json:"recipeType"`
	TimeToCook  string              `json:"timeToCook"`
	Description string              `json:"description"`
	Ingredients []ingredientRequest `json:"ingredients"`
	Steps       []stepRequest       `json:"steps"`
}

type ingredientRequest struct {
	Name     string `json:"name"`
	Quantity string `json:"quantity"`
}

type stepRequest struct {
	Instruction string `json:"instruction"`
}

type recipeResponse struct {
	ID           string               `json:"id"`
	Name         string               `json:"name"`
	RecipeType   string               `json:"recipeType"`
	TimeToCook   string               `json:"timeToCook"`
	Description  string               `json:"description"`
	UserID       string               `json:"userID"`
	DatePosted   time.Time            `json:"datePosted"`
	LastEditedAt time.Time            `json:"lastEditedAt"`
	Ingredients  []ingredientResponse `json:"ingredients"`
	Steps        []stepResponse       `json:"steps"`
	Images       []imageResponse      `json:"images"`
}

type recipeSummaryResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	RecipeType   string    `json:"recipeType"`
	TimeToCook   string    `json:"timeToCook"`
	Description  string    `json:"description"`
	UserID       string    `json:"userID"`
	DatePosted   time.Time `json:"datePosted"`
	LastEditedAt time.Time `json:"lastEditedAt"`
}

type ingredientResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Quantity string `json:"quantity"`
	Position int    `json:"position"`
}

type stepResponse struct {
	ID          string `json:"id"`
	StepNumber  int    `json:"stepNumber"`
	Instruction string `json:"instruction"`
}

type imageResponse struct {
	ID          string    `json:"id"`
	FileName    string    `json:"fileName"`
	ContentType string    `json:"contentType"`
	FileSize    int64     `json:"fileSize"`
	Position    int       `json:"position"`
	IsCover     bool      `json:"isCover"`
	UploadedAt  time.Time `json:"uploadedAt"`
}

func NewRecipeHandler(recipes repository.RecipeRepository) *RecipeHandler {
	return &RecipeHandler{recipes: recipes}
}

func (h *RecipeHandler) List(w http.ResponseWriter, r *http.Request) {
	recipes, err := h.recipes.List(r.Context())
	if err != nil {
		log.Printf("listing recipes: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	response := make([]recipeSummaryResponse, 0, len(recipes))
	for _, recipe := range recipes {
		response = append(response, newRecipeSummaryResponse(recipe))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *RecipeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	recipe, err := h.recipes.GetByID(r.Context(), id)
	if err != nil {
		writeRepositoryError(w, "getting recipe", err)
		return
	}
	writeJSON(w, http.StatusOK, newRecipeResponse(recipe))
}

func (h *RecipeHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := authenticatedUser(r)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	request, ok := decodeRecipeRequest(w, r)
	if !ok {
		return
	}
	recipe, validationCode := request.newRecipe()
	if validationCode != "" {
		writeError(w, http.StatusBadRequest, validationCode)
		return
	}
	recipe.UserID = user.ID

	created, err := h.recipes.Create(r.Context(), recipe)
	if err != nil {
		log.Printf("creating recipe: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	w.Header().Set("Location", "/api/recipes/"+created.ID)
	writeJSON(w, http.StatusCreated, newRecipeResponse(created))
}

func (h *RecipeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user, ok := authenticatedUser(r)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	existing, err := h.recipes.GetByID(r.Context(), id)
	if err != nil {
		writeRepositoryError(w, "loading recipe for update", err)
		return
	}
	if existing.UserID != user.ID && user.Role != "admin" {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	request, ok := decodeRecipeRequest(w, r)
	if !ok {
		return
	}
	recipe, validationCode := request.newRecipe()
	if validationCode != "" {
		writeError(w, http.StatusBadRequest, validationCode)
		return
	}
	recipe.ID = id

	updated, err := h.recipes.Update(r.Context(), recipe)
	if err != nil {
		writeRepositoryError(w, "updating recipe", err)
		return
	}
	writeJSON(w, http.StatusOK, newRecipeResponse(updated))
}

func (h *RecipeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user, ok := authenticatedUser(r)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	existing, err := h.recipes.GetByID(r.Context(), id)
	if err != nil {
		writeRepositoryError(w, "loading recipe for delete", err)
		return
	}
	if existing.UserID != user.ID && user.Role != "admin" {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := h.recipes.Delete(r.Context(), id); err != nil {
		writeRepositoryError(w, "deleting recipe", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeRecipeRequest(w http.ResponseWriter, r *http.Request) (recipeRequest, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRecipeRequestBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var request recipeRequest
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return recipeRequest{}, false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return recipeRequest{}, false
	}
	return request, true
}

func (request recipeRequest) newRecipe() (*repository.Recipe, string) {
	request.Name = strings.TrimSpace(request.Name)
	request.RecipeType = strings.TrimSpace(request.RecipeType)
	request.TimeToCook = strings.TrimSpace(request.TimeToCook)
	request.Description = strings.TrimSpace(request.Description)

	if request.Name == "" || len(request.Name) > 200 {
		return nil, "invalid_name"
	}
	if request.RecipeType != "structured" {
		if request.RecipeType == "image" {
			return nil, "image_upload_not_supported"
		}
		return nil, "invalid_recipe_type"
	}
	if len(request.TimeToCook) > 100 || len(request.Description) > 10_000 {
		return nil, "invalid_recipe"
	}
	if len(request.Ingredients) == 0 || len(request.Ingredients) > 200 {
		return nil, "invalid_ingredients"
	}
	if len(request.Steps) == 0 || len(request.Steps) > 100 {
		return nil, "invalid_steps"
	}

	recipe := &repository.Recipe{
		Name:        request.Name,
		RecipeType:  request.RecipeType,
		TimeToCook:  request.TimeToCook,
		Description: request.Description,
		Ingredients: make([]repository.Ingredient, len(request.Ingredients)),
		Steps:       make([]repository.RecipeStep, len(request.Steps)),
		Images:      make([]repository.RecipeImage, 0),
	}
	for index, ingredient := range request.Ingredients {
		name := strings.TrimSpace(ingredient.Name)
		quantity := strings.TrimSpace(ingredient.Quantity)
		if name == "" || len(name) > 200 || len(quantity) > 100 {
			return nil, "invalid_ingredients"
		}
		recipe.Ingredients[index] = repository.Ingredient{Name: name, Quantity: quantity, Position: index}
	}
	for index, step := range request.Steps {
		instruction := strings.TrimSpace(step.Instruction)
		if instruction == "" || len(instruction) > 5_000 {
			return nil, "invalid_steps"
		}
		recipe.Steps[index] = repository.RecipeStep{StepNumber: index + 1, Instruction: instruction}
	}
	return recipe, ""
}

func newRecipeResponse(recipe *repository.Recipe) recipeResponse {
	response := recipeResponse{
		ID:           recipe.ID,
		Name:         recipe.Name,
		RecipeType:   recipe.RecipeType,
		TimeToCook:   recipe.TimeToCook,
		Description:  recipe.Description,
		UserID:       recipe.UserID,
		DatePosted:   recipe.DatePosted,
		LastEditedAt: recipe.LastEditedAt,
		Ingredients:  make([]ingredientResponse, 0, len(recipe.Ingredients)),
		Steps:        make([]stepResponse, 0, len(recipe.Steps)),
		Images:       make([]imageResponse, 0, len(recipe.Images)),
	}
	for _, ingredient := range recipe.Ingredients {
		response.Ingredients = append(response.Ingredients, ingredientResponse{
			ID: ingredient.ID, Name: ingredient.Name, Quantity: ingredient.Quantity, Position: ingredient.Position,
		})
	}
	for _, step := range recipe.Steps {
		response.Steps = append(response.Steps, stepResponse{
			ID: step.ID, StepNumber: step.StepNumber, Instruction: step.Instruction,
		})
	}
	for _, image := range recipe.Images {
		response.Images = append(response.Images, imageResponse{
			ID: image.ID, FileName: image.FileName, ContentType: image.ContentType, FileSize: image.FileSize,
			Position: image.Position, IsCover: image.IsCover, UploadedAt: image.UploadedAt,
		})
	}
	return response
}

func newRecipeSummaryResponse(recipe *repository.Recipe) recipeSummaryResponse {
	return recipeSummaryResponse{
		ID: recipe.ID, Name: recipe.Name, RecipeType: recipe.RecipeType,
		TimeToCook: recipe.TimeToCook, Description: recipe.Description, UserID: recipe.UserID,
		DatePosted: recipe.DatePosted, LastEditedAt: recipe.LastEditedAt,
	}
}

func writeRepositoryError(w http.ResponseWriter, action string, err error) {
	if errors.Is(err, apperror.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	log.Printf("%s: %v", action, err)
	writeError(w, http.StatusInternalServerError, "internal_error")
}
