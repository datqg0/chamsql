-- =============================================
-- ROLES QUERIES
-- =============================================

-- name: GetRoleByID :one
SELECT * FROM roles WHERE id = $1;

-- name: GetRoleByName :one
SELECT * FROM roles WHERE name = $1;

-- name: ListRoles :many
SELECT * FROM roles ORDER BY id ASC;

-- name: CreateRole :one
INSERT INTO roles (name, description, is_extensible)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateRole :one
UPDATE roles SET name = COALESCE(sqlc.narg('name'), name),
                 description = COALESCE(sqlc.narg('description'), description),
                 is_extensible = COALESCE(sqlc.narg('is_extensible'), is_extensible),
                 updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles WHERE id = $1;

-- =============================================
-- PERMISSIONS QUERIES
-- =============================================

-- name: GetPermissionByID :one
SELECT * FROM permissions WHERE id = $1;

-- name: ListPermissions :many
SELECT * FROM permissions ORDER BY resource_type, action ASC;

-- name: ListPermissionsByResourceType :many
SELECT * FROM permissions WHERE resource_type = $1 ORDER BY action ASC;

-- name: CreatePermission :one
INSERT INTO permissions (resource_type, action, description)
VALUES ($1, $2, $3)
RETURNING *;

-- name: DeletePermission :exec
DELETE FROM permissions WHERE id = $1;

-- =============================================
-- ROLE PERMISSIONS QUERIES
-- =============================================

-- name: GetRolePermissions :many
SELECT p.* FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
WHERE rp.role_id = $1
ORDER BY p.resource_type, p.action;

-- name: GrantPermissionToRole :one
INSERT INTO role_permissions (role_id, permission_id)
VALUES ($1, $2)
ON CONFLICT (role_id, permission_id) DO NOTHING
RETURNING *;

-- name: RevokePermissionFromRole :exec
DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2;

-- name: HasPermission :one
SELECT EXISTS(
  SELECT 1 FROM role_permissions rp
  JOIN permissions p ON rp.permission_id = p.id
  WHERE rp.role_id = $1
    AND p.resource_type = $2
    AND p.action = $3
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

-- name: CheckUserHasPermission :one
SELECT EXISTS(
  SELECT 1 FROM user_roles ur
  JOIN role_permissions rp ON ur.role_id = rp.role_id
  JOIN permissions p ON rp.permission_id = p.id
  WHERE ur.user_id = $1
    AND p.resource_type = $2
    AND p.action = $3
) AS has_permission;

-- =============================================
-- RESOURCE ACCESS CONTROL QUERIES
-- =============================================

-- name: CreateResourceAccess :one
INSERT INTO resource_access_control (resource_type, resource_id, user_id, permission_type, granted_by)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (resource_type, resource_id, user_id, permission_type) DO NOTHING
RETURNING *;

-- name: GetResourceAccess :many
SELECT * FROM resource_access_control
WHERE resource_type = $1 AND resource_id = $2
ORDER BY permission_type, user_id;

-- name: GetUserResourceAccess :many
SELECT * FROM resource_access_control
WHERE user_id = $1 AND resource_type = $2
ORDER BY permission_type;

-- name: CheckResourceAccess :one
SELECT EXISTS(
  SELECT 1 FROM resource_access_control
  WHERE resource_type = $1
    AND resource_id = $2
    AND user_id = $3
    AND permission_type IN ('owner', 'editor', 'viewer')
) AS has_access;

-- name: CheckResourceOwner :one
SELECT EXISTS(
  SELECT 1 FROM resource_access_control
  WHERE resource_type = $1
    AND resource_id = $2
    AND user_id = $3
    AND permission_type = 'owner'
) AS is_owner;

-- name: RevokeResourceAccess :exec
DELETE FROM resource_access_control
WHERE resource_type = $1 AND resource_id = $2 AND user_id = $3;

-- name: RevokeAllResourceAccess :exec
DELETE FROM resource_access_control
WHERE resource_type = $1 AND resource_id = $2;

-- =============================================
-- AUDIT LOG QUERIES
-- =============================================

-- name: CreateAuditLog :one
INSERT INTO permission_audit_log (action, target_user_id, target_role_id, target_resource_type, target_resource_id, performed_by, details)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetAuditLog :many
SELECT * FROM permission_audit_log
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetUserAuditLog :many
SELECT * FROM permission_audit_log
WHERE target_user_id = $1 OR performed_by = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
