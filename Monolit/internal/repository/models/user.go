package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	FullName     string
	FullSurname  string
	NickName     string
	Role         string
	Post         sql.NullString
	CreatedAt    time.Time
}
