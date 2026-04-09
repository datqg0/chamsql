package permissions

import (
	"context"
	"encoding/json"
	"fmt"

	"backend/db"
	"backend/sql/models"
)

type PermissionService interface {
	// Check if user has permission for resource+action
	HasPermission(ctx context.Context, userID int64, resourceType, action string) (bool, error)

	// Check if user owns a specific resource
	IsResourceOwner(ctx context.Context, userID int64, resourceType string, resourceID int64) (bool, error)

	// Check if user has access to a specific resource
	HasResourceAccess(ctx context.Context, userID int64, resourceType string, resourceID int64) (bool, error)

	// Combined: user has permission AND has access to resource
	CanAccess(ctx context.Context, userID int64, resourceType, action string, resourceID int64) (bool, error)

	// Role management
	GrantRoleToUser(ctx context.Context, userID int64, roleID int32, grantedBy int64) error
	RevokeRoleFromUser(ctx context.Context, userID int64, roleID int32) error
	GetUserRoles(ctx context.Context, userID int64) ([]models.Role, error)

	// Resource access management
	GrantResourceAccess(ctx context.Context, resourceType string, resourceID int64, userID int64, permissionType string, grantedBy int64) error
	RevokeResourceAccess(ctx context.Context, resourceType string, resourceID int64, userID int64) error

	// Audit logging
	LogPermissionChange(ctx context.Context, action string, targetUserID *int64, targetRoleID *int32, performedBy int64, details interface{}) error
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

// HasPermission checks if user's role has permission for resource+action
// Returns true if user has ANY role that has this permission
func (s *permissionService) HasPermission(ctx context.Context, userID int64, resourceType, action string) (bool, error) {
	hasPermission, err := s.queries.CheckUserHasPermission(ctx, models.CheckUserHasPermissionParams{
		UserID:       int32(userID),
		ResourceType: resourceType,
		Action:       action,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}
	return hasPermission, nil
}

// IsResourceOwner checks if user owns a specific resource (is "owner" in resource_access_control)
func (s *permissionService) IsResourceOwner(ctx context.Context, userID int64, resourceType string, resourceID int64) (bool, error) {
	isOwner, err := s.queries.CheckResourceOwner(ctx, models.CheckResourceOwnerParams{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		UserID:       int32(userID),
	})
	if err != nil {
		return false, fmt.Errorf("failed to check resource owner: %w", err)
	}
	return isOwner, nil
}

// HasResourceAccess checks if user has ANY level of access to resource (owner, editor, or viewer)
func (s *permissionService) HasResourceAccess(ctx context.Context, userID int64, resourceType string, resourceID int64) (bool, error) {
	hasAccess, err := s.queries.CheckResourceAccess(ctx, models.CheckResourceAccessParams{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		UserID:       int32(userID),
	})
	if err != nil {
		return false, fmt.Errorf("failed to check resource access: %w", err)
	}
	return hasAccess, nil
}

// CanAccess is the combined check: user has permission AND has access to the resource
// For basic resources without ownership (like public exams), just checks permission
// For owned resources, checks both permission and ownership/access
func (s *permissionService) CanAccess(ctx context.Context, userID int64, resourceType, action string, resourceID int64) (bool, error) {
	// First check: does user have the permission?
	hasPerm, err := s.HasPermission(ctx, userID, resourceType, action)
	if err != nil || !hasPerm {
		return false, err
	}

	// Second check: if resourceID provided, verify access to that specific resource
	if resourceID > 0 {
		hasAccess, err := s.HasResourceAccess(ctx, userID, resourceType, resourceID)
		if err != nil || !hasAccess {
			return false, err
		}
	}

	return true, nil
}

// GrantRoleToUser assigns a role to a user
func (s *permissionService) GrantRoleToUser(ctx context.Context, userID int64, roleID int32, grantedBy int64) error {
	grantedByInt32 := int32(grantedBy)
	_, err := s.queries.GrantRoleToUser(ctx, models.GrantRoleToUserParams{
		UserID:     int32(userID),
		RoleID:     roleID,
		AssignedBy: &grantedByInt32,
	})
	if err != nil {
		return fmt.Errorf("failed to grant role: %w", err)
	}

	// Log audit
	_ = s.LogPermissionChange(ctx, "role_assigned", ptrInt64(userID), &roleID, grantedBy, nil)
	return nil
}

// RevokeRoleFromUser removes a role from a user
func (s *permissionService) RevokeRoleFromUser(ctx context.Context, userID int64, roleID int32) error {
	err := s.queries.RevokeRoleFromUser(ctx, models.RevokeRoleFromUserParams{
		UserID: int32(userID),
		RoleID: roleID,
	})
	if err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}

	// Log audit (performed_by would be current admin, but we'll get it from context in handler)
	return nil
}

// GetUserRoles returns all roles assigned to a user
func (s *permissionService) GetUserRoles(ctx context.Context, userID int64) ([]models.Role, error) {
	roles, err := s.queries.GetUserRoles(ctx, int32(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	return roles, nil
}

// GrantResourceAccess grants access to a specific resource for a user
func (s *permissionService) GrantResourceAccess(ctx context.Context, resourceType string, resourceID int64, userID int64, permissionType string, grantedBy int64) error {
	_, err := s.queries.CreateResourceAccess(ctx, models.CreateResourceAccessParams{
		ResourceType:   resourceType,
		ResourceID:     resourceID,
		UserID:         int32(userID),
		PermissionType: permissionType,
		GrantedBy:      ptrInt32(int32(grantedBy)),
	})
	if err != nil {
		return fmt.Errorf("failed to grant resource access: %w", err)
	}

	// Log audit
	_ = s.LogPermissionChange(ctx, "resource_access_granted", ptrInt64(userID), nil, grantedBy, map[string]interface{}{
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"permission":    permissionType,
	})
	return nil
}

// RevokeResourceAccess removes access to a specific resource
func (s *permissionService) RevokeResourceAccess(ctx context.Context, resourceType string, resourceID int64, userID int64) error {
	err := s.queries.RevokeResourceAccess(ctx, models.RevokeResourceAccessParams{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		UserID:       int32(userID),
	})
	if err != nil {
		return fmt.Errorf("failed to revoke resource access: %w", err)
	}
	return nil
}

// LogPermissionChange creates an audit log entry for permission changes
func (s *permissionService) LogPermissionChange(ctx context.Context, action string, targetUserID *int64, targetRoleID *int32, performedBy int64, details interface{}) error {
	var detailsJSON []byte
	if details != nil {
		var err error
		detailsJSON, err = json.Marshal(details)
		if err != nil {
			detailsJSON = []byte("{}")
		}
	}

	var targetUser *int32
	if targetUserID != nil {
		u := int32(*targetUserID)
		targetUser = &u
	}

	_, err := s.queries.CreateAuditLog(ctx, models.CreateAuditLogParams{
		Action:       action,
		TargetUserID: targetUser,
		TargetRoleID: targetRoleID,
		PerformedBy:  int32(performedBy),
		Details:      detailsJSON,
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
