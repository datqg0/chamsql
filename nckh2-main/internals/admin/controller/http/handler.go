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
