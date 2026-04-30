package http

import (
	"backend/internals/chatbot/controller/dto"
	"backend/internals/chatbot/usecase"
	"backend/pkgs/response"

	"github.com/gin-gonic/gin"
)

// ChatbotHandler handles chatbot HTTP requests
type ChatbotHandler struct {
	usecase usecase.IChatbotUseCase
}

// NewChatbotHandler creates a new chatbot handler
func NewChatbotHandler(uc usecase.IChatbotUseCase) *ChatbotHandler {
	return &ChatbotHandler{usecase: uc}
}

// Ask godoc
// @Summary     Ask the SQL guidance chatbot
// @Description Student sends a question about SQL, optionally with problem context and their SQL code
// @Tags        Chatbot
// @Accept      json
// @Produce     json
// @Param       request body dto.ChatRequest true "Chat message with optional context"
// @Success     200 {object} dto.ChatResponse
// @Router      /chatbot/ask [post]
func (h *ChatbotHandler) Ask(c *gin.Context) {
	var req dto.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.usecase.Ask(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}
