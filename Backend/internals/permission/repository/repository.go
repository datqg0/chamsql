package repository

import (
	"context"

	"backend/db"
	"backend/sql/models"
)

// IPermissionRepository provides access to permission data
type IPermissionRepository interface {
	// =============================================
	// Permission Operations
	// =============================================
	GetPermissionByID(ctx context.Context, id int32) (*models.Permission, error)
	GetPermissionByName(ctx context.Context, name string) (*models.Permission, error)
	ListPermissions(ctx context.Context) ([]models.Permission, error)
	ListPermissionsByCategory(ctx context.Context, category string) ([]models.Permission, error)
	CreatePermission(ctx context.Context, params models.CreatePermissionParams) (*models.Permission, error)
	DeletePermission(ctx context.Context, id int32) error

	// =============================================
	// Role Operations
	// =============================================
	GetRoleByID(ctx context.Context, id int32) (*models.Role, error)
	GetRoleByName(ctx context.Context, name string) (*models.Role, error)
	ListRoles(ctx context.Context) ([]models.Role, error)
	CreateRole(ctx context.Context, params models.CreateRoleParams) (*models.Role, error)
	UpdateRole(ctx context.Context, params models.UpdateRoleParams) (*models.Role, error)
	DeleteRole(ctx context.Context, id int32) error

	// =============================================
	// Role-Permission Operations
	// =============================================
	GetRolePermissions(ctx context.Context, roleID int32) ([]models.Permission, error)
	GrantPermissionToRole(ctx context.Context, params models.GrantPermissionToRoleParams) (*models.RolePermission, error)
	RevokePermissionFromRole(ctx context.Context, roleID int32, permissionID int32) error
	RoleHasPermission(ctx context.Context, roleID int32, permissionName string) (bool, error)

	// =============================================
	// User-Role Operations
	// =============================================
	GetUserRoles(ctx context.Context, userID int64) ([]models.Role, error)
	GetUserRoleIDs(ctx context.Context, userID int64) ([]int32, error)
	GrantRoleToUser(ctx context.Context, params models.GrantRoleToUserParams) (*models.UserRole, error)
	RevokeRoleFromUser(ctx context.Context, userID int64, roleID int32) error
	IsUserInRole(ctx context.Context, userID int64, roleID int32) (bool, error)
	UserHasPermission(ctx context.Context, userID int64, permissionName string) (bool, error)

	// =============================================
	// Permission Grant Operations
	// =============================================
	GetPermissionGrant(ctx context.Context, params models.GetPermissionGrantParams) (*models.PermissionGrant, error)
	ListUserPermissionGrants(ctx context.Context, userID int64) ([]models.PermissionGrant, error)
	ListResourcePermissionGrants(ctx context.Context, resourceType string, resourceID int64) ([]models.PermissionGrant, error)
	CreatePermissionGrant(ctx context.Context, params models.CreatePermissionGrantParams) (*models.PermissionGrant, error)
	RevokePermissionGrant(ctx context.Context, userID int64, resourceType string, resourceID int64, permission string) error
	RevokeAllResourcePermissionGrants(ctx context.Context, resourceType string, resourceID int64) error
	CheckPermissionGrant(ctx context.Context, userID int64, resourceType string, resourceID int64, permission string) (bool, error)
	CleanupExpiredPermissionGrants(ctx context.Context) error

	// =============================================
	// Audit Log Operations
	// =============================================
	CreateAuditLog(ctx context.Context, params models.CreateAuditLogParams) (*models.AuditLog, error)
	ListAuditLogs(ctx context.Context, limit, offset int32) ([]models.AuditLog, error)
	ListUserAuditLogs(ctx context.Context, userID int64, limit, offset int32) ([]models.AuditLog, error)
	ListResourceAuditLogs(ctx context.Context, resourceType string, resourceID int64, limit, offset int32) ([]models.AuditLog, error)
}

