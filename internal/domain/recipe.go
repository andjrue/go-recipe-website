// Package domain houses structs for entities and req's
package domain

import (
	"time"

	"github.com/google/uuid"
)

type Recipe struct {
	RecipeID     uuid.UUID
	UserID       uuid.UUID
	Name         string
	TimeToCook   int
	Description  string
	DatePosted   time.Time
	Deleted      bool
	LastEditedAt time.Time
}

type CreateRecipeRequest struct {
	Name        string
	TimeToCook  *int
	Description *string
	UserID      uuid.UUID
}

type UpdateRecipeRequest struct {
	Name        *string
	TimeToCook  *int
	Description *string
	RecipeID    uuid.UUID
}
