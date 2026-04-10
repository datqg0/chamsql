package http

import (
	"backend/pkgs/middlewares"

	"github.com/gin-gonic/gin"
)

// Routes registers all PDF routes
func Routes(rg *gin.RouterGroup, handler *PDFHandler, authMiddleware gin.HandlerFunc) {
	pdfGroup := rg.Group("/pdf")
	pdfGroup.Use(authMiddleware)
	pdfGroup.Use(middlewares.RoleMiddleware("lecturer", "admin"))
	{
		pdfGroup.POST("/upload", handler.Upload)
		pdfGroup.GET("/:id/status", handler.GetStatus)
		pdfGroup.GET("/:id/problems", handler.GetProblems)
	}
}
