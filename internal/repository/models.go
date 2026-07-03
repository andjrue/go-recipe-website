// Package repository holds domain models and the interfaces + Postgres
// implementations for persisting them. Handlers depend on the interfaces,
// never on the concrete DB types.
package repository

import "time"

// User is an account backed by an external identity provider (Google for now).
// We store no passwords — the provider owns authentication.
type User struct {
	ID             string
	Email          string
	Provider       string
	ProviderUserID string
	Alias          string
	Role           string
	DateJoined     time.Time
}
