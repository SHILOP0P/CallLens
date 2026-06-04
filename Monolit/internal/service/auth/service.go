package auth

import (
	repo "calllens/monolit/internal/repository"
	"time"
)

type Service struct {
	userRepository repo.UserRepository
	passwordPepper string
	jwtSecret      string
	accessTokenTTL time.Duration
}

func NewService(userRepository repo.UserRepository, passwordPepper string, jwtSecret string, accessTokenTTL time.Duration) *Service {
	return &Service{
		userRepository: userRepository,
		passwordPepper: passwordPepper,
		jwtSecret:      jwtSecret,
		accessTokenTTL: accessTokenTTL,
	}
}
