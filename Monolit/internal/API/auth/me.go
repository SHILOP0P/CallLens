package auth

import (
	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/converter"
	"calllens/monolit/internal/httpserver/middleware"
	model "calllens/monolit/internal/models"
	"encoding/json"

	"errors"
	"net/http"
)

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.service.Me(r.Context(), userID)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to get user", http.StatusInternalServerError)
		return
	}

	userResponse, err := converter.UserModelToAPI(user)
	if err != nil {
		http.Error(w, "failed to convert user", http.StatusInternalServerError)
		return
	}

	response := dto.UserResponse{
		ID:          userResponse.ID,
		Email:       userResponse.Email,
		FullName:    userResponse.FullName,
		FullSurname: userResponse.FullSurname,
		NickName:    userResponse.NickName,
		Role:        userResponse.Role,
		Post:        userResponse.Post,
		CreatedAt:   userResponse.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
