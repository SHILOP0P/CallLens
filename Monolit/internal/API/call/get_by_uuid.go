package call

import (
	"calllens/monolit/internal/converter"
	"calllens/monolit/internal/models"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *CallHandler) GetByUUID(w http.ResponseWriter, r *http.Request) {
	rawUUID := chi.URLParam(r, "uuid")

	callUUID, err := uuid.Parse(rawUUID)
	if err != nil {
		http.Error(w, "invalid call uuid", http.StatusBadRequest)
		return
	}

	call, err := h.service.GetByUUID(r.Context(), callUUID)
	if err != nil {
		if errors.Is(err, models.ErrCallNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "failed to find call", http.StatusInternalServerError)
		return
	}

	response, err := converter.CallModelToAPI(call)
	if err != nil {
		http.Error(w, "failed to convert call", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
