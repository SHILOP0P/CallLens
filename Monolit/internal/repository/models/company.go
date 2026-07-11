package models

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID              uuid.UUID
	Name            string
	Tag             string
	ManagerUserUUID uuid.UUID
	MemberLimit     int
	CreatedAt       time.Time
	DeletedAt       *time.Time
}

type CompanyMember struct {
	CompanyUUID uuid.UUID
	UserUUID    uuid.UUID
	Role        string
	Status      string
	CreatedAt   time.Time
}
