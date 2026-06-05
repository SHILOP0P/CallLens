package auth

import (
	"calllens/monolit/internal/httpserver/middleware"
	model "calllens/monolit/internal/models"
	"errors"
	"net/http"
)

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := middleware.SessionIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.service.Logout(r.Context(), sessionID); err != nil {
		if errors.Is(err, model.ErrRefreshSessionNotFound) {
			http.Error(w, "session not found", http.StatusUnauthorized)
			return
		}

		http.Error(w, "failed to logout", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
