package auth

import (
	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/converter"
	"calllens/monolit/internal/models"
	"encoding/json"
	"errors"
	"net/http"
)

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.service.Register(r.Context(), models.CreateUserInput{
		Email:       req.Email,
		Password:    req.Password,
		FullName:    req.FullName,
		FullSurname: req.FullSurname,
		NickName:    req.NickName,
		Post:        req.Post,
	})
	if err != nil {
		if errors.Is(err, models.ErrInvalidUserInput) {
			http.Error(w, "invalid user input", http.StatusBadRequest)
			return
		}
		if errors.Is(err, models.ErrUserAlreadyExists) {
			http.Error(w, "user already exists", http.StatusConflict)
			return
		}

		http.Error(w, "failed to register user", http.StatusInternalServerError)
		return
	}

	userResponse, err := converter.UserModelToAPI(user)
	if err != nil {
		http.Error(w, "failed to convert user", http.StatusInternalServerError)
		return
	}

	response := dto.RegisterResponse{
		User: userResponse,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
