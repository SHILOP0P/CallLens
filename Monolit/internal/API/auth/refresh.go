package auth

import (
	"bytes"
	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/converter"
	model "calllens/monolit/internal/models"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := refreshTokenFromRequest(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, response.CodeInvalidRequestBody, "invalid request body")
		return
	}

	user, accessToken, newRefreshToken, err := h.service.Refresh(r.Context(), model.RefreshTokenInput{
		RefreshToken: refreshToken,
	})
	if err != nil {
		if errors.Is(err, model.ErrInvalidRefreshToken) {
			h.clearAuthCookies(w, r)
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

	h.setAuthCookies(w, r, accessToken, newRefreshToken)

	resp := dto.AuthResponse{
		User: userResponse,
	}

	if err := response.WriteJSON(w, http.StatusOK, resp); err != nil {
		return
	}
}

func refreshTokenFromRequest(r *http.Request) (string, error) {
	if cookie, err := r.Cookie(refreshTokenCookieName); err == nil {
		return strings.TrimSpace(cookie.Value), nil
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return "", err
	}

	if len(bytes.TrimSpace(body)) == 0 {
		return "", nil
	}

	var req dto.RefreshRequest
	if err = json.Unmarshal(body, &req); err != nil {
		return "", err
	}

	return strings.TrimSpace(req.RefreshToken), nil
}
