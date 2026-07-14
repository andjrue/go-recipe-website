package repository

import (
	"errors"

	"recipe-website/internal/apperror"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func translatePostgresError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apperror.ErrConflict
	}

	return err
}

// translatePostgresLookupError treats a malformed UUID lookup like any other
// identifier that does not match a row. PostgreSQL reports it before running
// the query because UUID columns cannot be compared with arbitrary text.
func translatePostgresLookupError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "22P02" {
		return apperror.ErrNotFound
	}
	return translatePostgresError(err)
}
