package auth

import (
	"context"
	"strings"
	"time"

	"calllens/monolit/internal/models"

	"github.com/google/uuid"
)

const avatarURL = "/api/v1/auth/me/avatar"

func (s *Service) UpdateProfile(ctx context.Context, input models.UpdateUserProfileInput) (models.User, error) {
	if input.UserUUID == uuid.Nil {
		return models.User{}, models.ErrInvalidUserInput
	}

	input.FullName = normalizeRequiredPatchString(input.FullName)
	input.FullSurname = normalizeRequiredPatchString(input.FullSurname)
	input.Post = normalizeOptionalString(input.Post)
	input.Phone = normalizeOptionalString(input.Phone)
	input.Timezone = normalizeOptionalString(input.Timezone)

	if input.FullName != nil && *input.FullName == "" {
		return models.User{}, models.ErrInvalidUserInput
	}
	if input.FullSurname != nil && *input.FullSurname == "" {
		return models.User{}, models.ErrInvalidUserInput
	}
	if input.Timezone != nil {
		if _, err := time.LoadLocation(*input.Timezone); err != nil {
			return models.User{}, models.ErrInvalidUserInput
		}
	}

	return s.userRepository.UpdateProfile(ctx, input)
}

func (s *Service) UploadAvatar(ctx context.Context, input models.SaveUserAvatarInput) (models.UserAvatarResponse, error) {
	if s.avatarStorage == nil {
		return models.UserAvatarResponse{}, models.ErrInvalidUserInput
	}

	saved, err := s.avatarStorage.Save(ctx, input)
	if err != nil {
		return models.UserAvatarResponse{}, err
	}

	now := time.Now().UTC()
	_, err = s.userRepository.UpdateAvatar(ctx, models.UserAvatarUpdate{
		UserUUID:  input.UserUUID,
		Path:      &saved.Path,
		MimeType:  &saved.MimeType,
		SizeBytes: &saved.SizeBytes,
		UpdatedAt: &now,
	})
	if err != nil {
		_ = s.avatarStorage.Delete(context.Background(), saved.Path)
		return models.UserAvatarResponse{}, err
	}

	return models.UserAvatarResponse{AvatarURL: avatarURL, UpdatedAt: now}, nil
}

func (s *Service) DeleteAvatar(ctx context.Context, userID uuid.UUID) (models.UserAvatarResponse, error) {
	user, err := s.userRepository.GetUserByUUID(ctx, userID)
	if err != nil {
		return models.UserAvatarResponse{}, err
	}

	if user.AvatarPath != nil && s.avatarStorage != nil {
		_ = s.avatarStorage.Delete(ctx, *user.AvatarPath)
	}

	_, err = s.userRepository.DeleteAvatar(ctx, userID)
	if err != nil {
		return models.UserAvatarResponse{}, err
	}

	return models.UserAvatarResponse{AvatarURL: avatarURL, UpdatedAt: time.Now().UTC()}, nil
}

func normalizeRequiredPatchString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}
