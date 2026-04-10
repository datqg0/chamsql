package permissions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"backend/db"
	"backend/sql/models"
	"github.com/jackc/pgx/v5/pgtype"
)

type PermissionService interface {
	// Check if user has permission for resource+action
	HasPermission(ctx context.Context, userID int64, permission string) (bool, error)

	// Check if user has specific resource-level permission
	HasResourcePermission(ctx context.Context, userID int64, resourceType string, resourceID int64, permission string) (bool, error)

	// Combined: user has permission AND has access to resource
	CanAccess(ctx context.Context, userID int64, resourceType, action string, resourceID int64) (bool, error)

	// Role management
	GrantRoleToUser(ctx context.Context, userID int64, roleID int32, grantedBy int64) error
	RevokeRoleFromUser(ctx context.Context, userID int64, roleID int32) error
	GetUserRoles(ctx context.Context, userID int64) ([]models.Role, error)

	// Permission grant management (resource-level)
	GrantResourcePermission(ctx context.Context, resourceType string, resourceID int64, userID int64, permission string, grantedBy int64, expiresAt *time.Time) error
	RevokeResourcePermission(ctx context.Context, resourceType string, resourceID int64, userID int64, permission string) error

	// Audit logging
	LogPermissionChange(ctx context.Context, action string, userID *int64, details interface{}) error
}

type permissionService struct {
	db      *db.Database
	queries *models.Queries
}

func NewPermissionService(database *db.Database) PermissionService {
	return &permissionService{
		db:      database,
		queries: models.New(database.GetPool()),
	}
}

// HasPermission checks if user's role has the specified permission
func (s *permissionService) HasPermission(ctx context.Context, userID int64, permission string) (bool, error) {
	hasPermission, err := s.queries.UserHasPermission(ctx, models.UserHasPermissionParams{
		UserID: userID,
		Name:   permission,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}
	return hasPermission, nil
}

// HasResourcePermission checks if user has permission to a specific resource
func (s *permissionService) HasResourcePermission(ctx context.Context, userID int64, resourceType string, resourceID int64, permission string) (bool, error) {
	// Check resource-level permission grant first
	hasGrant, err := s.queries.CheckPermissionGrant(ctx, models.CheckPermissionGrantParams{
		UserID:       userID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Permission:   permission,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check permission grant: %w", err)
	}

	if hasGrant {
		return true, nil
	}

	// Fall back to role-based permission
	return s.HasPermission(ctx, userID, permission)
}

// CanAccess is the combined check: user has permission AND has access to the resource
func (s *permissionService) CanAccess(ctx context.Context, userID int64, resourceType, action string, resourceID int64) (bool, error) {
	// Check if user has the specific resource permission
	return s.HasResourcePermission(ctx, userID, resourceType, resourceID, action)
}

// GrantRoleToUser assigns a role to a user
func (s *permissionService) GrantRoleToUser(ctx context.Context, userID int64, roleID int32, grantedBy int64) error {
	_, err := s.queries.GrantRoleToUser(ctx, models.GrantRoleToUserParams{
		UserID:     userID,
		RoleID:     roleID,
		AssignedBy: ptrInt64(grantedBy),
	})
	if err != nil {
		return fmt.Errorf("failed to grant role: %w", err)
	}

	// Log audit
	_ = s.LogPermissionChange(ctx, "role_assigned", ptrInt64(userID), map[string]interface{}{
		"role_id":    roleID,
		"granted_by": grantedBy,
	})
	return nil
}

// RevokeRoleFromUser removes a role from a user
func (s *permissionService) RevokeRoleFromUser(ctx context.Context, userID int64, roleID int32) error {
	err := s.queries.RevokeRoleFromUser(ctx, models.RevokeRoleFromUserParams{
		UserID: userID,
		RoleID: roleID,
	})
	if err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}

	// Log audit
	_ = s.LogPermissionChange(ctx, "role_revoked", ptrInt64(userID), map[string]interface{}{
		"role_id": roleID,
	})
	return nil
}

// GetUserRoles returns all roles assigned to a user
func (s *permissionService) GetUserRoles(ctx context.Context, userID int64) ([]models.Role, error) {
	roles, err := s.queries.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	return roles, nil
}

// GrantResourcePermission grants permission to a specific resource for a user
func (s *permissionService) GrantResourcePermission(ctx context.Context, resourceType string, resourceID int64, userID int64, permission string, grantedBy int64, expiresAt *time.Time) error {
	var expiresAtTS pgtype.Timestamptz
	if expiresAt != nil {
		expiresAtTS = pgtype.Timestamptz{Time: *expiresAt, Valid: true}
	}

	_, err := s.queries.CreatePermissionGrant(ctx, models.CreatePermissionGrantParams{
		UserID:       userID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Permission:   permission,
		GrantedBy:    grantedBy,
		ExpiresAt:    expiresAtTS,
	})
	if err != nil {
		return fmt.Errorf("failed to grant resource permission: %w", err)
	}

	// Log audit
	_ = s.LogPermissionChange(ctx, "resource_permission_granted", ptrInt64(userID), map[string]interface{}{
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"permission":    permission,
	})
	return nil
}

// RevokeResourcePermission removes permission from a specific resource
func (s *permissionService) RevokeResourcePermission(ctx context.Context, resourceType string, resourceID int64, userID int64, permission string) error {
	err := s.queries.RevokePermissionGrant(ctx, models.RevokePermissionGrantParams{
		UserID:       userID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Permission:   permission,
	})
	if err != nil {
		return fmt.Errorf("failed to revoke resource permission: %w", err)
	}

	// Log audit
	_ = s.LogPermissionChange(ctx, "resource_permission_revoked", ptrInt64(userID), map[string]interface{}{
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"permission":    permission,
	})
	return nil
}

// LogPermissionChange creates an audit log entry for permission changes
func (s *permissionService) LogPermissionChange(ctx context.Context, action string, userID *int64, details interface{}) error {
	var detailsJSON []byte
	if details != nil {
		var err error
		detailsJSON, err = json.Marshal(details)
		if err != nil {
			detailsJSON = []byte("{}")
		}
	}

	_, err := s.queries.CreateAuditLog(ctx, models.CreateAuditLogParams{
		UserID:       userID,
		Action:       action,
		ResourceType: nil,
		ResourceID:   nil,
		OldValue:     nil,
		NewValue:     detailsJSON,
		Reason:       nil,
		IpAddress:    nil,
		UserAgent:    nil,
	})
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	return nil
}

// Helper functions
func ptrInt64(v int64) *int64 {
	return &v
}

func ptrInt32(v int32) *int32 {
	return &v
}
