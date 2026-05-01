package http

import (
    "github.com/gin-gonic/gin"
    "backend/pkgs/middlewares"
    "backend/pkgs/jwt"
    "backend/pkgs/redis"
)

func RegisterRoutes(r *gin.RouterGroup, h *AIHandler, jwtProvider jwt.JWTProvider, cache redis.IRedis) {
    ai := r.Group("/ai")
    ai.Use(middlewares.AuthMiddleware(jwtProvider, cache))
    // Chỉ lecturer và admin mới được generate đề
    ai.POST("/generate-problem", middlewares.RoleMiddleware("lecturer", "admin"), h.GenerateProblem)
    ai.POST("/validate-solution", middlewares.RoleMiddleware("lecturer", "admin"), h.ValidateSolution)
}
