package service

import "recipe-website/internal/repository"

type Services struct {
	Recipe *RecipeService
}

func SetServices(repos *repository.Repositories) *Services {
	return &Services{
		Recipe: NewRecipeService(repos.Recipe),
	}
}
