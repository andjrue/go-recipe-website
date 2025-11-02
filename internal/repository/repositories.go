package repository

import (

	"github.com/jackc/pgx/v5"
)

type Repositories struct {
	Recipe RecipeRepository
}

func SetRepositories(db *pgx.Conn) *Repositories {
	return &Repositories{
		Recipe: NewRecipeRepository(db),
	}
}