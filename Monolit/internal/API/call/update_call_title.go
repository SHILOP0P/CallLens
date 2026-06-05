package call

import (
	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/converter"
	"calllens/monolit/internal/models"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *CallHandler) UpdateCallTitle(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	rawUUID := chi.URLParam(r, "uuid")

	callUUID, err := uuid.Parse(rawUUID)
	if err != nil {
		http.Error(w, "invalid call uuid", http.StatusBadRequest)
		return
	}

	var req dto.UpdateCallTitleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	updatedCall, err := h.service.UpdateCallTitle(r.Context(), callUUID, userID, req.Title)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCallTitle) {
			http.Error(w, "invalid call title", http.StatusBadRequest)
			return
		}
		if errors.Is(err, models.ErrCallNotFound) {
			http.Error(w, "call not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update call title", http.StatusInternalServerError)
		return
	}

	response, err := converter.CallModelToAPI(updatedCall)
	if err != nil {
		http.Error(w, "failed to convert call", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}

}
