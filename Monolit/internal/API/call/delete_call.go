package call

import (
	"calllens/monolit/internal/models"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *CallHandler) DeleteCall(w http.ResponseWriter, r *http.Request) {
	rawUUID := chi.URLParam(r, "uuid")

	callUUID, err := uuid.Parse(rawUUID)
	if err != nil {
		http.Error(w, "invalid call uuid", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteCall(r.Context(), callUUID); err != nil {
		if errors.Is(err, models.ErrCallNotFound) {
			http.Error(w, "call not found", http.StatusNotFound)
			return
		}
		http.Error(w, "delete call failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
