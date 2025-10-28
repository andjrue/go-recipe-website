// Package repository provides repository layers for entites
package repository

import (
	"context"
	"recipe-website/internal/domain"
)

type RecipeFetcher interface {
	GetByPK(ctx context.Context, recipeID string) (*domain.Recipe, error)
	GetAllByUserID(ctx context.Context, userID string) ([]domain.Recipe, error)
	Create(ctx context.Context, recipe *domain.CreateRecipeRequest) (*domain.Recipe, error)
	Update(ctx context.Context, recipeID string, req *domain.UpdateRecipeRequest) (*domain.Recipe, error)
	Delete(ctx context.Context, recipeID string) error
}
