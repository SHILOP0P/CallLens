package converter

import (
	"time"

	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/models"
)

func UserModelToAPI(user models.User) (dto.UserResponse, error) {
	return dto.UserResponse{
		ID:          user.ID.String(),
		Email:       user.Email,
		FullName:    user.FullName,
		FullSurname: user.FullSurname,
		Username:    user.Username,
		Role:        string(user.Role),
		Post:        user.Post,
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
	}, nil
}
