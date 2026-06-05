package service

import (
	"calllens/monolit/internal/models"
	"context"

	"github.com/google/uuid"
)

type CallService interface {
	//POST
	CreateCall(ctx context.Context, input models.CreateCallInput) (models.Call, error)

	//GET
	List(ctx context.Context, userID uuid.UUID) ([]models.Call, error)
	GetByUUID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (models.Call, error)
	GetAudioByUUID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (models.File, error)

	//UPDATE
	UpdateCallTitle(ctx context.Context, id uuid.UUID, userID uuid.UUID, title string) (models.Call, error)
	//DELETE
	DeleteCall(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type AuthService interface {
	Register(ctx context.Context, input models.CreateUserInput) (models.User, error)
	Login(ctx context.Context, input models.LoginInput) (models.User, string, string, error)
	Refresh(ctx context.Context, input models.RefreshTokenInput) (models.User, string, string, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
	LogoutAll(ctx context.Context, userID uuid.UUID) error
	Me(ctx context.Context, userID uuid.UUID) (models.User, error)
}
