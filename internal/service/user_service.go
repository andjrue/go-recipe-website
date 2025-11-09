package service

import (
	"context"
	"fmt"
	"recipe-website/internal/domain"
	"recipe-website/internal/repository"
	"unicode"
)

type UserService struct {
	ur repository.UserRepository
}

func NewUserService(ur repository.UserRepository) *UserService {
	return &UserService{
		ur: ur,
	}
}

func (s *UserService) validateAlias(a string) error {
	
	if a == "" {
		return fmt.Errorf("alias is required for user creation")
	}
	
	if len(a) > 25 || len(a) < 4 {
		return fmt.Errorf("alias must be between 4 and 25 characters")
	}
	
	for _, r := range a {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return fmt.Errorf("special characters are not allowed in alias")
		}
	}
	
	return nil
}

func (s *UserService) validateCreateUser(params domain.CreateUserRequest) error {
	if err := s.validateAlias(params.Alias); err != nil {
		return fmt.Errorf("alias invalid: %w", err)
	}
	
	if params.Email == "" {
		return fmt.Errorf("email required for user creation")
	}
	
	if params.Password == "" {
		return fmt.Errorf("password required for user creation")
	}
	
	return nil
}

func (s *UserService) CreateUser(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error) {
	if err := s.validateCreateUser(*req); err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}
	
	recipe, err := s.ur.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	return recipe, nil
}