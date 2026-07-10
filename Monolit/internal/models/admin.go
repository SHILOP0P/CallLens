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
