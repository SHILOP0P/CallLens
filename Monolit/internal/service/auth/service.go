package auth

import (
	"context"
	"time"

	"calllens/monolit/internal/logger"
	"calllens/monolit/internal/models"
	repo "calllens/monolit/internal/repository"
	"calllens/monolit/internal/storage"
)

type BillingRepository interface {
	UpsertSubscription(ctx context.Context, input models.UpsertSubscriptionInput) (models.Subscription, error)
}

type Service struct {
	userRepository           repo.UserRepository
	refreshSessionRepository repo.RefreshSessionRepository
	billingRepository        BillingRepository
	companyRepository        repo.CompanyRepository
	preferencesRepository    repo.UserPreferencesRepository
	avatarStorage            storage.AvatarStorage

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

func (s *Service) SetCompanyRepository(repository repo.CompanyRepository) {
	s.companyRepository = repository
}

func (s *Service) SetPreferencesRepository(repository repo.UserPreferencesRepository) {
	s.preferencesRepository = repository
}

func (s *Service) SetAvatarStorage(avatarStorage storage.AvatarStorage) {
	s.avatarStorage = avatarStorage
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
