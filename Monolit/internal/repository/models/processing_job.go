package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type ProcessingJob struct {
	ID          uuid.UUID
	Type        string
	EntityUUID  uuid.UUID
	Status      string
	Attempts    int
	MaxAttempts int
	AvailableAt time.Time
	LockedAt    sql.NullTime
	LockedBy    sql.NullString
	LastError   sql.NullString
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
