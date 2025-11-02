// Package domain houses structs for entities and req's
package domain

import (
	"time"

	"github.com/google/uuid"
)

// TODO: Add db tags to domain objs
type Recipe struct {
	RecipeID     uuid.UUID `json:"recipe_id"`
	UserID       uuid.UUID `json:"user_id"`
	Name         string    `json:"name"`
	TimeToCook   *int      `json:"time_to_cook"`
	Description  *string   `json:"decription"`
	DatePosted   time.Time `json:"date_posted"`
	Deleted      bool      `json:"deleted"`
	LastEditedAt time.Time `json:"last_edited_at"`
}

type CreateRecipeRequest struct {
	Name        string    `json:"name"`
	TimeToCook  *int      `json:"time_to_cook"`
	Description *string   `json:"decription"`
	UserID      uuid.UUID `json:"user_id"`
}

type UpdateRecipeRequest struct {
	Name        string    `json:"name"`
	TimeToCook  *int      `json:"time_to_cook"`
	Description *string   `json:"decription"`
	RecipeID    uuid.UUID `json:"recipe_id"`
}
