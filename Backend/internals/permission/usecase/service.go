package usecase

import (
	"context"
	"errors"

	"backend/internals/permission/repository"
)

var (
	ErrPermissionNotFound = errors.New("permission not found")
	ErrRoleNotFound       = errors.New("role not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
)

type IPermissionService interface {
	// Role-based permission checking
	UserHasPermission(ctx context.Context, userID int64, permission string) (bool, error)
	UserHasRole(ctx context.Context, userID int64, roleName string) (bool, error)

	// Resource-level permission checking
	UserHasResourcePermission(ctx context.Context, userID int64, resourceType string, resourceID int64, permission string) (bool, error)

	// Getting permissions for a user
	GetUserPermissions(ctx context.Context, userID int64) ([]string, error)
	GetUserRoles(ctx context.Context, userID int64) ([]string, error)
}

type permissionService struct {
	repo repository.IPermissionRepository
}

func NewPermissionService(repo repository.IPermissionRepository) IPermissionService {
	return &permissionService{
		repo: repo,
	}
}

// =============================================
// Role-based Permission Checking
// =============================================

func (s *permissionService) UserHasPermission(ctx context.Context, userID int64, permission string) (bool, error) {
	// Check if user has permission through their roles
	return s.repo.UserHasPermission(ctx, userID, permission)
}

func (s *permissionService) UserHasRole(ctx context.Context, userID int64, roleName string) (bool, error) {
	// Get user roles and check if roleName matches any of them
	role, err := s.repo.GetRoleByName(ctx, roleName)
	if err != nil {
		return false, err
	}

	return s.repo.IsUserInRole(ctx, userID, role.ID)
}

// =============================================
// Resource-level Permission Checking
// =============================================

func (s *permissionService) UserHasResourcePermission(ctx context.Context, userID int64, resourceType string, resourceID int64, permission string) (bool, error) {
	// Check resource-level permission grants first (takes precedence)
	hasGrant, err := s.repo.CheckPermissionGrant(ctx, userID, resourceType, resourceID, permission)
	if err != nil {
		return false, err
	}

	if hasGrant {
		return true, nil
	}

	// Fall back to role-based permissions
	// This is a basic check - in real implementation, you might want to check
	// if the permission exists and is role-based
	return s.repo.UserHasPermission(ctx, userID, permission)
}

// =============================================
// Getting Permissions
// =============================================

func (s *permissionService) GetUserPermissions(ctx context.Context, userID int64) ([]string, error) {
	// Get all roles for user
	roles, err := s.repo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Collect all permissions from all roles
	permissionSet := make(map[string]bool)
	for _, role := range roles {
		permissions, err := s.repo.GetRolePermissions(ctx, role.ID)
		if err != nil {
			return nil, err
		}

		for _, perm := range permissions {
			permissionSet[perm.Name] = true
		}
	}

	// Convert map to slice
	permissions := make([]string, 0, len(permissionSet))
	for perm := range permissionSet {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

func (s *permissionService) GetUserRoles(ctx context.Context, userID int64) ([]string, error) {
	// Get all roles for user
	roles, err := s.repo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Convert to role names
	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	return roleNames, nil
}
