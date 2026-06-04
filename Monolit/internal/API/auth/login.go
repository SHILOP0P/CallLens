package auth

import (
	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/converter"
	model "calllens/monolit/internal/models"
	"encoding/json"
	"errors"
	"net/http"
)

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, accessToken, err := h.service.Login(r.Context(), model.LoginInput{Email: req.Email, Password: req.Password})
	if err != nil {
		if errors.Is(err, model.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusBadRequest)
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
		AccessToken: accessToken,
		User:        userResponse,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
