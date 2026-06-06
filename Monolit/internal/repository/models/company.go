package models

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID              uuid.UUID
	Name            string
	ManagerUserUUID uuid.UUID
	MemberLimit     int
	CreatedAt       time.Time
}

type CompanyMember struct {
	CompanyUUID uuid.UUID
	UserUUID    uuid.UUID
	Role        string
	Status      string
	CreatedAt   time.Time
}
