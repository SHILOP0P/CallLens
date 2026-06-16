package models

import (
	"time"

	"github.com/google/uuid"
)

type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusDeclined InvitationStatus = "declined"
	InvitationStatusCanceled InvitationStatus = "canceled"
	InvitationStatusExpired  InvitationStatus = "expired"
)

type MembershipInvitation struct {
	ID                uuid.UUID
	CompanyUUID       uuid.UUID
	DepartmentUUID    uuid.NullUUID
	InvitedUserUUID   uuid.UUID
	InvitedByUserUUID uuid.UUID
	CompanyRole       CompanyMemberRole
	DepartmentRole    *DepartmentMemberRole
	Status            InvitationStatus
	ExpiresAt         time.Time
	RespondedAt       *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type CreateCompanyInvitationInput struct {
	CompanyUUID uuid.UUID
	RequestUser uuid.UUID
	UserUUID    uuid.UUID
	Username    string
	Role        CompanyMemberRole
}

type CreateDepartmentInvitationInput struct {
	CompanyUUID    uuid.UUID
	DepartmentUUID uuid.UUID
	RequestUser    uuid.UUID
	UserUUID       uuid.UUID
	Username       string
	Role           DepartmentMemberRole
}

type ListUserInvitationsInput struct {
	UserUUID uuid.UUID
	Status   InvitationStatus
}

type AcceptInvitationInput struct {
	InvitationUUID uuid.UUID
	RequestUser    uuid.UUID
}

type DeclineInvitationInput struct {
	InvitationUUID uuid.UUID
	RequestUser    uuid.UUID
}

type CancelInvitationInput struct {
	CompanyUUID    uuid.UUID
	DepartmentUUID uuid.NullUUID
	InvitationUUID uuid.UUID
	RequestUser    uuid.UUID
}
