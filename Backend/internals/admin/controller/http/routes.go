package http

import (
	"backend/db"
	"backend/internals/admin/usecase"
	"backend/pkgs/middlewares"

	"github.com/gin-gonic/gin"
)

// Routes - Register all admin endpoints
// Requires authentication and admin role middleware
func Routes(rg *gin.RouterGroup, database *db.Database, authMiddleware gin.HandlerFunc) {
	uc := usecase.NewAdminUseCase(database)
	handler := NewAdminHandler(uc)

	admin := rg.Group("/admin")
	admin.Use(authMiddleware)
	admin.Use(middlewares.RoleMiddleware("admin"))
	{
		// System stats
		admin.GET("/stats", handler.GetSystemStats)

		// =============================================
		// ROLE MANAGEMENT ENDPOINTS
		// =============================================
		admin.GET("/roles", handler.ListRoles)

		// =============================================
		// USER MANAGEMENT ENDPOINTS
		// =============================================
		admin.GET("/users", handler.ListUsers)
		admin.POST("/users/import", handler.ImportUsers)
		admin.PUT("/users/:id", handler.UpdateUser)
		admin.PUT("/users/:id/role", handler.UpdateUserRole)
		admin.PUT("/users/:id/active", handler.ToggleUserActive)

		// =============================================
		// USER ROLE ASSIGNMENT ENDPOINTS
		// Assign/revoke roles to/from users
		// =============================================
		admin.POST("/users/:userId/roles", handler.GrantRoleToUser)
		admin.DELETE("/users/:userId/roles", handler.RevokeRoleFromUser)
		admin.GET("/users/:userId/roles", handler.GetUserRoles)

		// =============================================
		// PERMISSION MANAGEMENT ENDPOINTS
		// List permissions and manage role-permission assignments
		// =============================================
		admin.GET("/permissions", handler.ListPermissions)
		admin.GET("/roles/:roleId/permissions", handler.GetRolePermissions)
		admin.POST("/roles/:roleId/permissions", handler.GrantPermissionToRole)
		admin.DELETE("/roles/:roleId/permissions/:permissionId", handler.RevokePermissionFromRole)

		// =============================================
		// AUDIT LOG ENDPOINT
		// View all permission/role changes
		// =============================================
		admin.GET("/audit-log", handler.GetAuditLog)
	}
}
