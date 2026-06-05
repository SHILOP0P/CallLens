package auth

import (
	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/converter"
	model "calllens/monolit/internal/models"
	"encoding/json"
	"errors"
	"net/http"
)

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, response.CodeInvalidRequestBody, "invalid request body")
		return
	}

	user, accessToken, refreshToken, err := h.service.Refresh(r.Context(), model.RefreshTokenInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		if errors.Is(err, model.ErrInvalidRefreshToken) {
			response.WriteError(w, http.StatusUnauthorized, response.CodeInvalidRefreshToken, "invalid refresh token")
			return
		}

		response.WriteError(w, http.StatusInternalServerError, response.CodeFailedToRefreshToken, "failed to refresh token")
		return
	}

	userResponse, err := converter.UserModelToAPI(user)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, response.CodeFailedToConvertUser, "failed to convert user")
		return
	}

	resp := dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userResponse,
	}

	if err := response.WriteJSON(w, http.StatusOK, resp); err != nil {
		return
	}
}
