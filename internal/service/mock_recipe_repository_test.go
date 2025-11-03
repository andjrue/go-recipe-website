package service

import (
	"context"
	"recipe-website/internal/domain"
	"recipe-website/internal/repository"
	"testing"
	"time"

	"github.com/google/uuid"
)

type MockRecipeRepository struct {
	GetByPKFn        func(ctx context.Context, recipeID uuid.UUID) (*domain.Recipe, error)
	GetAllByUserIDFn func(ctx context.Context, userID uuid.UUID) ([]domain.Recipe, error)
	CreateFn         func(ctx context.Context, recipe *domain.CreateRecipeRequest) (*domain.Recipe, error)
	UpdateFn         func(ctx context.Context, req *domain.UpdateRecipeRequest) error
	DeleteFn         func(ctx context.Context, recipeID uuid.UUID) error
}

var _ repository.RecipeRepository = (*MockRecipeRepository)(nil)

func (m *MockRecipeRepository) GetByPK(ctx context.Context, id uuid.UUID) (*domain.Recipe, error) {
	return m.GetByPKFn(ctx, id)
}

func (m *MockRecipeRepository) GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Recipe, error) {
	return m.GetAllByUserIDFn(ctx, userID)
}

func (m *MockRecipeRepository) Create(ctx context.Context, recipe *domain.CreateRecipeRequest) (*domain.Recipe, error) {
	return m.CreateFn(ctx, recipe)
}

func (m *MockRecipeRepository) Update(ctx context.Context, req *domain.UpdateRecipeRequest) error {
	return m.UpdateFn(ctx, req)
}

func (m *MockRecipeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFn(ctx, id)
}

func TestRecipeService_Create_Success(t *testing.T) {
	expectedRecipeID := uuid.New()
	testUserID := uuid.New()
	testDescription := "Base Test Creation Description"
	testCookTime := 45

	createReq := &domain.CreateRecipeRequest{
		Name:        "Base Test Creation Name",
		UserID:      testUserID,
		Description: &testDescription,
		TimeToCook:  &testCookTime,
	}

	mockRepo := &MockRecipeRepository{
		CreateFn: func(ctx context.Context, recipe *domain.CreateRecipeRequest) (*domain.Recipe, error) {
			return &domain.Recipe{
				RecipeID:     expectedRecipeID,
				UserID:       testUserID,
				Name:         createReq.Name,
				TimeToCook:   createReq.TimeToCook,
				Description:  createReq.Description,
				DatePosted:   time.Now(),
				Deleted:      false,
				LastEditedAt: time.Now(),
			}, nil
		},
	}

	service := NewRecipeService(mockRepo)
	createdRecipe, err := service.Create(context.Background(), createReq)

	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}

	if createdRecipe.RecipeID != expectedRecipeID {
		t.Errorf("expected recipe id: %s\n received recipe id: %s", expectedRecipeID, createdRecipe.RecipeID)
	}

	if createdRecipe.UserID != testUserID {
		t.Errorf("expected user id: %s\n received recipe id: %s", testUserID, createdRecipe.UserID)
	}

	if createdRecipe.Name != createReq.Name {
		t.Errorf("expected recipe name: %s\n received recipe name: %s", createReq.Name, createdRecipe.Name)
	}

	if createdRecipe.Description != createReq.Description {
		t.Errorf("expected recipe description: %s\n received recipe description: %s", *createReq.Description, *createdRecipe.Description)
	}

	if createdRecipe.TimeToCook != createReq.TimeToCook {
		t.Errorf("expected recipe time to cook: %v\n received recipe time to cook: %v", *createReq.TimeToCook, *createdRecipe.TimeToCook)
	}
}
