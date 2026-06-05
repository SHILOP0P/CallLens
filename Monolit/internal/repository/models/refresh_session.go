package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type RefreshSession struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	RefreshTokenHash string
	UserAgent        sql.NullString
	IPAddress        sql.NullString
	CreatedAt        time.Time
	LastUsedAt       sql.NullTime
	ExpiresAt        time.Time
	RevokedAt        sql.NullTime
	RevokedReason    sql.NullString
}
