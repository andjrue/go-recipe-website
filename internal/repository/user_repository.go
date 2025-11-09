package repository

import (
	"context"
	"recipe-website/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	GetByPK(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	Create(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error)
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}