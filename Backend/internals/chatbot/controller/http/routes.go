package http

import (
	"backend/configs"
	"backend/internals/chatbot/usecase"

	"github.com/gin-gonic/gin"
)

// Routes registers chatbot routes
// Available to authenticated users (students, lecturers, admins)
func Routes(rg *gin.RouterGroup, cfg *configs.Config, authMiddleware gin.HandlerFunc) {
	uc := usecase.NewChatbotUseCase(cfg)
	handler := NewChatbotHandler(uc)

	chatbot := rg.Group("/chatbot")
	chatbot.Use(authMiddleware)
	{
		chatbot.POST("/ask", handler.Ask)
	}
}
