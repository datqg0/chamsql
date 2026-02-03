package http

import (
	"backend/db"
	"backend/internals/topic/repository"
	"backend/internals/topic/usecase"
	"backend/pkgs/middlewares"

	"github.com/gin-gonic/gin"
)

func Routes(rg *gin.RouterGroup, database *db.Database, authMiddleware gin.HandlerFunc) {
	repo := repository.NewTopicRepository(database)
	uc := usecase.NewTopicUseCase(repo)
	handler := NewTopicHandler(uc)

	topics := rg.Group("/topics")
	{
		// Public routes
		topics.GET("", handler.List)
		topics.GET("/:slug", handler.GetBySlug)

		// Protected routes (lecturer and admin only)
		protected := topics.Group("")
		protected.Use(authMiddleware)
		protected.Use(middlewares.RoleMiddleware("lecturer", "admin"))
		{
			protected.POST("", handler.Create)
			protected.PUT("/:id", handler.Update)
			protected.DELETE("/:id", handler.Delete)
		}
	}
}
