package handler

import "recipe-website/internal/service"

type Handlers struct {
	Recipe *RecipeHandler
}

func NewHandler(service *service.Services) *Handlers {
	return &Handlers{
		Recipe: NewRecipeHandler(service.Recipe),
	}
}

// TODO - add handlers