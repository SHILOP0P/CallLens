package repository

import (
	"calllens/monolit/internal/models"
	"context"
	"time"

	"github.com/google/uuid"
)

type CallRepository interface {
	//POST
	CreateCall(ctx context.Context, call models.Call) (models.Call, error)
	CreateCallWithProcessingJob(ctx context.Context, call models.Call, job models.ProcessingJob) (models.Call, error)
	//GET
	List(ctx context.Context, userID uuid.UUID) ([]models.Call, error)
	GetByUUID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (models.Call, error)
	GetByUUIDForProcessing(ctx context.Context, id uuid.UUID) (models.Call, error)
	//UPDATE
	UpdateCallTitle(ctx context.Context, id uuid.UUID, userID uuid.UUID, title string) (models.Call, error)
	UpdateCallStatus(ctx context.Context, id uuid.UUID, status models.CallStatus) (models.Call, error)
	//DELETE
	DeleteCall(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	//PROCESSING
	TakeNextForProcessing(ctx context.Context) (models.Call, error)
}

type UserRepository interface {
	//GET
	GetUserByUUID(ctx context.Context, id uuid.UUID) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	//POST
	CreateUser(ctx context.Context, user models.User) (models.User, error)
}

type CompanyRepository interface {
	CreateCompany(ctx context.Context, company models.Company, member models.CompanyMember) (models.Company, error)
	AddCompanyMember(ctx context.Context, member models.CompanyMember) (models.CompanyMember, error)
	UpdateCompanyMemberRole(ctx context.Context, companyID uuid.UUID, userID uuid.UUID, role models.CompanyMemberRole) (models.CompanyMember, error)
	UpdateCompanyMemberStatus(ctx context.Context, companyID uuid.UUID, userID uuid.UUID, status models.MembershipStatus) (models.CompanyMember, error)
	ListUserCompanies(ctx context.Context, userID uuid.UUID) ([]models.Company, error)
	GetCompanyByUUID(ctx context.Context, companyID uuid.UUID, userID uuid.UUID) (models.Company, error)
	GetManagedCompanyByUserUUID(ctx context.Context, userID uuid.UUID) (models.Company, error)
	GetCompanyMember(ctx context.Context, companyID uuid.UUID, userID uuid.UUID) (models.CompanyMember, error)
	GetCompanyMembersOverview(ctx context.Context, companyID uuid.UUID) (models.CompanyMembersOverview, error)
}

type DepartmentRepository interface {
	CreateDepartment(ctx context.Context, department models.Department) (models.Department, error)
	AddDepartmentMember(ctx context.Context, companyID uuid.UUID, member models.DepartmentMember) (models.DepartmentMember, error)
	ListDepartmentMembers(ctx context.Context, companyID uuid.UUID, departmentID uuid.UUID) ([]models.DepartmentMember, error)
	UpdateDepartmentMemberRole(ctx context.Context, companyID uuid.UUID, departmentID uuid.UUID, userID uuid.UUID, role models.DepartmentMemberRole) (models.DepartmentMember, error)
	UpdateDepartmentMemberStatus(ctx context.Context, companyID uuid.UUID, departmentID uuid.UUID, userID uuid.UUID, status models.MembershipStatus) (models.DepartmentMember, error)
	ListVisibleCompanyDepartments(ctx context.Context, companyID uuid.UUID, userID uuid.UUID) ([]models.Department, error)
	GetDepartmentMember(ctx context.Context, companyID uuid.UUID, departmentID uuid.UUID, userID uuid.UUID) (models.DepartmentMember, error)
}

type RefreshSessionRepository interface {
	CreateRefreshSession(ctx context.Context, session models.RefreshSession) (models.RefreshSession, error)
	GetRefreshSessionByHash(ctx context.Context, refreshTokenHash string) (models.RefreshSession, error)
	GetRefreshSessionByUUID(ctx context.Context, sessionID uuid.UUID) (models.RefreshSession, error)
	RotateRefreshSession(ctx context.Context, oldRefreshTokenHash string, newRefreshTokenHash string, expiresAt time.Time) (models.RefreshSession, error)
	RevokeRefreshSession(ctx context.Context, sessionID uuid.UUID, reason string) error
	RevokeAllUserRefreshSessions(ctx context.Context, userID uuid.UUID, reason string) error
}

type TranscriptionRepository interface {
	Create(ctx context.Context, transcription models.Transcription) (models.Transcription, error)
	GetByCallUUID(ctx context.Context, callID uuid.UUID) (models.Transcription, error)
	MarkTranscribed(ctx context.Context, id uuid.UUID, text string, language *string) (models.Transcription, error)
	MarkFailed(ctx context.Context, id uuid.UUID, errorMessage string) (models.Transcription, error)
}

type AnalysisRepository interface {
	Create(ctx context.Context, analysis models.CallAnalysis) (models.CallAnalysis, error)
	GetByCallUUID(ctx context.Context, callID uuid.UUID) (models.CallAnalysis, error)
	MarkProcessing(ctx context.Context, id uuid.UUID) (models.CallAnalysis, error)
	MarkDone(ctx context.Context, id uuid.UUID, result models.AnalysisResult) (models.CallAnalysis, error)
	MarkFailed(ctx context.Context, id uuid.UUID, errorMessage string) (models.CallAnalysis, error)
}

type ProcessingJobRepository interface {
	Create(ctx context.Context, job models.ProcessingJob) (models.ProcessingJob, error)
	TakeNext(ctx context.Context, workerID string, staleAfter time.Duration) (models.ProcessingJob, error)
	MarkDone(ctx context.Context, id uuid.UUID) (models.ProcessingJob, error)
	MarkRetry(ctx context.Context, id uuid.UUID, lastError string, delay time.Duration) (models.ProcessingJob, error)
	MarkFailed(ctx context.Context, id uuid.UUID, lastError string) (models.ProcessingJob, error)
}

type AnalysisInstructionRepository interface {
	Create(ctx context.Context, instruction models.AnalysisInstruction) (models.AnalysisInstruction, error)
	GetByUUID(ctx context.Context, id uuid.UUID) (models.AnalysisInstruction, error)
	List(ctx context.Context, input models.ListAnalysisInstructionsInput) ([]models.AnalysisInstruction, error)
	CountActive(ctx context.Context, input models.ListAnalysisInstructionsInput) (int, error)
	Deactivate(ctx context.Context, id uuid.UUID) error
}
