package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID
	Email           string
	PasswordHash    string
	FullName        string
	FullSurname     string
	Username        string
	Role            string
	Post            sql.NullString
	Phone           sql.NullString
	Timezone        sql.NullString
	AvatarPath      sql.NullString
	AvatarMime      sql.NullString
	AvatarSize      sql.NullInt64
	AvatarUpdatedAt sql.NullTime
	CreatedAt       time.Time
}
