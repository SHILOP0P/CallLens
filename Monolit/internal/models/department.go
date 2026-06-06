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
	Role           DepartmentMemberRole
	Status         MembershipStatus
	CreatedAt      time.Time
}

type DepartmentMemberRole string

const (
	DepartmentMemberRoleLeader   DepartmentMemberRole = "department_leader"
	DepartmentMemberRoleEmployee DepartmentMemberRole = "employee"
)

type CreateDepartmentInput struct {
	CompanyUUID uuid.UUID
	UserID      uuid.UUID
	Name        string
}

type AddDepartmentMemberInput struct {
	CompanyUUID    uuid.UUID
	DepartmentUUID uuid.UUID
	RequestUser    uuid.UUID
	UserUUID       uuid.UUID
	Role           DepartmentMemberRole
}
