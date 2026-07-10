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
