package http

import (
	"backend/db"
	"backend/internals/admin/usecase"
	"backend/pkgs/middlewares"

	"github.com/gin-gonic/gin"
)

func Routes(rg *gin.RouterGroup, database *db.Database, authMiddleware gin.HandlerFunc) {
	uc := usecase.NewAdminUseCase(database)
	handler := NewAdminHandler(uc)

	admin := rg.Group("/admin")
	admin.Use(authMiddleware)
	admin.Use(middlewares.RoleMiddleware("admin"))
	{
		// System stats
		admin.GET("/stats", handler.GetSystemStats)

		// User management
		admin.GET("/users", handler.ListUsers)
		admin.POST("/users/import", handler.ImportUsers)
		admin.PUT("/users/:id/role", handler.UpdateUserRole)
		admin.PUT("/users/:id/active", handler.ToggleUserActive)
	}
}
