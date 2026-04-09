package http

import (
	"strconv"

	"backend/internals/admin/controller/dto"
	"backend/internals/admin/usecase"
	"backend/pkgs/response"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	usecase usecase.IAdminUseCase
}

func NewAdminHandler(uc usecase.IAdminUseCase) *AdminHandler {
	return &AdminHandler{usecase: uc}
}

// ImportUsers godoc
// @Summary     Import users from JSON
// @Tags        Admin
// @Accept      json
// @Produce     json
// @Param       request body dto.ImportUsersRequest true "Users to import"
// @Success     200 {object} dto.ImportResult
// @Router      /admin/users/import [post]
func (h *AdminHandler) ImportUsers(c *gin.Context) {
	var req dto.ImportUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.usecase.ImportUsers(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// GetSystemStats godoc
// @Summary     Get system statistics
// @Tags        Admin
// @Produce     json
// @Success     200 {object} dto.SystemStatsResponse
// @Router      /admin/stats [get]
func (h *AdminHandler) GetSystemStats(c *gin.Context) {
	result, err := h.usecase.GetSystemStats(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// ListUsers godoc
// @Summary     List all users
// @Tags        Admin
// @Produce     json
// @Param       page query int false "Page number" default(1)
// @Param       pageSize query int false "Page size" default(20)
// @Success     200 {object} dto.UserListResponse
// @Router      /admin/users [get]
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	result, err := h.usecase.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// UpdateUserRole godoc
// @Summary     Update user role
// @Tags        Admin
// @Accept      json
// @Produce     json
// @Param       id path int true "User ID"
// @Param       request body dto.UpdateUserRoleRequest true "New role"
// @Success     200 {object} response.Response
// @Router      /admin/users/{id}/role [put]
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req dto.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err = h.usecase.UpdateUserRole(c.Request.Context(), id, req.Role)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "User role updated"})
}

// UpdateUser godoc
// @Summary     Update user details
// @Tags        Admin
// @Accept      json
// @Produce     json
// @Param       id path int true "User ID"
// @Param       request body dto.UpdateUserRequest true "Update details"
// @Success     200 {object} response.Response
// @Router      /admin/users/{id} [put]
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err = h.usecase.UpdateUser(c.Request.Context(), id, &req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "User updated successfully"})
}

// ToggleUserActive godoc
// @Summary     Toggle user active status
// @Tags        Admin
// @Accept      json
// @Produce     json
// @Param       id path int true "User ID"
// @Param       request body dto.ToggleUserActiveRequest true "Active status"
// @Success     200 {object} response.Response
// @Router      /admin/users/{id}/active [put]
func (h *AdminHandler) ToggleUserActive(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req dto.ToggleUserActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err = h.usecase.ToggleUserActive(c.Request.Context(), id, req.IsActive)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "User status updated"})
}

// ListRoles godoc
// @Summary     List all roles in the system
// @Tags        Admin
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} dto.RoleResponse
// @Router      /admin/roles [get]
func (h *AdminHandler) ListRoles(c *gin.Context) {
	roles, err := h.usecase.ListRoles(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, roles)
}

// =============================================
// PERMISSION MANAGEMENT HANDLERS
// =============================================

// GrantRoleToUser godoc
// @Summary     Assign role to user
// @Tags        Admin
// @Accept      json
// @Produce     json
// @Param       userId path int true "User ID"
// @Param       request body dto.GrantRoleRequest true "Role to grant"
// @Success     200 {object} response.Response
// @Router      /admin/users/{userId}/roles [post]
func (h *AdminHandler) GrantRoleToUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req dto.GrantRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	performedBy := c.GetInt64("user_id")
	err = h.usecase.GrantRoleToUser(c.Request.Context(), userID, req.RoleID, performedBy)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "Role granted successfully"})
}

// RevokeRoleFromUser godoc
// @Summary     Remove role from user
// @Tags        Admin
// @Accept      json
// @Produce     json
// @Param       userId path int true "User ID"
// @Param       request body dto.RevokeRoleRequest true "Role to revoke"
// @Success     200 {object} response.Response
// @Router      /admin/users/{userId}/roles [delete]
func (h *AdminHandler) RevokeRoleFromUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req dto.RevokeRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err = h.usecase.RevokeRoleFromUser(c.Request.Context(), userID, req.RoleID)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "Role revoked successfully"})
}

// GetUserRoles godoc
// @Summary     Get all roles for a user
// @Tags        Admin
// @Produce     json
// @Param       userId path int true "User ID"
// @Success     200 {object} dto.UserRoleResponse
// @Router      /admin/users/{userId}/roles [get]
func (h *AdminHandler) GetUserRoles(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	result, err := h.usecase.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// ListPermissions godoc
// @Summary     List all permissions in the system
// @Tags        Admin
// @Produce     json
// @Success     200 {object} dto.ListPermissionsResponse
// @Router      /admin/permissions [get]
func (h *AdminHandler) ListPermissions(c *gin.Context) {
	result, err := h.usecase.ListPermissions(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// GetRolePermissions godoc
// @Summary     Get all permissions for a role
// @Tags        Admin
// @Produce     json
// @Param       roleId path int true "Role ID"
// @Success     200 {object} dto.RolePermissionsResponse
// @Router      /admin/roles/{roleId}/permissions [get]
func (h *AdminHandler) GetRolePermissions(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("roleId"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	result, err := h.usecase.GetRolePermissions(c.Request.Context(), int32(roleID))
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// GrantPermissionToRole godoc
// @Summary     Assign permission to role
// @Tags        Admin
// @Accept      json
// @Produce     json
// @Param       roleId path int true "Role ID"
// @Param       request body dto.GrantPermissionRequest true "Permission to grant"
// @Success     200 {object} response.Response
// @Router      /admin/roles/{roleId}/permissions [post]
func (h *AdminHandler) GrantPermissionToRole(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("roleId"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	var req dto.GrantPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	performedBy := c.GetInt64("user_id")
	err = h.usecase.GrantPermissionToRole(c.Request.Context(), int32(roleID), req.PermissionID, performedBy)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "Permission granted successfully"})
}

// RevokePermissionFromRole godoc
// @Summary     Remove permission from role
// @Tags        Admin
// @Produce     json
// @Param       roleId path int true "Role ID"
// @Param       permissionId path int true "Permission ID"
// @Success     200 {object} response.Response
// @Router      /admin/roles/{roleId}/permissions/{permissionId} [delete]
func (h *AdminHandler) RevokePermissionFromRole(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("roleId"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	permissionID, err := strconv.ParseInt(c.Param("permissionId"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid permission ID")
		return
	}

	err = h.usecase.RevokePermissionFromRole(c.Request.Context(), int32(roleID), int32(permissionID))
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "Permission revoked successfully"})
}

// GetAuditLog godoc
// @Summary     Get permission audit log
// @Tags        Admin
// @Produce     json
// @Param       page query int false "Page number" default(1)
// @Param       pageSize query int false "Page size" default(20)
// @Success     200 {object} dto.AuditLogResponse
// @Router      /admin/audit-log [get]
func (h *AdminHandler) GetAuditLog(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	result, err := h.usecase.GetAuditLog(c.Request.Context(), page, pageSize)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}
