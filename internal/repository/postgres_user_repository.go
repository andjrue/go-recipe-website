package repository

import (
	"context"
	"fmt"
	"recipe-website/internal/domain"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func (r *PostgresUserRepository) scanUser(row pgx.Row) (*domain.User, error) {
	var user domain.User
	
	err := row.Scan(
		&user.UserID,
		&user.HashedPassword,
		&user.Email,
		&user.Alias,
		&user.DateJoined,
	)
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

func (r *PostgresUserRepository) hashPassword(p string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(bytes), err
}

func (r *PostgresUserRepository) GetByPK(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	q := `SELECT * 
	FROM users
	WHERE user_id = $1
	`
	
	row := r.db.QueryRow(ctx, q, userID)
	user, err := r.scanUser(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	
	return user, nil	
}

func (r *PostgresUserRepository) Create(ctx context.Context, req *domain.CreateUserRequest)	(*domain.User, error) {
	hashedPassword, err := r.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("unable to hash pass: %w", err)
	}
	
	q := `INSERT INTO users (user_id, hashed_password, email, alias, date_joined)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING user_id, hashed_password, email, alias, date_joined`
	
	now := time.Now()
	id := uuid.New()
	
	row:= r.db.QueryRow(ctx, q, 
		id,
		hashedPassword,
		req.Email,
		req.Alias,
		now,
	)

	user, err := r.scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("unable to scan created user: %w", err)
	}
	
	return user, nil 
}