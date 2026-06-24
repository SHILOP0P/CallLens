package auth

import (
	"net/http"

	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/httpserver/middleware"
)

func (h *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, response.CodeUnauthorized, "unauthorized")
		return
	}

	if err := h.service.LogoutAll(r.Context(), userID); err != nil {
		response.WriteError(w, http.StatusInternalServerError, response.CodeFailedToLogoutAll, "failed to logout all")
		return
	}

	h.clearAuthCookies(w, r)
	response.WriteNoContent(w)
}
