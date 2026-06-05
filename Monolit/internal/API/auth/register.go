package auth

import (
	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/converter"
	"calllens/monolit/internal/models"
	"encoding/json"
	"errors"
	"net/http"
)

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, response.CodeInvalidRequestBody, "invalid request body")
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
			response.WriteError(w, http.StatusBadRequest, response.CodeInvalidUserInput, "invalid user input")
			return
		}
		if errors.Is(err, models.ErrUserAlreadyExists) {
			response.WriteError(w, http.StatusConflict, response.CodeUserAlreadyExists, "user already exists")
			return
		}

		response.WriteError(w, http.StatusInternalServerError, response.CodeFailedToRegisterUser, "failed to register user")
		return
	}

	userResponse, err := converter.UserModelToAPI(user)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, response.CodeFailedToConvertUser, "failed to convert user")
		return
	}

	resp := dto.RegisterResponse{
		User: userResponse,
	}

	if err := response.WriteJSON(w, http.StatusCreated, resp); err != nil {
		return
	}
}
