package auth

import (
	repo "calllens/monolit/internal/repository"
	"time"
)

type Service struct {
	userRepository           repo.UserRepository
	refreshSessionRepository repo.RefreshSessionRepository

	passwordPepper     string
	jwtSecret          string
	accessTokenTTL     time.Duration
	refreshTokenSecret string
	refreshTokenTTL    time.Duration
}

func NewService(
	userRepository repo.UserRepository,
	refreshSessionRepository repo.RefreshSessionRepository,
	passwordPepper string,
	jwtSecret string,
	accessTokenTTL time.Duration,
	refreshTokenSecret string,
	refreshTokenTTL time.Duration,
) *Service {
	return &Service{
		userRepository:           userRepository,
		refreshSessionRepository: refreshSessionRepository,
		passwordPepper:           passwordPepper,
		jwtSecret:                jwtSecret,
		accessTokenTTL:           accessTokenTTL,
		refreshTokenSecret:       refreshTokenSecret,
		refreshTokenTTL:          refreshTokenTTL,
	}
}
