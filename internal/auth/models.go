package auth

import "time"

// UserRole represents a role in the system.
type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleEditor   UserRole = "editor"
	RoleViewer   UserRole = "viewer"
)

// Permission represents a specific permission.
type Permission string

const (
	// Team permissions
	PermissionViewTeam     Permission = "team:view"
	PermissionEditTeam     Permission = "team:edit"
	PermissionDeleteTeam   Permission = "team:delete"
	PermissionManageUsers  Permission = "team:manage_users"
	PermissionUpdateBudget Permission = "team:update_budget"

	// Analytics permissions
	PermissionViewAnalytics    Permission = "analytics:view"
	PermissionExportAnalytics  Permission = "analytics:export"
	PermissionCreateReports    Permission = "analytics:create_reports"

	// Audit permissions
	PermissionViewAuditLogs Permission = "audit:view"

	// Filter permissions
	PermissionCreateFilter Permission = "filters:create"
	PermissionEditFilter   Permission = "filters:edit"
	PermissionDeleteFilter Permission = "filters:delete"
)

// RolePermissions defines which permissions each role has.
var RolePermissions = map[UserRole][]Permission{
	RoleAdmin: {
		PermissionViewTeam,
		PermissionEditTeam,
		PermissionDeleteTeam,
		PermissionManageUsers,
		PermissionUpdateBudget,
		PermissionViewAnalytics,
		PermissionExportAnalytics,
		PermissionCreateReports,
		PermissionViewAuditLogs,
		PermissionCreateFilter,
		PermissionEditFilter,
		PermissionDeleteFilter,
	},
	RoleEditor: {
		PermissionViewTeam,
		PermissionViewAnalytics,
		PermissionExportAnalytics,
		PermissionCreateReports,
		PermissionCreateFilter,
		PermissionEditFilter,
		PermissionDeleteFilter,
	},
	RoleViewer: {
		PermissionViewTeam,
		PermissionViewAnalytics,
		PermissionExportAnalytics,
	},
}

// AuthContext represents the authenticated user context.
type AuthContext struct {
	UserID  string
	TeamID  string
	Role    UserRole
	Email   string
	Name    string
	IsActive bool
	ExpiresAt time.Time
}

// HasPermission checks if the user has the required permission.
func (ac *AuthContext) HasPermission(perm Permission) bool {
	permissions, ok := RolePermissions[ac.Role]
	if !ok {
		return false
	}

	for _, p := range permissions {
		if p == perm {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if the user has any of the provided permissions.
func (ac *AuthContext) HasAnyPermission(perms ...Permission) bool {
	for _, perm := range perms {
		if ac.HasPermission(perm) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if the user has all of the provided permissions.
func (ac *AuthContext) HasAllPermissions(perms ...Permission) bool {
	for _, perm := range perms {
		if !ac.HasPermission(perm) {
			return false
		}
	}
	return true
}

// IsValid checks if the auth context is still valid.
func (ac *AuthContext) IsValid() bool {
	return ac.IsActive && time.Now().Before(ac.ExpiresAt)
}

// UserSession represents a user session record.
type UserSession struct {
	ID        string
	UserID    string
	TeamID    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
	LastUsedAt time.Time
}
