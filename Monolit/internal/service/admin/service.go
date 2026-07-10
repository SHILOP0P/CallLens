package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strings"
	"time"

	"calllens/monolit/internal/models"
	"calllens/monolit/internal/storage"

	"github.com/google/uuid"
)

var errAuditRepositoryNotConfigured = errors.New("admin audit repository is not configured")

type AuditRepository interface {
	CreateAdminAuditLog(ctx context.Context, audit models.AdminAuditLog) (models.AdminAuditLog, error)
	ListAdminUsers(ctx context.Context, input models.ListAdminUsersInput) (models.ListAdminUsersResult, error)
	GetAdminUserByUUID(ctx context.Context, userID uuid.UUID) (models.AdminUser, error)
	ChangeAdminUserRole(ctx context.Context, input models.ChangeAdminUserRoleInput) (models.AdminUser, error)
	ListAdminUserSessions(ctx context.Context, userID uuid.UUID) ([]models.AdminUserSession, error)
	RevokeAdminUserSession(ctx context.Context, input models.AdminSessionMutationInput) error
	RevokeAllAdminUserSessions(ctx context.Context, input models.AdminSessionMutationInput) error
	ListAdminCompanies(ctx context.Context, input models.ListAdminCompaniesInput) (models.ListAdminCompaniesResult, error)
	GetAdminCompanyByUUID(ctx context.Context, companyID uuid.UUID) (models.AdminCompany, error)
	GetAdminPersonalSubscription(ctx context.Context, userID uuid.UUID) (models.AdminSubscription, error)
	GetAdminCompanySubscription(ctx context.Context, companyID uuid.UUID) (models.AdminSubscription, error)
	GrantAdminSubscription(ctx context.Context, input models.GrantAdminSubscriptionInput) (models.AdminSubscription, error)
	CancelAdminSubscription(ctx context.Context, input models.CancelAdminSubscriptionInput) (models.AdminSubscription, error)
}

func (s *Service) ListCompanies(ctx context.Context, input models.ListAdminCompaniesInput) (models.ListAdminCompaniesResult, error) {
	if s.auditRepository == nil || input.Limit < 1 || input.Limit > 100 || input.Offset < 0 {
		return models.ListAdminCompaniesResult{}, models.ErrInvalidAdminInput
	}
	return s.auditRepository.ListAdminCompanies(ctx, input)
}
func (s *Service) GetCompany(ctx context.Context, id uuid.UUID) (models.AdminCompany, error) {
	if s.auditRepository == nil || id == uuid.Nil {
		return models.AdminCompany{}, models.ErrInvalidAdminInput
	}
	return s.auditRepository.GetAdminCompanyByUUID(ctx, id)
}
func (s *Service) GetPersonalSubscription(ctx context.Context, id uuid.UUID) (models.AdminSubscription, error) {
	if s.auditRepository == nil || id == uuid.Nil {
		return models.AdminSubscription{}, models.ErrInvalidAdminInput
	}
	return s.auditRepository.GetAdminPersonalSubscription(ctx, id)
}
func (s *Service) GetCompanySubscription(ctx context.Context, id uuid.UUID) (models.AdminSubscription, error) {
	if s.auditRepository == nil || id == uuid.Nil {
		return models.AdminSubscription{}, models.ErrInvalidAdminInput
	}
	return s.auditRepository.GetAdminCompanySubscription(ctx, id)
}
func (s *Service) GrantSubscription(ctx context.Context, in models.GrantAdminSubscriptionInput) (models.AdminSubscription, error) {
	if s.auditRepository == nil || in.ActorUserUUID == uuid.Nil || in.PlanCode == "" || in.EndsAt.IsZero() || in.StartsAt.IsZero() || !in.EndsAt.After(in.StartsAt) || strings.TrimSpace(in.Metadata.Reason) == "" || ((in.UserUUID == uuid.Nil) == (in.CompanyUUID == uuid.Nil)) || in.StartsAt.After(s.now().Add(time.Minute)) {
		return models.AdminSubscription{}, models.ErrInvalidAdminInput
	}
	return s.auditRepository.GrantAdminSubscription(ctx, in)
}
func (s *Service) CancelSubscription(ctx context.Context, in models.CancelAdminSubscriptionInput) (models.AdminSubscription, error) {
	if s.auditRepository == nil || in.ActorUserUUID == uuid.Nil || strings.TrimSpace(in.Metadata.Reason) == "" || ((in.UserUUID == uuid.Nil) == (in.CompanyUUID == uuid.Nil)) {
		return models.AdminSubscription{}, models.ErrInvalidAdminInput
	}
	return s.auditRepository.CancelAdminSubscription(ctx, in)
}

