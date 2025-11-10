package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"recipe-website/internal/domain"
)

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// TODO: Set up custom error handling for this. Is and As can get us a long way.
func respondError(w http.ResponseWriter, err error) {
	var status int
	var message string

	switch {
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
		message = domain.ErrNotFound.Error()
	case errors.Is(err, domain.ErrValidation):
		status = http.StatusBadRequest
		message = err.Error() // This will be slightly different, we want to return full messages when we have bad requests
	case errors.Is(err, domain.ErrDuplicateEntry):
		status = http.StatusConflict
		message = domain.ErrDuplicateEntry.Error()
	default:
		status = http.StatusInternalServerError
		message = domain.ErrUnexpected.Error()
		log.Printf("unexpected error: %w", err)
	}

	respondJSON(w, status, map[string]string{"error": message})
}
