package database

import (
	"context"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrate applies every pending migration packaged into the API binary.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("setting migration dialect: %w", err)
	}
	goose.SetBaseFS(migrationFiles)

	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()
	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("applying database migrations: %w", err)
	}
	return nil
}