func (s *Service) ListUsers(ctx context.Context, input models.ListAdminUsersInput) (models.ListAdminUsersResult, error) {
	if s.auditRepository == nil || input.Limit < 1 || input.Limit > 100 || input.Offset < 0 || (input.Role != nil && !models.IsValidUserRole(*input.Role)) {
		return models.ListAdminUsersResult{}, models.ErrInvalidAdminInput
	}
	return s.auditRepository.ListAdminUsers(ctx, input)
}

func (s *Service) GetUser(ctx context.Context, userID uuid.UUID) (models.AdminUser, error) {
	if s.auditRepository == nil || userID == uuid.Nil {
		return models.AdminUser{}, models.ErrInvalidAdminInput
	}
	return s.auditRepository.GetAdminUserByUUID(ctx, userID)
}

func (s *Service) ChangeUserRole(ctx context.Context, input models.ChangeAdminUserRoleInput) (models.AdminUser, error) {
	if s.auditRepository == nil || input.ActorUserUUID == uuid.Nil || input.TargetUserUUID == uuid.Nil || !models.IsValidUserRole(input.ExpectedRole) || !models.IsValidUserRole(input.Role) || strings.TrimSpace(input.Metadata.Reason) == "" {
		return models.AdminUser{}, models.ErrInvalidAdminInput
	}
	if input.ActorUserUUID == input.TargetUserUUID {
		return models.AdminUser{}, models.ErrCannotChangeOwnRole
	}
	return s.auditRepository.ChangeAdminUserRole(ctx, input)
}

func (s *Service) ListUserSessions(ctx context.Context, actorUserID uuid.UUID, targetUserID uuid.UUID) ([]models.AdminUserSession, error) {
	if s.auditRepository == nil || actorUserID == uuid.Nil || targetUserID == uuid.Nil {
		return nil, models.ErrInvalidAdminInput
	}
	actor, err := s.auditRepository.GetAdminUserByUUID(ctx, actorUserID)
	if err != nil {
		return nil, err
	}
	target, err := s.auditRepository.GetAdminUserByUUID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}
	if actor.ID == target.ID {
		return nil, models.ErrAdminSessionManagementForbidden
	}
	if err := models.ValidateAdminSessionTarget(actor.Role, target.Role); err != nil {
		return nil, err
	}
	return s.auditRepository.ListAdminUserSessions(ctx, targetUserID)
}

func (s *Service) RevokeUserSession(ctx context.Context, input models.AdminSessionMutationInput) error {
	return s.validateSessionMutation(ctx, input, false)
}
func (s *Service) RevokeAllUserSessions(ctx context.Context, input models.AdminSessionMutationInput) error {
	return s.validateSessionMutation(ctx, input, true)
}
func (s *Service) validateSessionMutation(ctx context.Context, input models.AdminSessionMutationInput, all bool) error {
	if s.auditRepository == nil || input.ActorUserUUID == uuid.Nil || input.TargetUserUUID == uuid.Nil || (!all && input.SessionUUID == uuid.Nil) || strings.TrimSpace(input.Metadata.Reason) == "" {
		return models.ErrInvalidAdminInput
	}
	if input.ActorUserUUID == input.TargetUserUUID {
		return models.ErrAdminSessionManagementForbidden
	}
	if all {
		return s.auditRepository.RevokeAllAdminUserSessions(ctx, input)
	}
	return s.auditRepository.RevokeAdminUserSession(ctx, input)
}

