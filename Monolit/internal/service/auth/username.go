package auth

import (
	"context"
	"errors"

	"calllens/monolit/internal/models"
	"calllens/monolit/internal/username"

	"github.com/google/uuid"
)

func (s *Service) UpdateUsername(ctx context.Context, input models.UpdateUsernameInput) (models.User, error) {
	if input.UserUUID == uuid.Nil {
		return models.User{}, models.ErrInvalidUserInput
	}

	normalized, ok := username.Normalize(input.Username)
	if !ok {
		return models.User{}, models.ErrInvalidUserInput
	}

	existing, err := s.userRepository.GetUserByUsername(ctx, normalized)
	if err == nil && existing.ID != input.UserUUID {
		return models.User{}, models.ErrUserAlreadyExists
	}
	if err != nil && !errors.Is(err, models.ErrUserNotFound) {
		return models.User{}, err
	}

	return s.userRepository.UpdateUsername(ctx, models.UpdateUsernameInput{
		UserUUID: input.UserUUID,
		Username: normalized,
	})
}

func (s *Service) GetUserByUsername(ctx context.Context, value string) (models.User, error) {
	normalized, ok := username.Normalize(value)
	if !ok {
		return models.User{}, models.ErrInvalidUserInput
	}

	return s.userRepository.GetUserByUsername(ctx, normalized)
}
