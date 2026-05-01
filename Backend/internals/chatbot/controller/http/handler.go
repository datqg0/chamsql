package http

import (
    "backend/internals/chatbot/controller/dto"
    "backend/internals/chatbot/usecase"
    "backend/pkgs/middlewares"
    "backend/pkgs/response"

    "github.com/gin-gonic/gin"
    "strings"
)

type ChatbotHandler struct {
    usecase usecase.IChatbotUseCase
}

func NewChatbotHandler(uc usecase.IChatbotUseCase) *ChatbotHandler {
    return &ChatbotHandler{usecase: uc}
}

// Ask godoc
// @Summary     Ask chatbot
// @Tags        Chatbot
// @Accept      json
// @Produce     json
// @Param       request body dto.ChatRequest true "Chat data"
// @Success     200 {object} dto.ChatResponse
// @Router      /chatbot/ask [post]
func (h *ChatbotHandler) Ask(c *gin.Context) {
    var req dto.ChatRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())
        return
    }

    // Gắn UserID từ JWT nếu có
    if userID, ok := middlewares.GetUserID(c); ok {
        req.UserID = &userID
    }

    result, err := h.usecase.Ask(c.Request.Context(), &req)
    if err != nil {
        // Rate limit error → 429
        if strings.Contains(err.Error(), "hết lượt") {
            c.JSON(429, gin.H{"error": err.Error()})
            return
        }
        response.InternalServerError(c, err.Error())
        return
    }
    response.Success(c, result)
}
