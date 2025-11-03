package service

import (
	"recipe-website/internal/domain"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestValidateRecipe(t *testing.T) {
	s := NewRecipeService(nil)
	baseReq := func() domain.CreateRecipeRequest {
		baseTime := 45
		return domain.CreateRecipeRequest{
			Name:       "Base Recipe",
			TimeToCook: &baseTime,
			UserID:     uuid.New(),
		}
	}

	t.Run("Happy Path", func(t *testing.T) {
		req := baseReq()

		err := s.validateRecipe(req)
		if err != nil {
			t.Errorf("expected no error, but got: %v", err)
		}
	})

	t.Run("Invalid Name", func(t *testing.T) {
		req := baseReq()
		req.Name = ""

		err := s.validateRecipe(req)
		if err == nil {
			t.Errorf("expected error for invalid name, did not receive one")
		}
	})

	t.Run("Nil Time To Cook", func(t *testing.T) {
		req := baseReq()
		req.TimeToCook = nil

		err := s.validateRecipe(req)
		if err != nil {
			t.Errorf("expected no error for nil time to cook, but got: %v", err)
		}
	})

	t.Run("Name Length", func(t *testing.T) {
		req := baseReq()
		req.Name = ""

		char := "a"
		repeat := 200
		req.Name = strings.Repeat(char, repeat)

		err := s.validateRecipe(req)
		if err == nil {
			t.Errorf("expected error for name length check, did not receive one")
		}
	})

	t.Run("Name Length", func(t *testing.T) {
		req := baseReq()
		req.UserID = uuid.Nil

		err := s.validateRecipe(req)
		if err == nil {
			t.Errorf("expected error for user id check, did not receive one")
		}
	})
}

