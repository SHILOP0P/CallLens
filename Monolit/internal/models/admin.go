package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AdminPermission string

const (
	AdminPermissionPanelAccess         AdminPermission = "admin.panel.access"
	AdminPermissionUsersRead           AdminPermission = "admin.users.read"
	AdminPermissionCompaniesRead       AdminPermission = "admin.companies.read"
	AdminPermissionRolesManageHelpers  AdminPermission = "admin.roles.manage_helpers"
	AdminPermissionRolesManageAdmins   AdminPermission = "admin.roles.manage_admins"
	AdminPermissionSessionsRead        AdminPermission = "admin.sessions.read"
	AdminPermissionSessionsManage      AdminPermission = "admin.sessions.manage"
	AdminPermissionSubscriptionsRead   AdminPermission = "admin.subscriptions.read"
	AdminPermissionSubscriptionsManage AdminPermission = "admin.subscriptions.manage"
	AdminPermissionCallsRead           AdminPermission = "admin.calls.read"
	AdminPermissionMonitoringRead      AdminPermission = "admin.monitoring.read"
	AdminPermissionDashboardRead       AdminPermission = "admin.dashboard.read"
	AdminPermissionAuditRead           AdminPermission = "admin.audit.read"
)

var adminPermissionsByRole = map[UserRole][]AdminPermission{
	UserRoleHelper: {
		AdminPermissionPanelAccess,
		AdminPermissionUsersRead,
		AdminPermissionCompaniesRead,
		AdminPermissionSubscriptionsRead,
	},
	UserRoleAdmin: {
		AdminPermissionPanelAccess,
		AdminPermissionUsersRead,
		AdminPermissionCompaniesRead,
		AdminPermissionRolesManageHelpers,
		AdminPermissionSessionsRead,
		AdminPermissionSessionsManage,
		AdminPermissionSubscriptionsRead,
		AdminPermissionSubscriptionsManage,
		AdminPermissionCallsRead,
		AdminPermissionMonitoringRead,
		AdminPermissionDashboardRead,
		AdminPermissionAuditRead,
	},
	UserRoleSuperAdmin: {
		AdminPermissionPanelAccess,
		AdminPermissionUsersRead,
		AdminPermissionCompaniesRead,
		AdminPermissionRolesManageHelpers,
		AdminPermissionRolesManageAdmins,
		AdminPermissionSessionsRead,
		AdminPermissionSessionsManage,
		AdminPermissionSubscriptionsRead,
		AdminPermissionSubscriptionsManage,
		AdminPermissionCallsRead,
		AdminPermissionMonitoringRead,
		AdminPermissionDashboardRead,
		AdminPermissionAuditRead,
	},
}

type AdminCapabilities struct {
	Role        UserRole
	Permissions []AdminPermission
}

func AdminPermissionsForRole(role UserRole) []AdminPermission {
	permissions := adminPermissionsByRole[role]
	return append([]AdminPermission(nil), permissions...)
}

func HasAdminPermission(role UserRole, permission AdminPermission) bool {
	for _, candidate := range adminPermissionsByRole[role] {
		if candidate == permission {
			return true
		}
	}

	return false
}

type AdminAuditLog struct {
	ID            uuid.UUID
	ActorUserUUID uuid.UUID
	ActorRole     UserRole
	Action        string
	TargetType    string
	TargetUUID    uuid.NullUUID
	BeforeData    json.RawMessage
	AfterData     json.RawMessage
	Reason        *string
	RequestID     *string
	IPAddress     *string
	UserAgent     *string
	CreatedAt     time.Time
}

type CreateAdminAuditLogInput struct {
	ActorUserUUID uuid.UUID
	ActorRole     UserRole
	Action        string
	TargetType    string
	TargetUUID    uuid.NullUUID
	BeforeData    json.RawMessage
	AfterData     json.RawMessage
	Reason        *string
	RequestID     *string
	IPAddress     *string
	UserAgent     *string
}

type AdminSubscriptionStatusFilter string

const (
	AdminSubscriptionStatusActive   AdminSubscriptionStatusFilter = "active"
	AdminSubscriptionStatusCanceled AdminSubscriptionStatusFilter = "canceled"
	AdminSubscriptionStatusExpired  AdminSubscriptionStatusFilter = "expired"
	AdminSubscriptionStatusNone     AdminSubscriptionStatusFilter = "none"
)

type AdminUser struct {
	ID          uuid.UUID
	Email       string
	FullName    string
	FullSurname string
	Username    string
	Role        UserRole
	Post        *string
	Phone       *string
	Timezone    *string
	CreatedAt   time.Time
}

type ListAdminUsersInput struct {
	Query              string
	Role               *UserRole
	SubscriptionStatus *AdminSubscriptionStatusFilter
	PlanCode           *PlanCode
	CreatedFrom        *time.Time
	CreatedTo          *time.Time
	Limit              int
	Offset             int
}

type ListAdminUsersResult struct {
	Users  []AdminUser
	Total  int
	Limit  int
	Offset int
}

type AdminMutationMetadata struct {
	Reason    string
	RequestID *string
	IPAddress *string
	UserAgent *string
}

type ChangeAdminUserRoleInput struct {
	ActorUserUUID  uuid.UUID
	TargetUserUUID uuid.UUID
	ExpectedRole   UserRole
	Role           UserRole
	Metadata       AdminMutationMetadata
}

type AdminUserSession struct {
	ID         uuid.UUID
	UserAgent  *string
	IPAddress  *string
	CreatedAt  time.Time
	LastSeenAt *time.Time
	ExpiresAt  time.Time
}

type AdminUserSessions struct {
	UserUUID uuid.UUID
	Sessions []AdminUserSession
}

type AdminSessionMutationInput struct {
	ActorUserUUID  uuid.UUID
	TargetUserUUID uuid.UUID
	SessionUUID    uuid.UUID
	Metadata       AdminMutationMetadata
}

func IsValidUserRole(role UserRole) bool {
	switch role {
	case UserRoleUser, UserRoleHelper, UserRoleAdmin, UserRoleSuperAdmin:
		return true
	default:
		return false
	}
}

func ValidateAdminRoleTransition(actorRole UserRole, targetRole UserRole, requestedRole UserRole) error {
	if targetRole == UserRoleSuperAdmin || requestedRole == UserRoleSuperAdmin {
		return ErrProtectedSuperAdmin
	}
	if targetRole == requestedRole {
		return ErrRoleTransitionForbidden
	}

	switch actorRole {
	case UserRoleAdmin:
		if (targetRole == UserRoleUser && requestedRole == UserRoleHelper) ||
			(targetRole == UserRoleHelper && requestedRole == UserRoleUser) {
			return nil
		}
	case UserRoleSuperAdmin:
		if (targetRole == UserRoleUser || targetRole == UserRoleHelper || targetRole == UserRoleAdmin) &&
			(requestedRole == UserRoleUser || requestedRole == UserRoleHelper || requestedRole == UserRoleAdmin) {
			return nil
		}
	}

	return ErrRoleTransitionForbidden
}

func ValidateAdminSessionTarget(actorRole UserRole, targetRole UserRole) error {
	if targetRole == UserRoleSuperAdmin {
		return ErrProtectedSuperAdmin
	}

	switch actorRole {
	case UserRoleAdmin:
		if targetRole == UserRoleUser || targetRole == UserRoleHelper {
			return nil
		}
	case UserRoleSuperAdmin:
		if targetRole == UserRoleUser || targetRole == UserRoleHelper || targetRole == UserRoleAdmin {
			return nil
		}
	}

	return ErrAdminSessionManagementForbidden
}
