package service

import (
	"calllens/monolit/internal/models"
	"context"

	"github.com/google/uuid"
)

type CallService interface {
	//POST
	CreateCall(ctx context.Context, input models.CreateCallInput) (models.Call, error)

	//GET
	List(ctx context.Context, userID uuid.UUID) ([]models.Call, error)
	GetByUUID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (models.Call, error)
	GetAudioByUUID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (models.File, error)
	GetTranscriptionByCallUUID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (models.Transcription, error)

	//UPDATE
	UpdateCallTitle(ctx context.Context, id uuid.UUID, userID uuid.UUID, title string) (models.Call, error)
	UpdateCallStatus(ctx context.Context, input models.UpdateCallStatusInput) (models.Call, error)
	//DELETE
	DeleteCall(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type AuthService interface {
	Register(ctx context.Context, input models.CreateUserInput) (models.User, error)
	Login(ctx context.Context, input models.LoginInput) (models.User, string, string, error)
	Refresh(ctx context.Context, input models.RefreshTokenInput) (models.User, string, string, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
	LogoutAll(ctx context.Context, userID uuid.UUID) error
	Me(ctx context.Context, userID uuid.UUID) (models.User, error)
}

type CompanyService interface {
	CreateCompany(ctx context.Context, input models.CreateCompanyInput) (models.Company, error)
	AddCompanyMember(ctx context.Context, input models.AddCompanyMemberInput) (models.CompanyMember, error)
	UpdateCompanyMemberRole(ctx context.Context, input models.UpdateCompanyMemberRoleInput) (models.CompanyMember, error)
	UpdateCompanyMemberStatus(ctx context.Context, input models.UpdateCompanyMemberStatusInput) (models.CompanyMember, error)
	ListUserCompanies(ctx context.Context, userID uuid.UUID) ([]models.Company, error)
	GetCompanyByUUID(ctx context.Context, companyID uuid.UUID, userID uuid.UUID) (models.Company, error)
	GetCompanyMembersOverview(ctx context.Context, companyID uuid.UUID, userID uuid.UUID) (models.CompanyMembersOverview, error)
}

type DepartmentService interface {
	CreateDepartment(ctx context.Context, input models.CreateDepartmentInput) (models.Department, error)
	AddDepartmentMember(ctx context.Context, input models.AddDepartmentMemberInput) (models.DepartmentMember, error)
	ListDepartmentMembers(ctx context.Context, companyID uuid.UUID, departmentID uuid.UUID, userID uuid.UUID) ([]models.DepartmentMember, error)
	UpdateDepartmentMemberRole(ctx context.Context, input models.UpdateDepartmentMemberRoleInput) (models.DepartmentMember, error)
	UpdateDepartmentMemberStatus(ctx context.Context, input models.UpdateDepartmentMemberStatusInput) (models.DepartmentMember, error)
	ListCompanyDepartments(ctx context.Context, companyID uuid.UUID, userID uuid.UUID) ([]models.Department, error)
}

type AnalysisInstructionService interface {
	Create(ctx context.Context, input models.CreateAnalysisInstructionInput) (models.AnalysisInstruction, error)
	List(ctx context.Context, input models.ListAnalysisInstructionsInput) ([]models.AnalysisInstruction, error)
	GetFile(ctx context.Context, id uuid.UUID, userID uuid.UUID) (models.File, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type AnalysisService interface {
	AnalyzeCall(ctx context.Context, input models.AnalyzeCallInput) (models.CallAnalysis, error)
	GetByCallUUID(ctx context.Context, callUUID uuid.UUID, userID uuid.UUID) (models.CallAnalysis, error)
}

type BillingService interface {
	ListPlans(ctx context.Context) ([]models.Plan, error)
	GetPersonalSubscription(ctx context.Context, userID uuid.UUID) (models.Subscription, error)
	GetCompanySubscription(ctx context.Context, input models.GetCompanySubscriptionInput) (models.Subscription, error)
	ActivatePersonalSubscription(ctx context.Context, input models.ActivatePersonalSubscriptionInput) (models.Subscription, error)
	ActivateCompanySubscription(ctx context.Context, input models.ActivateCompanySubscriptionInput) (models.Subscription, error)
	CancelCompanySubscription(ctx context.Context, input models.CancelCompanySubscriptionInput) (models.Subscription, error)
}

type InvitationService interface {
	CreateCompanyInvitation(ctx context.Context, input models.CreateCompanyInvitationInput) (models.MembershipInvitation, error)
	CreateDepartmentInvitation(ctx context.Context, input models.CreateDepartmentInvitationInput) (models.MembershipInvitation, error)
	ListUserInvitations(ctx context.Context, input models.ListUserInvitationsInput) ([]models.MembershipInvitation, error)
	AcceptInvitation(ctx context.Context, input models.AcceptInvitationInput) (models.MembershipInvitation, error)
	DeclineInvitation(ctx context.Context, input models.DeclineInvitationInput) (models.MembershipInvitation, error)
	CancelInvitation(ctx context.Context, input models.CancelInvitationInput) (models.MembershipInvitation, error)
}
