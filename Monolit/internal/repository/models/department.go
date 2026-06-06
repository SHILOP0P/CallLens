package models

import (
	"time"

	"github.com/google/uuid"
)

type Department struct {
	ID          uuid.UUID
	CompanyUUID uuid.UUID
	Name        string
	CreatedAt   time.Time
}

type DepartmentMember struct {
	DepartmentUUID uuid.UUID
	UserUUID       uuid.UUID
	Role           string
	Status         string
	CreatedAt      time.Time
}
