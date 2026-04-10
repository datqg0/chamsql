-- =============================================
-- PERMISSIONS QUERIES
-- =============================================

-- name: GetPermissionByID :one
SELECT * FROM permissions WHERE id = $1;

-- name: GetPermissionByName :one
SELECT * FROM permissions WHERE name = $1;

-- name: ListPermissions :many
SELECT * FROM permissions ORDER BY category, name ASC;

-- name: ListPermissionsByCategory :many
SELECT * FROM permissions WHERE category = $1 ORDER BY name ASC;

-- name: CreatePermission :one
INSERT INTO permissions (name, description, category)
VALUES ($1, $2, $3)
RETURNING *;

-- name: DeletePermission :exec
DELETE FROM permissions WHERE id = $1;

-- =============================================
-- ROLES QUERIES
-- =============================================

-- name: GetRoleByID :one
SELECT * FROM roles WHERE id = $1;

-- name: GetRoleByName :one
SELECT * FROM roles WHERE name = $1;

-- name: ListRoles :many
SELECT * FROM roles ORDER BY name ASC;

-- name: CreateRole :one
INSERT INTO roles (name, description, is_system)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateRole :one
UPDATE roles SET description = COALESCE(sqlc.narg('description'), description)
WHERE id = $1
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles WHERE id = $1 AND is_system = false;

-- =============================================
-- ROLE PERMISSIONS QUERIES
-- =============================================

-- name: GetRolePermissions :many
SELECT p.* FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
WHERE rp.role_id = $1
ORDER BY p.category, p.name;

-- name: GrantPermissionToRole :one
INSERT INTO role_permissions (role_id, permission_id)
VALUES ($1, $2)
ON CONFLICT (role_id, permission_id) DO NOTHING
RETURNING *;

-- name: RevokePermissionFromRole :exec
DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2;

-- name: RoleHasPermission :one
SELECT EXISTS(
  SELECT 1 FROM role_permissions rp
  JOIN permissions p ON rp.permission_id = p.id
  WHERE rp.role_id = $1 AND p.name = $2
) AS has_permission;

-- =============================================
-- USER ROLES QUERIES
-- =============================================

-- name: GetUserRoles :many
SELECT r.* FROM roles r
JOIN user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = $1
ORDER BY r.name;

-- name: GetUserRoleIDs :many
SELECT role_id FROM user_roles WHERE user_id = $1;

-- name: GrantRoleToUser :one
INSERT INTO user_roles (user_id, role_id, assigned_by)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, role_id) DO NOTHING
RETURNING *;

-- name: RevokeRoleFromUser :exec
DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2;

-- name: IsUserInRole :one
SELECT EXISTS(
  SELECT 1 FROM user_roles
  WHERE user_id = $1 AND role_id = $2
) AS is_in_role;

-- name: UserHasPermission :one
SELECT EXISTS(
  SELECT 1 FROM user_roles ur
  JOIN role_permissions rp ON ur.role_id = rp.role_id
  JOIN permissions p ON rp.permission_id = p.id
  WHERE ur.user_id = $1 AND p.name = $2
) AS has_permission;

-- =============================================
-- PERMISSION GRANTS QUERIES
-- =============================================

-- name: GetPermissionGrant :one
SELECT * FROM permission_grants
WHERE user_id = $1 AND resource_type = $2 AND resource_id = $3 AND permission = $4;

-- name: ListUserPermissionGrants :many
SELECT * FROM permission_grants
WHERE user_id = $1
ORDER BY resource_type, resource_id, permission;

-- name: ListResourcePermissionGrants :many
SELECT * FROM permission_grants
WHERE resource_type = $1 AND resource_id = $2
ORDER BY user_id, permission;

-- name: CreatePermissionGrant :one
INSERT INTO permission_grants (user_id, resource_type, resource_id, permission, granted_by, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id, resource_type, resource_id, permission) DO UPDATE
SET expires_at = EXCLUDED.expires_at, granted_by = EXCLUDED.granted_by
RETURNING *;

-- name: RevokePermissionGrant :exec
DELETE FROM permission_grants
WHERE user_id = $1 AND resource_type = $2 AND resource_id = $3 AND permission = $4;

-- name: RevokeAllResourcePermissionGrants :exec
DELETE FROM permission_grants
WHERE resource_type = $1 AND resource_id = $2;

-- name: CheckPermissionGrant :one
SELECT EXISTS(
  SELECT 1 FROM permission_grants
  WHERE user_id = $1
    AND resource_type = $2
    AND resource_id = $3
    AND permission = $4
    AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
) AS has_grant;

-- name: CleanupExpiredPermissionGrants :exec
DELETE FROM permission_grants
WHERE expires_at IS NOT NULL AND expires_at <= CURRENT_TIMESTAMP;

-- =============================================
-- AUDIT LOG QUERIES
-- =============================================

-- name: CreateAuditLog :one
INSERT INTO audit_logs (user_id, action, resource_type, resource_id, old_value, new_value, reason, ip_address, user_agent)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: ListAuditLogs :many
SELECT * FROM audit_logs
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListUserAuditLogs :many
SELECT * FROM audit_logs
WHERE user_id = $1 OR old_value->>'user_id' = $2::text OR new_value->>'user_id' = $2::text
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListResourceAuditLogs :many
SELECT * FROM audit_logs
WHERE resource_type = $1 AND resource_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;
