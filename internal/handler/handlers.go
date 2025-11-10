package handler

import "recipe-website/internal/service"

type Handlers struct {
	Recipe *RecipeHandler
	User *UserHandler
}

func NewHandler(service *service.Services) *Handlers {
	return &Handlers{
		Recipe: NewRecipeHandler(service.Recipe),
		User: NewUserHandler(service.User),
	}
}

// TODO - add handlers
