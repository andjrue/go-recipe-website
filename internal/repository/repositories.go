package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repositories struct {
	Recipe RecipeRepository
	User UserRepository
}

func SetRepositories(db *pgxpool.Pool) *Repositories {
	return &Repositories{
		Recipe: NewRecipeRepository(db),
		User: NewUserRepository(db),
	}
}
