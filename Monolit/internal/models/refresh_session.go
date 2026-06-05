package models

import (
	"time"

	"github.com/google/uuid"
)

type RefreshSession struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	RefreshTokenHash string
	UserAgent        *string
	IPAddress        *string
	CreatedAt        time.Time
	LastUsedAt       *time.Time
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	RevokedReason    *string
}
