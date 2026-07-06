// Package apperror defines shared application-level errors.
package apperror

import "errors"

var (
	// ErrNotFound is returned when a requested resource does not exist.
	ErrNotFound = errors.New("not found")
	// ErrConflict is returned when a write conflicts with existing state.
	ErrConflict = errors.New("conflict")
)
