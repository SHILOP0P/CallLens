package models

import (
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
	Role         UserRole
	Post         *string
	CreatedAt    time.Time
}

type UserRole string

const (
	UserRoleUser       UserRole = "user"
	UserRoleHelper     UserRole = "helper"
	UserRoleAdmin      UserRole = "admin"
	UserRoleSuperAdmin UserRole = "superadmin"
)

type CreateUserInput struct {
	Email       string
	Password    string
	FullName    string
	FullSurname string
	NickName    string
	Post        *string
}

type LoginInput struct {
	Email     string
	Password  string
	UserAgent *string
	IPAddress *string
}

type RefreshTokenInput struct {
	RefreshToken string
}
