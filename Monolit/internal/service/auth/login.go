package auth

import (
	"calllens/monolit/internal/auth/password"
	"calllens/monolit/internal/auth/token"
	model "calllens/monolit/internal/models"
	"context"
	"strings"
)

func (s *Service) Login(ctx context.Context, input model.LoginInput) (model.User, string, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))

	if email == "" || input.Password == "" {
		return model.User{}, "", model.ErrInvalidCredentials
	}

	user, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return model.User{}, "", err
	}

	if err := password.Compare(input.Password, user.PasswordHash, s.passwordPepper); err != nil {
		return model.User{}, "", model.ErrInvalidCredentials
	}

	accessToken, err := token.GenerateAccessToken(user.ID, string(user.Role), s.jwtSecret, s.accessTokenTTL)
	if err != nil {
		return model.User{}, "", err
	}

	return user, accessToken, nil
}
