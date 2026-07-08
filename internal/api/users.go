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
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	Provider       string    `json:"provider"`
	ProviderUserID string    `json:"providerUserID"`
	Alias          string    `json:"alias"`
	Role           string    `json:"role"`
	DateJoined     time.Time `json:"dateJoined"`
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

	writeJSON(w, http.StatusOK, newUserResponse(user))
}

func newUserResponse(user *repository.User) userResponse {
	return userResponse{
		ID:             user.ID,
		Email:          user.Email,
		Provider:       user.Provider,
		ProviderUserID: user.ProviderUserID,
		Alias:          user.Alias,
		Role:           user.Role,
		DateJoined:     user.DateJoined,
	}
}