type Service struct {
	auditRepository AuditRepository
	now             func() time.Time
	callReader      interface {
		GetByUUIDForProcessing(context.Context, uuid.UUID) (models.Call, error)
	}
	audioStorage storage.AudioStorage
}

func (s *Service) SetCallReader(reader interface {
	GetByUUIDForProcessing(context.Context, uuid.UUID) (models.Call, error)
})                                                                   { s.callReader = reader }
func (s *Service) SetAudioStorage(audioStorage storage.AudioStorage) { s.audioStorage = audioStorage }
func (s *Service) GetCall(ctx context.Context, id uuid.UUID) (models.Call, error) {
	if s.callReader == nil || id == uuid.Nil {
		return models.Call{}, models.ErrInvalidAdminInput
	}
	return s.callReader.GetByUUIDForProcessing(ctx, id)
}
func (s *Service) GetCallAudio(ctx context.Context, id uuid.UUID) (models.File, error) {
	call, err := s.GetCall(ctx, id)
	if err != nil {
		return models.File{}, err
	}
	if s.audioStorage == nil {
		return models.File{}, errAuditRepositoryNotConfigured
	}
	content, err := s.audioStorage.OpenReadSeeker(ctx, call.AudioPath)
	if err != nil {
		return models.File{}, err
	}
	return models.File{Content: content, ReadSeeker: content, Path: call.AudioPath, OriginalFilename: call.OriginalFilename, MimeType: call.MimeType, SizeBytes: call.SizeBytes}, nil
}

func NewService(auditRepository AuditRepository) *Service {
	return &Service{
		auditRepository: auditRepository,
		now:             func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) GetCapabilities(_ context.Context, role models.UserRole) (models.AdminCapabilities, error) {
	if !models.HasAdminPermission(role, models.AdminPermissionPanelAccess) {
		return models.AdminCapabilities{}, models.ErrForbidden
	}

	return models.AdminCapabilities{
		Role:        role,
		Permissions: models.AdminPermissionsForRole(role),
	}, nil
}

func (s *Service) RecordAudit(ctx context.Context, input models.CreateAdminAuditLogInput) (models.AdminAuditLog, error) {
	input.Action = strings.TrimSpace(input.Action)
	input.TargetType = strings.TrimSpace(input.TargetType)

	if input.ActorUserUUID == uuid.Nil ||
		!models.HasAdminPermission(input.ActorRole, models.AdminPermissionPanelAccess) ||
		input.Action == "" ||
		input.TargetType == "" ||
		!validJSONObject(input.BeforeData) ||
		!validJSONObject(input.AfterData) ||
		!validOptionalIPAddress(input.IPAddress) {
		return models.AdminAuditLog{}, models.ErrInvalidAdminInput
	}

	if s.auditRepository == nil {
		return models.AdminAuditLog{}, errAuditRepositoryNotConfigured
	}

	auditID, err := uuid.NewV7()
	if err != nil {
		return models.AdminAuditLog{}, err
	}

	return s.auditRepository.CreateAdminAuditLog(ctx, models.AdminAuditLog{
		ID:            auditID,
		ActorUserUUID: input.ActorUserUUID,
		ActorRole:     input.ActorRole,
		Action:        input.Action,
		TargetType:    input.TargetType,
		TargetUUID:    input.TargetUUID,
		BeforeData:    input.BeforeData,
		AfterData:     input.AfterData,
		Reason:        normalizeOptionalString(input.Reason),
		RequestID:     normalizeOptionalString(input.RequestID),
		IPAddress:     normalizeOptionalString(input.IPAddress),
		UserAgent:     normalizeOptionalString(input.UserAgent),
		CreatedAt:     s.now().UTC(),
	})
}

func validJSONObject(value json.RawMessage) bool {
	if len(value) == 0 {
		return true
	}

	var object map[string]any
	return json.Unmarshal(value, &object) == nil && object != nil
}

func validOptionalIPAddress(value *string) bool {
	if value == nil || strings.TrimSpace(*value) == "" {
		return true
	}

	return net.ParseIP(strings.TrimSpace(*value)) != nil
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
