package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdminPermissionMatrix(t *testing.T) {
	tests := []struct {
		name        string
		role        UserRole
		allowed     []AdminPermission
		denied      []AdminPermission
		permissions int
	}{
		{
			name:        "user",
			role:        UserRoleUser,
			denied:      []AdminPermission{AdminPermissionPanelAccess, AdminPermissionUsersRead},
			permissions: 0,
		},
		{
			name: "helper",
			role: UserRoleHelper,
			allowed: []AdminPermission{
				AdminPermissionPanelAccess,
				AdminPermissionUsersRead,
				AdminPermissionCompaniesRead,
				AdminPermissionSubscriptionsRead,
			},
			denied: []AdminPermission{
				AdminPermissionRolesManageHelpers,
				AdminPermissionSessionsManage,
				AdminPermissionSubscriptionsManage,
				AdminPermissionCallsRead,
				AdminPermissionMonitoringRead,
				AdminPermissionDashboardRead,
				AdminPermissionAuditRead,
			},
			permissions: 4,
		},
		{
			name: "admin",
			role: UserRoleAdmin,
			allowed: []AdminPermission{
				AdminPermissionPanelAccess,
				AdminPermissionRolesManageHelpers,
				AdminPermissionSessionsManage,
				AdminPermissionSubscriptionsManage,
				AdminPermissionCallsRead,
				AdminPermissionMonitoringRead,
				AdminPermissionDashboardRead,
				AdminPermissionAuditRead,
			},
			denied:      []AdminPermission{AdminPermissionRolesManageAdmins},
			permissions: 12,
		},
		{
			name: "superadmin",
			role: UserRoleSuperAdmin,
			allowed: []AdminPermission{
				AdminPermissionPanelAccess,
				AdminPermissionRolesManageHelpers,
				AdminPermissionRolesManageAdmins,
				AdminPermissionSessionsManage,
				AdminPermissionSubscriptionsManage,
				AdminPermissionCallsRead,
				AdminPermissionMonitoringRead,
				AdminPermissionDashboardRead,
				AdminPermissionAuditRead,
			},
			permissions: 13,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permissions := AdminPermissionsForRole(tt.role)
			require.Len(t, permissions, tt.permissions)

			for _, permission := range tt.allowed {
				require.True(t, HasAdminPermission(tt.role, permission), "permission %q", permission)
				require.Contains(t, permissions, permission)
			}

			for _, permission := range tt.denied {
				require.False(t, HasAdminPermission(tt.role, permission), "permission %q", permission)
				require.NotContains(t, permissions, permission)
			}
		})
	}
}

func TestAdminPermissionsForRoleReturnsCopy(t *testing.T) {
	permissions := AdminPermissionsForRole(UserRoleHelper)
	require.NotEmpty(t, permissions)
	permissions[0] = AdminPermissionAuditRead

	require.Equal(t, AdminPermissionPanelAccess, AdminPermissionsForRole(UserRoleHelper)[0])
}
