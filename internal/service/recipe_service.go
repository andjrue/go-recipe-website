// Package service is used to add business logic between our router and db interactions
package service

import (
	"context"
	"fmt"
	"recipe-website/internal/domain"
	"recipe-website/internal/repository"

	"github.com/google/uuid"
)

type RecipeService struct {
	recipeRepo repository.RecipeRepository
}

func NewRecipeService(recipeRepo repository.RecipeRepository) *RecipeService {
	return &RecipeService{
		recipeRepo: recipeRepo,
	}
}

func (s *RecipeService) validateRecipe(params domain.CreateRecipeRequest) error {

	if params.UserID == uuid.Nil {
		return fmt.Errorf("user id is required")
	}

	if params.Name == "" {
		return fmt.Errorf("recipe name is required")
	}

	if len(params.Name) > 100 {
		return fmt.Errorf("recipe name cannot be longer than 100 characters")
	}

	if params.TimeToCook != nil && *params.TimeToCook <= 0 {
		return fmt.Errorf("time to cook must be greater than 0")
	}

	return nil
}

func (s *RecipeService) validateUpdateRecipe(params domain.UpdateRecipeRequest, recipe *domain.Recipe) error {
	if len(params.Name) > 100 {
		return fmt.Errorf("recipe name cannot be longer than 100 characters")
	}

	if params.RecipeID == uuid.Nil {
		return fmt.Errorf("recipe id is required")
	}

	if params.UserID != recipe.UserID {
		return fmt.Errorf("user id must match existing recipe")
	}

	return nil
}

func (s *RecipeService) Create(ctx context.Context, req domain.CreateRecipeRequest) (*domain.Recipe, error) {
	if err := s.validateRecipe(req); err != nil {
		return nil, fmt.Errorf("recipe create validation failed: %w", err)
	}

	recipe, err := s.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("[recipe service] - failed to create recipe: %w", err)
	}

	return recipe, nil
}

func (s *RecipeService) GetByPK(ctx context.Context, recipeID uuid.UUID) (*domain.Recipe, error) {
	if err := uuid.Validate(recipeID.String()); err != nil {
		return nil, fmt.Errorf("recipe id malformed, %v", err)
	}

	recipe, err := s.GetByPK(ctx, recipeID)
	if err != nil {
		return nil, fmt.Errorf("[recipe service] - failed to fetch recipe: %w", err)
	}

	return recipe, nil
}

func (s *RecipeService) GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Recipe, error) {
	if err := uuid.Validate(userID.String()); err != nil {
		return nil, fmt.Errorf("recipe id is required, GetByID")
	}

	recipes, err := s.GetAllByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("[recipe service] - unable to fetch all for user: %v", err)
	}

	return recipes, nil
}

func (s *RecipeService) Update(ctx context.Context, req *domain.UpdateRecipeRequest) (*domain.Recipe, error) {
	currentRecipe, err := s.GetByPK(ctx, req.RecipeID)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch recipe to update: %w", err)
	}

	if err := s.validateUpdateRecipe(*req, currentRecipe); err != nil {
		return nil, fmt.Errorf("cannot update recipe: %v", err)
	}

	update, err := s.Update(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("[recipe service] - unable to update recipe: %v", err)
	}

	return update, nil
}

func (s *RecipeService) Delete(ctx context.Context, recipeID uuid.UUID) error {
	if err := uuid.Validate(recipeID.String()); err != nil {
		return fmt.Errorf("recipe ID is invalid: %v", err)
	}

	if err := s.Delete(ctx, recipeID); err != nil {
		return fmt.Errorf("error deleting recipe: %v", err)
	}

	return nil
}
