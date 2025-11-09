package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserID uuid.UUID `json:"user_id"`
	HashedPassword string `json:"-"`
	Email string `json:"email"`
	Alias string `json:"alias"`
	DateJoined time.Time `json:"date_joined"`
}

type CreateUserRequest struct {
	Password string `json:"password"`
	Email string `json:"email"`
	Alias string `json:"alias"`
}

/**
 * Will need:
 * Reset password
 * Delete user - This will require DB cascading for recipes, comments, etc. 
 * 	-> We could also have it like reddit, "comment posted by deleted user" or something like that. 
 * 
 * That might be it? We can at least get started now. 
 */
