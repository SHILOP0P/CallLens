package auth

import (
	"calllens/monolit/internal/auth/password"
	"calllens/monolit/internal/auth/refresh"
	"calllens/monolit/internal/auth/token"
	model "calllens/monolit/internal/models"
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Service) Login(ctx context.Context, input model.LoginInput) (model.User, string, string, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))

	if email == "" || input.Password == "" {
		s.log.Warn(ctx, "login failed", zap.String("reason", "empty_credentials"))
		return model.User{}, "", "", model.ErrInvalidCredentials
	}

	user, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		s.log.Warn(ctx, "login failed", zap.String("reason", "invalid_credentials"), zap.Error(err))
		return model.User{}, "", "", model.ErrInvalidCredentials
	}

	if err := password.Compare(input.Password, user.PasswordHash, s.passwordPepper); err != nil {
		s.log.Warn(ctx, "login failed", zap.String("reason", "invalid_credentials"), zap.String("user_id", user.ID.String()))
		return model.User{}, "", "", model.ErrInvalidCredentials
	}

	refreshToken, err := refresh.Generate()
	if err != nil {
		s.log.Error(ctx, "failed to generate refresh token", zap.String("user_id", user.ID.String()), zap.Error(err))
		return model.User{}, "", "", err
	}

	refreshTokenHash, err := refresh.Hash(refreshToken, s.refreshTokenSecret)
	if err != nil {
		s.log.Error(ctx, "failed to hash refresh token", zap.String("user_id", user.ID.String()), zap.Error(err))
		return model.User{}, "", "", err
	}

	sessionID, err := uuid.NewV7()
	if err != nil {
		s.log.Error(ctx, "failed to generate refresh session uuid", zap.String("user_id", user.ID.String()), zap.Error(err))
		return model.User{}, "", "", err
	}

	now := time.Now().UTC()
	session := model.RefreshSession{
		ID:               sessionID,
		UserID:           user.ID,
		RefreshTokenHash: refreshTokenHash,
		UserAgent:        input.UserAgent,
		IPAddress:        input.IPAddress,
		CreatedAt:        now,
		ExpiresAt:        now.Add(s.refreshTokenTTL),
	}

	createdSession, err := s.refreshSessionRepository.CreateRefreshSession(ctx, session)
	if err != nil {
		s.log.Error(ctx, "failed to create refresh session", zap.String("user_id", user.ID.String()), zap.String("session_id", sessionID.String()), zap.Error(err))
		return model.User{}, "", "", err
	}

	accessToken, err := token.GenerateAccessTokenWithSession(
		user.ID,
		createdSession.ID,
		string(user.Role),
		s.jwtSecret,
		s.accessTokenTTL,
	)
	if err != nil {
		s.log.Error(ctx, "failed to generate access token", zap.String("user_id", user.ID.String()), zap.String("session_id", createdSession.ID.String()), zap.Error(err))
		return model.User{}, "", "", err
	}

	s.log.Info(ctx, "user logged in", zap.String("user_id", user.ID.String()), zap.String("session_id", createdSession.ID.String()))

	return user, accessToken, refreshToken, nil
}
