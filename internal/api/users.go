package api

import (
	"errors"
	"net/http"
	"time"

	"recipe-website/internal/apperror"
	"recipe-website/internal/repository"
)

type UserHandler struct {
	users repository.UserRepository
}

type userResponse struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	Alias      string    `json:"alias"`
	Role       string    `json:"role"`
	DateJoined time.Time `json:"dateJoined"`
}

type userSummaryResponse struct {
	ID         string    `json:"id"`
	Alias      string    `json:"alias"`
	DateJoined time.Time `json:"dateJoined"`
}

func NewUserHandler(users repository.UserRepository) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	user, err := h.users.GetByID(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found")
			return
		}

		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, newUserSummaryResponse(user))
}

func newUserResponse(user *repository.User) userResponse {
	return userResponse{
		ID:         user.ID,
		Email:      user.Email,
		Alias:      user.Alias,
		Role:       user.Role,
		DateJoined: user.DateJoined,
	}
}

func newUserSummaryResponse(user *repository.User) userSummaryResponse {
	return userSummaryResponse{
		ID:         user.ID,
		Alias:      user.Alias,
		DateJoined: user.DateJoined,
	}
}
