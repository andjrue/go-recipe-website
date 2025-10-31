// Package domain houses structs for entities and req's
package domain

import "time"

type Recipe struct {
	RecipeID     string
	Name         string
	TimeToCook   int
	Description  string
	UserID       string
	DatePosted   time.Time
	Deleted      bool
	LastEditedAt time.Time
}

type CreateRecipeRequest struct {
	Name        string
	TimeToCook  *int
	Description *string
	UserID      string
}

type UpdateRecipeRequest struct {
	Name        *string
	TimeToCook  *int
	Description *string
	RecipeID    string
}
