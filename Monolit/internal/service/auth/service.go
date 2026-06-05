package auth

import (
	"calllens/monolit/internal/logger"
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
	log                logger.Logger
}

func NewService(
	userRepository repo.UserRepository,
	refreshSessionRepository repo.RefreshSessionRepository,
	passwordPepper string,
	jwtSecret string,
	accessTokenTTL time.Duration,
	refreshTokenSecret string,
	refreshTokenTTL time.Duration,
	log logger.Logger,
) *Service {
	if log == nil {
		log = logger.NewNop()
	}

	return &Service{
		userRepository:           userRepository,
		refreshSessionRepository: refreshSessionRepository,
		passwordPepper:           passwordPepper,
		jwtSecret:                jwtSecret,
		accessTokenTTL:           accessTokenTTL,
		refreshTokenSecret:       refreshTokenSecret,
		refreshTokenTTL:          refreshTokenTTL,
		log:                      log,
	}
}
