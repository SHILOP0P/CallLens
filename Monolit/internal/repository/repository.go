package repository

import (
	"calllens/monolit/internal/models"
	"context"
	"time"

	"github.com/google/uuid"
)

type CallRepository interface {
	//POST
	CreateCall(ctx context.Context, call models.Call) (models.Call, error)
	//GET
	List(ctx context.Context, userID uuid.UUID) ([]models.Call, error)
	GetByUUID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (models.Call, error)
	//UPDATE
	UpdateCallTitle(ctx context.Context, id uuid.UUID, userID uuid.UUID, title string) (models.Call, error)
	//DELETE
	DeleteCall(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type UserRepository interface {
	//GET
	GetUserByUUID(ctx context.Context, id uuid.UUID) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	//POST
	CreateUser(ctx context.Context, user models.User) (models.User, error)
}

type RefreshSessionRepository interface {
	CreateRefreshSession(ctx context.Context, session models.RefreshSession) (models.RefreshSession, error)
	GetRefreshSessionByHash(ctx context.Context, refreshTokenHash string) (models.RefreshSession, error)
	GetRefreshSessionByUUID(ctx context.Context, sessionID uuid.UUID) (models.RefreshSession, error)
	RotateRefreshSession(ctx context.Context, oldRefreshTokenHash string, newRefreshTokenHash string, expiresAt time.Time) (models.RefreshSession, error)
	RevokeRefreshSession(ctx context.Context, sessionID uuid.UUID, reason string) error
	RevokeAllUserRefreshSessions(ctx context.Context, userID uuid.UUID, reason string) error
}
