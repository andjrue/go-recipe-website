package repository

import (
	"context"
	"recipe-website/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func (r *PostgresUserRepository) GetByPK(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return nil, nil 
}

func (r *PostgresUserRepository) Create(ctx context.Context, req *domain.CreateUserRequest)	(*domain.User, error) {
	return nil, nil
}