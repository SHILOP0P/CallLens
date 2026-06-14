package auth

import (
	"calllens/monolit/internal/logger"
	"calllens/monolit/internal/models"
	repo "calllens/monolit/internal/repository"
	"context"
	"time"
)

type BillingRepository interface {
	UpsertSubscription(ctx context.Context, input models.UpsertSubscriptionInput) (models.Subscription, error)
}

type Service struct {
	userRepository           repo.UserRepository
	refreshSessionRepository repo.RefreshSessionRepository
	billingRepository        BillingRepository

	passwordPepper     string
	jwtSecret          string
	accessTokenTTL     time.Duration
	refreshTokenSecret string
	refreshTokenTTL    time.Duration
	log                logger.Logger
}

func (s *Service) SetBillingRepository(repository BillingRepository) {
	s.billingRepository = repository
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
