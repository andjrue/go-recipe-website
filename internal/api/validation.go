package api

import "net/http"

func isUUID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return false
	}

	for index, character := range value {
		if index == 8 || index == 13 || index == 18 || index == 23 {
			continue
		}
		if !((character >= '0' && character <= '9') ||
			(character >= 'a' && character <= 'f') ||
			(character >= 'A' && character <= 'F')) {
			return false
		}
	}

	return true
}

func recipeIDFromRequest(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := r.PathValue("id")
	if !isUUID(id) {
		writeError(w, http.StatusBadRequest, "invalid_id")
		return "", false
	}
	return id, true
}
