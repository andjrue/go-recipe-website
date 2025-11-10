// Package repository provides repository layers for entites
package repository

import (
	"context"
	"recipe-website/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RecipeRepository interface {
	GetByPK(ctx context.Context, recipeID uuid.UUID) (*domain.Recipe, error)
	GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Recipe, error)
	Create(ctx context.Context, recipe *domain.CreateRecipeRequest) (*domain.Recipe, error)
	Update(ctx context.Context, req *domain.UpdateRecipeRequest) error
	Delete(ctx context.Context, recipeID uuid.UUID) error
	// recover -- we only soft delete, so recipes will be recoverable
	// get all deleted -- need a menu to see all deleted recipes by a user
	// perm delete -- remove recipe from db
}

func NewRecipeRepository(db *pgxpool.Pool) RecipeRepository {
	return &PostgresRecipeRepository{
		db: db,
	}
}
