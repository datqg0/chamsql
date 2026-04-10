package usecase

import (
	"context"
	"errors"
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
	// Will be implemented after DB migrations and SQLC code generation
}

func NewPermissionService() IPermissionService {
	return &permissionService{}
}

// =============================================
// Role-based Permission Checking
// =============================================

func (s *permissionService) UserHasPermission(ctx context.Context, userID int64, permission string) (bool, error) {
	// TODO: Implement after SQLC models are generated
	// This will check if user's role has the permission
	return true, nil
}

func (s *permissionService) UserHasRole(ctx context.Context, userID int64, roleName string) (bool, error) {
	// TODO: Implement after SQLC models are generated
	return true, nil
}

// =============================================
// Resource-level Permission Checking
// =============================================

func (s *permissionService) UserHasResourcePermission(ctx context.Context, userID int64, resourceType string, resourceID int64, permission string) (bool, error) {
	// TODO: Implement after SQLC models are generated
	// This will check:
	// 1. Direct permission grants in permission_grants table
	// 2. Role-based permissions
	return true, nil
}

// =============================================
// Getting Permissions
// =============================================

func (s *permissionService) GetUserPermissions(ctx context.Context, userID int64) ([]string, error) {
	// TODO: Implement after SQLC models are generated
	return []string{}, nil
}

func (s *permissionService) GetUserRoles(ctx context.Context, userID int64) ([]string, error) {
	// TODO: Implement after SQLC models are generated
	return []string{}, nil
}
