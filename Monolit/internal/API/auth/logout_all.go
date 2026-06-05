package auth

import (
	"calllens/monolit/internal/httpserver/middleware"
	"net/http"
)

func (h *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.service.LogoutAll(r.Context(), userID); err != nil {
		http.Error(w, "failed to logout all", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
