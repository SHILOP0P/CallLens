package auth

import (
	"calllens/monolit/internal/auth/refresh"
	"calllens/monolit/internal/auth/token"
	model "calllens/monolit/internal/models"
	"context"
	"errors"
	"strings"
	"time"
)

func (s *Service) Refresh(ctx context.Context, input model.RefreshTokenInput) (model.User, string, string, error) {
	input.RefreshToken = strings.TrimSpace(input.RefreshToken)

	if input.RefreshToken == "" {
		return model.User{}, "", "", model.ErrInvalidRefreshToken
	}

	oldRefreshTokenHash, err := refresh.Hash(input.RefreshToken, s.refreshTokenSecret)
	if err != nil {
		return model.User{}, "", "", err
	}

	currentSession, err := s.refreshSessionRepository.GetRefreshSessionByHash(ctx, oldRefreshTokenHash)
	if err != nil {
		if errors.Is(err, model.ErrRefreshSessionNotFound) {
			return model.User{}, "", "", model.ErrInvalidRefreshToken
		}

		return model.User{}, "", "", err
	}

	now := time.Now().UTC()
	if currentSession.RevokedAt != nil || !currentSession.ExpiresAt.After(now) {
		return model.User{}, "", "", model.ErrInvalidRefreshToken
	}

	newRefreshToken, err := refresh.Generate()
	if err != nil {
		return model.User{}, "", "", err
	}

	newRefreshTokenHash, err := refresh.Hash(newRefreshToken, s.refreshTokenSecret)
	if err != nil {
		return model.User{}, "", "", err
	}

	rotatedSession, err := s.refreshSessionRepository.RotateRefreshSession(
		ctx,
		oldRefreshTokenHash,
		newRefreshTokenHash,
		now.Add(s.refreshTokenTTL),
	)
	if err != nil {
		if errors.Is(err, model.ErrRefreshSessionNotFound) {
			return model.User{}, "", "", model.ErrInvalidRefreshToken
		}

		return model.User{}, "", "", err
	}

	user, err := s.userRepository.GetUserByUUID(ctx, rotatedSession.UserID)
	if err != nil {
		return model.User{}, "", "", err
	}

	accessToken, err := token.GenerateAccessTokenWithSession(
		user.ID,
		rotatedSession.ID,
		string(user.Role),
		s.jwtSecret,
		s.accessTokenTTL,
	)
	if err != nil {
		return model.User{}, "", "", err
	}

	return user, accessToken, newRefreshToken, nil
}
