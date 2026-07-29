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

// Recipe is a cookbook entry. Structured recipes contain ingredients and
// steps; image recipes contain normalized image metadata.
type Recipe struct {
	ID           string
	Name         string
	RecipeType   string
	TimeToCook   string
	Description  string
	UserID       string
	DatePosted   time.Time
	LastEditedAt time.Time
	CoverImageID string
	Ingredients  []Ingredient
	Steps        []RecipeStep
	Images       []RecipeImage
}

type Ingredient struct {
	ID       string
	RecipeID string
	Name     string
	Quantity string
	Position int
}

type RecipeStep struct {
	ID          string
	RecipeID    string
	StepNumber  int
	Instruction string
}

type RecipeImage struct {
	ID          string
	RecipeID    string
	S3Key       string
	FileName    string
	ContentType string
	FileSize    int64
	Position    int
	IsCover     bool
	UploadedAt  time.Time
}
