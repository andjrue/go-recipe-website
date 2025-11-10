package handler

import (
	"encoding/json"
	"net/http"
	"recipe-website/internal/domain"
	"recipe-website/internal/service"
)

type UserHandler struct {
	us *service.UserService
}

func NewUserHandler(us *service.UserService) *UserHandler {
	return &UserHandler{
		us: us,
	}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateUserRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, domain.ErrValidation)
		return
	}
	
	user, err := h.us.CreateUser(r.Context(), &req)
	if err != nil {
		respondError(w, err)
		return
	}
	
	respondJSON(w, http.StatusCreated, user)
}