type permissionRepository struct {
	db      *db.Database
	queries *models.Queries
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(database *db.Database) IPermissionRepository {
	return &permissionRepository{
		db:      database,
		queries: models.New(database.GetPool()),
	}
}

// =============================================
// Permission Operations
// =============================================

func (r *permissionRepository) GetPermissionByID(ctx context.Context, id int32) (*models.Permission, error) {
	permission, err := r.queries.GetPermissionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) GetPermissionByName(ctx context.Context, name string) (*models.Permission, error) {
	permission, err := r.queries.GetPermissionByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	return r.queries.ListPermissions(ctx)
}

func (r *permissionRepository) ListPermissionsByCategory(ctx context.Context, category string) ([]models.Permission, error) {
	return r.queries.ListPermissionsByCategory(ctx, &category)
}

func (r *permissionRepository) CreatePermission(ctx context.Context, params models.CreatePermissionParams) (*models.Permission, error) {
	permission, err := r.queries.CreatePermission(ctx, params)
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) DeletePermission(ctx context.Context, id int32) error {
	return r.queries.DeletePermission(ctx, id)
}

// =============================================
// Role Operations
// =============================================

func (r *permissionRepository) GetRoleByID(ctx context.Context, id int32) (*models.Role, error) {
	role, err := r.queries.GetRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *permissionRepository) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	role, err := r.queries.GetRoleByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *permissionRepository) ListRoles(ctx context.Context) ([]models.Role, error) {
	return r.queries.ListRoles(ctx)
}

func (r *permissionRepository) CreateRole(ctx context.Context, params models.CreateRoleParams) (*models.Role, error) {
	role, err := r.queries.CreateRole(ctx, params)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *permissionRepository) UpdateRole(ctx context.Context, params models.UpdateRoleParams) (*models.Role, error) {
	role, err := r.queries.UpdateRole(ctx, params)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *permissionRepository) DeleteRole(ctx context.Context, id int32) error {
	return r.queries.DeleteRole(ctx, id)
}

// =============================================
// Role-Permission Operations
// =============================================

func (r *permissionRepository) GetRolePermissions(ctx context.Context, roleID int32) ([]models.Permission, error) {
	return r.queries.GetRolePermissions(ctx, roleID)
}

func (r *permissionRepository) GrantPermissionToRole(ctx context.Context, params models.GrantPermissionToRoleParams) (*models.RolePermission, error) {
	rp, err := r.queries.GrantPermissionToRole(ctx, params)
	if err != nil {
		return nil, err
	}
	return &rp, nil
}

