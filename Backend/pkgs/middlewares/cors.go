package middlewares

import (
	"strings"

	"backend/configs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CorsMiddleware(cfg *configs.Config) gin.HandlerFunc {
	allowedOrigins := []string{
		"http://localhost:3000",
		"http://localhost:5173",
		"http://localhost:4200",
	}

	if cfg != nil && cfg.AllowedOrigins != "" {
		for _, origin := range strings.Split(cfg.AllowedOrigins, ",") {
			o := strings.TrimSpace(origin)
			if o != "" {
				allowedOrigins = append(allowedOrigins, o)
			}
		}
	}

	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           86400,
	})
}
