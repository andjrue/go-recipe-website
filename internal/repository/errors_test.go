package repository

import (
	"errors"
	"testing"

	"recipe-website/internal/apperror"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestTranslatePostgresError(t *testing.T) {
	otherErr := errors.New("boom")

	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "nil",
			err:  nil,
			want: nil,
		},
		{
			name: "no rows",
			err:  pgx.ErrNoRows,
			want: apperror.ErrNotFound,
		},
		{
			name: "unique violation",
			err:  &pgconn.PgError{Code: "23505"},
			want: apperror.ErrConflict,
		},
		{
			name: "other error",
			err:  otherErr,
			want: otherErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translatePostgresError(tt.err)
			if tt.want == nil {
				if got != nil {
					t.Fatalf("translatePostgresError() = %v, want nil", got)
				}
				return
			}

			if !errors.Is(got, tt.want) {
				t.Fatalf("translatePostgresError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTranslatePostgresLookupError(t *testing.T) {
	invalidUUID := &pgconn.PgError{Code: "22P02"}
	if got := translatePostgresLookupError(invalidUUID); !errors.Is(got, apperror.ErrNotFound) {
		t.Fatalf("translatePostgresLookupError() = %v, want not found", got)
	}
}