func (r *permissionRepository) RevokePermissionFromRole(ctx context.Context, roleID int32, permissionID int32) error {
	return r.queries.RevokePermissionFromRole(ctx, models.RevokePermissionFromRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
}

func (r *permissionRepository) RoleHasPermission(ctx context.Context, roleID int32, permissionName string) (bool, error) {
	result, err := r.queries.RoleHasPermission(ctx, models.RoleHasPermissionParams{
		RoleID: roleID,
		Name:   permissionName,
	})
	if err != nil {
		return false, err
	}
	return result, nil
}

// =============================================
// User-Role Operations
// =============================================

func (r *permissionRepository) GetUserRoles(ctx context.Context, userID int64) ([]models.Role, error) {
	return r.queries.GetUserRoles(ctx, userID)
}

func (r *permissionRepository) GetUserRoleIDs(ctx context.Context, userID int64) ([]int32, error) {
	return r.queries.GetUserRoleIDs(ctx, userID)
}

func (r *permissionRepository) GrantRoleToUser(ctx context.Context, params models.GrantRoleToUserParams) (*models.UserRole, error) {
	ur, err := r.queries.GrantRoleToUser(ctx, params)
	if err != nil {
		return nil, err
	}
	return &ur, nil
}

func (r *permissionRepository) RevokeRoleFromUser(ctx context.Context, userID int64, roleID int32) error {
	return r.queries.RevokeRoleFromUser(ctx, models.RevokeRoleFromUserParams{
		UserID: userID,
		RoleID: roleID,
	})
}

func (r *permissionRepository) IsUserInRole(ctx context.Context, userID int64, roleID int32) (bool, error) {
	result, err := r.queries.IsUserInRole(ctx, models.IsUserInRoleParams{
		UserID: userID,
		RoleID: roleID,
	})
	if err != nil {
		return false, err
	}
	return result, nil
}

func (r *permissionRepository) UserHasPermission(ctx context.Context, userID int64, permissionName string) (bool, error) {
	result, err := r.queries.UserHasPermission(ctx, models.UserHasPermissionParams{
		UserID: userID,
		Name:   permissionName,
	})
	if err != nil {
		return false, err
	}
	return result, nil
}

// =============================================
// Permission Grant Operations
// =============================================

func (r *permissionRepository) GetPermissionGrant(ctx context.Context, params models.GetPermissionGrantParams) (*models.PermissionGrant, error) {
	grant, err := r.queries.GetPermissionGrant(ctx, params)
	if err != nil {
		return nil, err
	}
	return &grant, nil
}

func (r *permissionRepository) ListUserPermissionGrants(ctx context.Context, userID int64) ([]models.PermissionGrant, error) {
	return r.queries.ListUserPermissionGrants(ctx, userID)
}

func (r *permissionRepository) ListResourcePermissionGrants(ctx context.Context, resourceType string, resourceID int64) ([]models.PermissionGrant, error) {
	return r.queries.ListResourcePermissionGrants(ctx, models.ListResourcePermissionGrantsParams{
		ResourceType: resourceType,
		ResourceID:   resourceID,
	})
}

func (r *permissionRepository) CreatePermissionGrant(ctx context.Context, params models.CreatePermissionGrantParams) (*models.PermissionGrant, error) {
	grant, err := r.queries.CreatePermissionGrant(ctx, params)
	if err != nil {
		return nil, err
	}
	return &grant, nil
}

func (r *permissionRepository) RevokePermissionGrant(ctx context.Context, userID int64, resourceType string, resourceID int64, permission string) error {
	return r.queries.RevokePermissionGrant(ctx, models.RevokePermissionGrantParams{
		UserID:       userID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Permission:   permission,
	})
}

func (r *permissionRepository) RevokeAllResourcePermissionGrants(ctx context.Context, resourceType string, resourceID int64) error {
	return r.queries.RevokeAllResourcePermissionGrants(ctx, models.RevokeAllResourcePermissionGrantsParams{
		ResourceType: resourceType,
		ResourceID:   resourceID,
	})
}

func (r *permissionRepository) CheckPermissionGrant(ctx context.Context, userID int64, resourceType string, resourceID int64, permission string) (bool, error) {
	result, err := r.queries.CheckPermissionGrant(ctx, models.CheckPermissionGrantParams{
		UserID:       userID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Permission:   permission,
	})
	if err != nil {
		return false, err
	}
	return result, nil
}

func (r *permissionRepository) CleanupExpiredPermissionGrants(ctx context.Context) error {
	return r.queries.CleanupExpiredPermissionGrants(ctx)
}

// =============================================
// Audit Log Operations
// =============================================

func (r *permissionRepository) CreateAuditLog(ctx context.Context, params models.CreateAuditLogParams) (*models.AuditLog, error) {
	log, err := r.queries.CreateAuditLog(ctx, params)
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *permissionRepository) ListAuditLogs(ctx context.Context, limit, offset int32) ([]models.AuditLog, error) {
	return r.queries.ListAuditLogs(ctx, models.ListAuditLogsParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *permissionRepository) ListUserAuditLogs(ctx context.Context, userID int64, limit, offset int32) ([]models.AuditLog, error) {
	return r.queries.ListUserAuditLogs(ctx, models.ListUserAuditLogsParams{
		UserID:  ptrInt64(userID),
		Column2: "", // Filter column - empty for all logs
		Limit:   limit,
		Offset:  offset,
	})
}

func (r *permissionRepository) ListResourceAuditLogs(ctx context.Context, resourceType string, resourceID int64, limit, offset int32) ([]models.AuditLog, error) {
	return r.queries.ListResourceAuditLogs(ctx, models.ListResourceAuditLogsParams{
		ResourceType: ptrStr(resourceType),
		ResourceID:   ptrInt64(resourceID),
		Limit:        limit,
		Offset:       offset,
	})
}

// Helper functions
func ptrStr(s string) *string {
	return &s
}

func ptrInt64(i int64) *int64 {
	return &i
}
