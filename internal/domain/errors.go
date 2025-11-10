package domain

import "errors"

var (
	ErrValidation = errors.New("input validation failed")
	ErrNotFound = errors.New("resource not found")
	ErrDuplicateEntry = errors.New("resource already exists")
	ErrUnexpected = errors.New("an unexpected error occurred")
)