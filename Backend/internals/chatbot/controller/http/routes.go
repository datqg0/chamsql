package http

import (
	"github.com/gin-gonic/gin"
)

// Routes registers chatbot routes
// Available to authenticated users (students, lecturers, admins)
func Routes(rg *gin.RouterGroup, handler *ChatbotHandler, authMiddleware gin.HandlerFunc) {
	chatbot := rg.Group("/chatbot")
	chatbot.Use(authMiddleware)
	{
		chatbot.POST("/ask", handler.Ask)
	}
}
