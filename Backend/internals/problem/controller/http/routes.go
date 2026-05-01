package http

import (
	"backend/pkgs/middlewares"

	"github.com/gin-gonic/gin"
)

func Routes(rg *gin.RouterGroup, handler *ProblemHandler, authMiddleware gin.HandlerFunc) {
	problems := rg.Group("/problems")
	{
		// Public routes
		problems.GET("", handler.List)
		problems.GET("/:slug", handler.GetBySlug)
		problems.GET("/:slug/pdf", handler.DownloadProblemPDF)

		// Protected routes (lecturer and admin only)
		protected := problems.Group("")
		protected.Use(authMiddleware)
		protected.Use(middlewares.RoleMiddleware("lecturer", "admin"))
		{
			protected.POST("", handler.Create)
			protected.GET("/mine", handler.ListMine)
			protected.PUT("/:id", handler.Update)
			protected.DELETE("/:id", handler.Delete)
		}
	}
}
