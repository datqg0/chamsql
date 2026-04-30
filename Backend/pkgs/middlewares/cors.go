package middlewares

import (
	"strings"
	"time"

	"backend/configs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CorsMiddleware(cfg *configs.Config) gin.HandlerFunc {
	origins := []string{"http://localhost:3000", "http://localhost:5173", "http://localhost:4200", "http://localhost:9000", "http://localhost:9001", "http://localhost:5672", "http://localhost:15672"}

	if cfg != nil && cfg.AllowedOrigins != "" {
		for _, o := range strings.Split(cfg.AllowedOrigins, ",") {
			origins = append(origins, strings.TrimSpace(o))
		}
	}

	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
