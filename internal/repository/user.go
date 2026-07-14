package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository is the persistence seam for users. Handlers depend on this
// interface so the DB can be swapped for a mock in tests.
type UserRepository interface {
	// Create inserts a user and returns it populated with DB-generated fields
	// (ID, Role, DateJoined).
	Create(ctx context.Context, u *User) (*User, error)

	// GetByID looks up a single user by primary key.
	GetByID(ctx context.Context, id string) (*User, error)

	// GetByProviderUserID looks up a user by their external identity — the
	// lookup we do on every OAuth login to find-or-create the account.
	GetByProviderUserID(ctx context.Context, provider, providerUserID string) (*User, error)
}

// UserPostgres is the pgx-backed implementation of UserRepository.
type UserPostgres struct {
	db *pgxpool.Pool
}

// Compile-time check that we actually satisfy the interface.
var _ UserRepository = (*UserPostgres)(nil)

func NewUserPostgres(db *pgxpool.Pool) *UserPostgres {
	return &UserPostgres{db: db}
}

// Create inserts a user and returns the DB-populated fields so the *User comes
// back fully populated.
func (r *UserPostgres) Create(ctx context.Context, u *User) (*User, error) {
	if u.Provider == "" {
		u.Provider = "google"
	}

	const query = `
		INSERT INTO users (email, provider, provider_user_id, alias)
		VALUES ($1, $2, $3, $4)
		RETURNING user_id, provider, role, date_joined`

	err := r.db.QueryRow(ctx, query, u.Email, u.Provider, u.ProviderUserID, u.Alias).Scan(
		&u.ID,
		&u.Provider,
		&u.Role,
		&u.DateJoined,
	)
	if err != nil {
		return nil, translatePostgresError(err)
	}
	return u, nil
}

// GetByID looks up a single user by primary key.
func (r *UserPostgres) GetByID(ctx context.Context, id string) (*User, error) {
	const query = `
		SELECT user_id, email, provider, provider_user_id, alias, role, date_joined
		FROM users
		WHERE user_id = $1`

	var u User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
		&u.Provider,
		&u.ProviderUserID,
		&u.Alias,
		&u.Role,
		&u.DateJoined,
	)
	if err != nil {
		return nil, translatePostgresLookupError(err)
	}
	return &u, nil
}

// GetByProviderUserID looks up the account tied to an external identity — the
// (provider, provider_user_id) unique pair we resolve on every OAuth login.
func (r *UserPostgres) GetByProviderUserID(ctx context.Context, provider, providerUserID string) (*User, error) {
	const query = `
		SELECT user_id, email, provider, provider_user_id, alias, role, date_joined
		FROM users
		WHERE provider = $1 AND provider_user_id = $2`

	var u User
	err := r.db.QueryRow(ctx, query, provider, providerUserID).Scan(
		&u.ID,
		&u.Email,
		&u.Provider,
		&u.ProviderUserID,
		&u.Alias,
		&u.Role,
		&u.DateJoined,
	)
	if err != nil {
		return nil, translatePostgresError(err)
	}
	return &u, nil
}
