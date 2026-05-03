package http

import (
	"backend/configs"
	"backend/db"
	"backend/internals/admin/usecase"
	"backend/pkgs/middlewares"
	"backend/pkgs/redis"
	"backend/pkgs/runner"

	"github.com/gin-gonic/gin"
)

// Routes - Register all admin endpoints
// Requires authentication and admin role middleware
func Routes(rg *gin.RouterGroup, database *db.Database, cache redis.IRedis, authMiddleware gin.HandlerFunc, cfg *configs.Config, r runner.Runner) {
	uc := usecase.NewAdminUseCase(database, cache)
	handler := NewAdminHandler(uc)
	sandboxHandler := NewSandboxHandler(cfg, r)

	admin := rg.Group("/admin")
	admin.Use(authMiddleware)
	admin.Use(middlewares.RoleMiddleware("admin"))
	{
		// Sandbox management (accessible to admin and lecturer)
		RegisterSandboxRoutes(admin, sandboxHandler)
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
		admin.POST("/users/batch-import", handler.ImportUsers)
		admin.PUT("/users/:id", handler.UpdateUser)
		admin.DELETE("/users/:id", handler.DeleteUser)
		admin.PUT("/users/:id/role", handler.UpdateUserRole)
		admin.PUT("/users/:id/active", handler.ToggleUserActive)

		// =============================================
		// USER ROLE ASSIGNMENT ENDPOINTS
		// Assign/revoke roles to/from users
		// =============================================
		admin.POST("/users/:id/roles", handler.GrantRoleToUser)
		admin.DELETE("/users/:id/roles", handler.RevokeRoleFromUser)
		admin.GET("/users/:id/roles", handler.GetUserRoles)

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

		// =============================================
		// DASHBOARD & ANALYTICS ENDPOINTS
		// For research paper data and system monitoring
		// =============================================
		admin.GET("/dashboard", handler.GetDashboard)
		admin.GET("/timeline", handler.GetPerformanceTimeline)
	}
}
