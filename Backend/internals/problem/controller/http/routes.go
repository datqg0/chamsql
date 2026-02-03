package http

import (
	"backend/db"
	"backend/internals/problem/repository"
	"backend/internals/problem/usecase"
	"backend/pkgs/middlewares"

	"github.com/gin-gonic/gin"
)

func Routes(rg *gin.RouterGroup, database *db.Database, authMiddleware gin.HandlerFunc) {
	repo := repository.NewProblemRepository(database)
	uc := usecase.NewProblemUseCase(repo)
	handler := NewProblemHandler(uc)

	problems := rg.Group("/problems")
	{
		// Public routes (can be accessed without auth, but auth adds user progress)
		problems.GET("", handler.List)
		problems.GET("/:slug", handler.GetBySlug)

		// Protected routes (lecturer and admin only)
		protected := problems.Group("")
		protected.Use(authMiddleware)
		protected.Use(middlewares.RoleMiddleware("lecturer", "admin"))
		{
			protected.POST("", handler.Create)
			protected.PUT("/:id", handler.Update)
			protected.DELETE("/:id", handler.Delete)
		}
	}
}
