// Package service is used to add business logic between our router and db interactions
package service

import (
	"context"
	"fmt"
	"recipe-website/internal/domain"
	"recipe-website/internal/repository"
)

type RecipeService struct {
	recipeRepo *repository.PostgresRecipeRepository
}

func NewRecipeService(rr *repository.PostgresRecipeRepository) *RecipeService {
	return &RecipeService{
		recipeRepo: rr,
	}
}

func (s *RecipeService) validateRecipe(params domain.CreateRecipeRequest) error {
	if params.Name == "" {
		return fmt.Errorf("recipe name is required")
	}

	if len(params.Name) > 100 {
		return fmt.Errorf("recipe name cannot be longer than 100 characters")
	}

	if params.UserID == "" {
		return fmt.Errorf("user id is required to create a recipe")
	}

	return nil
}

func (s *RecipeService) validateUpdateRecipe(params domain.UpdateRecipeRequest) error {
	if len(*params.Name) > 100 {
		return fmt.Errorf("recipe name cannot be longer than 100 characters")
	}

	if params.RecipeID == "" {
		return fmt.Errorf("recipe id is required for updates")
	}

	return nil
}

func (s *RecipeService) Create(ctx context.Context, req domain.CreateRecipeRequest) (*domain.Recipe, error) {
	if err := s.validateRecipe(req); err != nil {
		return nil, fmt.Errorf("recipe create validation failed: %w", err)
	}

	recipe, err := s.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create recipe: %w", err)
	}

	return recipe, nil
}

func (s *RecipeService) GetByID(ctx context.Context, recipeID string) (*domain.Recipe, error) {
	if recipeID == "" {
		return nil, fmt.Errorf("recipe id is required, GetByID")
	}

	recipe, err := s.GetByID(ctx, recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recipe: %w", err)
	}

	return recipe, nil
}
