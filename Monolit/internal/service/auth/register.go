package auth

import (
	"calllens/monolit/internal/auth/password"
	model "calllens/monolit/internal/models"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const defaultUserRole = model.UserRoleUser

func (s *Service) Register(ctx context.Context, input model.CreateUserInput) (model.User, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.FullName = strings.TrimSpace(input.FullName)
	input.FullSurname = strings.TrimSpace(input.FullSurname)
	input.NickName = strings.TrimSpace(input.NickName)
	input.Post = normalizeOptionalString(input.Post)

	if input.Email == "" ||
		input.Password == "" ||
		input.FullName == "" ||
		input.FullSurname == "" ||
		input.NickName == "" {
		return model.User{}, model.ErrInvalidUserInput
	}

	if len(input.Password) < 8 {
		return model.User{}, model.ErrInvalidUserInput
	}

	_, err := s.userRepository.GetUserByEmail(ctx, input.Email)
	if err == nil {
		return model.User{}, model.ErrUserAlreadyExists
	}
	if !errors.Is(err, model.ErrUserNotFound) {
		return model.User{}, err
	}

	passwordHash, err := password.Hash(input.Password, s.passwordPepper)
	if err != nil {
		return model.User{}, err
	}

	userID, err := uuid.NewV7()
	if err != nil {
		return model.User{}, err
	}

	user := model.User{
		ID:           userID,
		Email:        input.Email,
		PasswordHash: passwordHash,
		FullName:     input.FullName,
		FullSurname:  input.FullSurname,
		NickName:     input.NickName,
		Role:         defaultUserRole,
		Post:         input.Post,
		CreatedAt:    time.Now().UTC(),
	}

	createUser, err := s.userRepository.CreateUser(ctx, user)
	if err != nil {
		return model.User{}, err
	}

	if s.billingRepository != nil {
		_, err = s.billingRepository.UpsertSubscription(ctx, model.UpsertSubscriptionInput{
			PlanCode: model.PlanCodePersonalStart,
			UserUUID: uuid.NullUUID{
				UUID:  createUser.ID,
				Valid: true,
			},
			Status:   model.SubscriptionStatusActive,
			StartsAt: time.Now().UTC(),
		})
		if err != nil {
			return model.User{}, err
		}
	}

	return createUser, nil
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
