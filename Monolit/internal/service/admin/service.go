package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strings"
	"time"

	"calllens/monolit/internal/models"

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
