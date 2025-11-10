package service

import "recipe-website/internal/repository"

type Services struct {
	Recipe *RecipeService
	User   *UserService
}

func SetServices(repos *repository.Repositories) *Services {
	return &Services{
		Recipe: NewRecipeService(repos.Recipe),
		User:   NewUserService(repos.User),
	}
}
