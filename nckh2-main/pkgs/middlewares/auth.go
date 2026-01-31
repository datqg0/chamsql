package middlewares

import (
	"backend/pkgs/jwt"
	"backend/pkgs/redis"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtProv jwt.JWTProvider, cache redis.IRedis) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			return
		}

		token := parts[1]
		tokenString := parts[1]

		// 1. Validate Token Signature
		claims, err := jwtProv.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// 2. Check Blacklist (Redis)
		if cache != nil && cache.IsConnected() {
			blacklistKey := fmt.Sprintf("blacklist:%s", tokenString)
			var val string
			_ = cache.Get(blacklistKey, &val)
			if val == "revoked" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token revoked"})
				return
			}
		}

		// 3. Set User Claims to Context
		if userID, ok := (*claims)["user_id"].(float64); ok {
			c.Set("userID", int64(userID))
		}
		if role, ok := (*claims)["role"].(string); ok {
			c.Set("role", role)
		}

		_ = tokenString // unused suppressed

		c.Set("token", token)
		c.Next()
	}
}

// RoleMiddleware checks if the user has one of the allowed roles
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Role not found in token"})
			return
		}

		userRole, ok := role.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid role format"})
			return
		}

		// Check if user's role is in the allowed list
		for _, allowed := range allowedRoles {
			if userRole == allowed {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": fmt.Sprintf("Access denied. Required roles: %v", allowedRoles),
		})
	}
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	id, ok := userID.(int64)
	return id, ok
}

// GetUserRole extracts user role from context
func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get("role")
	if !exists {
		return "", false
	}
	r, ok := role.(string)
	return r, ok
}
