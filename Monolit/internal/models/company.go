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
	Username    string
	FullName    string
	FullSurname string
	Role        CompanyMemberRole
	Status      MembershipStatus
	CreatedAt   time.Time
}

type CompanyMemberRole string

const (
	CompanyMemberRoleManager  CompanyMemberRole = "company_manager"
	CompanyMemberRoleEmployee CompanyMemberRole = "employee"
)

type MembershipStatus string

const (
	MembershipStatusActive    MembershipStatus = "active"
	MembershipStatusSuspended MembershipStatus = "suspended"
	MembershipStatusLeft      MembershipStatus = "left"
)

type CreateCompanyInput struct {
	Name          string
	ManagerUserID uuid.UUID
}

type AddCompanyMemberInput struct {
	CompanyUUID uuid.UUID
	RequestUser uuid.UUID
	UserUUID    uuid.UUID
	Role        CompanyMemberRole
}

type UpdateCompanyMemberRoleInput struct {
	CompanyUUID uuid.UUID
	RequestUser uuid.UUID
	UserUUID    uuid.UUID
	Role        CompanyMemberRole
}

type UpdateCompanyMemberStatusInput struct {
	CompanyUUID uuid.UUID
	RequestUser uuid.UUID
	UserUUID    uuid.UUID
	Status      MembershipStatus
}

type CompanyMembersOverview struct {
	CompanyUUID      uuid.UUID
	Manager          *CompanyMember
	CompanyEmployees []CompanyMember
	Departments      []DepartmentMembersOverview
}
