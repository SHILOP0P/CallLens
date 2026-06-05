package auth

import (
	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/converter"
	model "calllens/monolit/internal/models"
	"encoding/json"
	"errors"
	"net/http"
)

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, accessToken, refreshToken, err := h.service.Refresh(r.Context(), model.RefreshTokenInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		if errors.Is(err, model.ErrInvalidRefreshToken) {
			http.Error(w, "invalid refresh token", http.StatusUnauthorized)
			return
		}

		http.Error(w, "failed to refresh token", http.StatusInternalServerError)
		return
	}

	userResponse, err := converter.UserModelToAPI(user)
	if err != nil {
		http.Error(w, "failed to convert user", http.StatusInternalServerError)
		return
	}

	response := dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userResponse,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
