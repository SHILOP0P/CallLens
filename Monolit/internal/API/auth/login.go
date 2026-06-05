package auth

import (
	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/converter"
	model "calllens/monolit/internal/models"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
)

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, accessToken, refreshToken, err := h.service.Login(r.Context(), model.LoginInput{
		Email:     req.Email,
		Password:  req.Password,
		UserAgent: optionalString(r.UserAgent()),
		IPAddress: clientIPAddress(r),
	})
	if err != nil {
		if errors.Is(err, model.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "failed to login", http.StatusInternalServerError)
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

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return &value
}

func clientIPAddress(r *http.Request) *string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		if parsedIP := net.ParseIP(host); parsedIP != nil {
			return &host
		}
	}

	remoteAddr := strings.TrimSpace(r.RemoteAddr)
	if parsedIP := net.ParseIP(remoteAddr); parsedIP != nil {
		return &remoteAddr
	}

	return nil
}
