package middlewares

import (
	"fmt"
	"strconv"

	"backend/pkgs/logger"
	"backend/pkgs/permissions"

	"github.com/gin-gonic/gin"
)

// PermissionMiddleware checks if user has permission for resource+action
// Usage: router.POST("/exams", authMiddleware, PermissionMiddleware("exam", "add"), handler)
func PermissionMiddleware(permService permissions.PermissionService, resourceType, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")
		if userID == 0 {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// Check if user has permission
		hasPermission, err := permService.HasPermission(c.Request.Context(), userID, resourceType, action)
		if err != nil {
			logger.Error("Permission check failed: %v", err)
			c.JSON(500, gin.H{"error": "internal server error"})
			c.Abort()
			return
		}

		if !hasPermission {
			logger.Warn("User %d denied permission %s:%s", userID, resourceType, action)
			c.JSON(403, gin.H{
				"error":   "forbidden",
				"message": fmt.Sprintf("You don't have permission to %s %s", action, resourceType),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// parseIDParam extracts int64 from :id parameter
func parseIDParam(c *gin.Context) int64 {
	idStr := c.Param("id")
	if idStr == "" {
		return 0
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

// ResourcePermissionMiddleware checks if user has permission AND access to a specific resource
// Usage: router.PUT("/exams/:id", authMiddleware, ResourcePermissionMiddleware("exam", "update"), handler)
// Expects :id param in URL
func ResourcePermissionMiddleware(permService permissions.PermissionService, resourceType, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")
		if userID == 0 {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		resourceID := parseIDParam(c)
		if resourceID == 0 {
			c.JSON(400, gin.H{"error": "missing resource id"})
			c.Abort()
			return
		}

		// Check combined: permission + resource access
		canAccess, err := permService.CanAccess(c.Request.Context(), userID, resourceType, action, resourceID)
		if err != nil {
			logger.Error("Permission check failed: %v", err)
			c.JSON(500, gin.H{"error": "internal server error"})
			c.Abort()
			return
		}

		if !canAccess {
			logger.Warn("User %d denied access to %s:%d action:%s", userID, resourceType, resourceID, action)
			c.JSON(403, gin.H{
				"error":   "forbidden",
				"message": fmt.Sprintf("You don't have permission to %s this resource", action),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminMiddleware checks if user is admin
// Usage: router.GET("/admin/users", authMiddleware, AdminMiddleware(permService), handler)
func AdminMiddleware(permService permissions.PermissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")
		if userID == 0 {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// Check if user has any admin permission
		hasAdmin, err := permService.HasPermission(c.Request.Context(), userID, "role", "view")
		if err != nil || !hasAdmin {
			logger.Warn("User %d attempted admin access", userID)
			c.JSON(403, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OwnerOrAdminMiddleware checks if user owns the resource or is admin
// Usage: router.DELETE("/exams/:id", authMiddleware, OwnerOrAdminMiddleware(permService, "exam"), handler)
func OwnerOrAdminMiddleware(permService permissions.PermissionService, resourceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")
		if userID == 0 {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		resourceID := parseIDParam(c)
		if resourceID == 0 {
			c.JSON(400, gin.H{"error": "missing resource id"})
			c.Abort()
			return
		}

		// Check if user is owner
		isOwner, err := permService.IsResourceOwner(c.Request.Context(), userID, resourceType, resourceID)
		if err != nil {
			logger.Error("Owner check failed: %v", err)
			c.JSON(500, gin.H{"error": "internal server error"})
			c.Abort()
			return
		}

		// If not owner, check if admin
		if !isOwner {
			isAdmin, err := permService.HasPermission(c.Request.Context(), userID, resourceType, "delete")
			if err != nil || !isAdmin {
				logger.Warn("User %d not owner of %s:%d", userID, resourceType, resourceID)
				c.JSON(403, gin.H{"error": "forbidden"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
