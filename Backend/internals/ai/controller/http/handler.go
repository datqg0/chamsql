package http

import (
    "github.com/gin-gonic/gin"

    aiUsecase "backend/internals/ai/usecase"
    "backend/pkgs/response"
)

type AIHandler struct {
    orchestrator aiUsecase.IAIOrchestrator
}

func NewAIHandler(orchestrator aiUsecase.IAIOrchestrator) *AIHandler {
    return &AIHandler{orchestrator: orchestrator}
}

// GenerateProblem godoc
// @Summary     Sinh đề SQL từ mô tả + schema
// @Tags        AI
// @Accept      json
// @Produce     json
// @Param       body body map[string]string true "Thông tin đề"
// @Success     200 {object} aiUsecase.CompleteProblem
// @Router      /ai/generate-problem [post]
func (h *AIHandler) GenerateProblem(c *gin.Context) {
    var req struct {
        Description string `json:"description" binding:"required"`
        SchemaSQL   string `json:"schema_sql" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "Invalid request: "+err.Error())
        return
    }

    result, err := h.orchestrator.GenerateCompleteProblem(c.Request.Context(), req.Description, req.SchemaSQL)
    if err != nil {
        response.InternalServerError(c, "AI generation failed: "+err.Error())
        return
    }
    response.Success(c, result)
}

// ValidateSolution godoc
// @Summary     Validate SQL solution với schema
// @Tags        AI
// @Produce     json
// @Param       schema_sql query string true "Schema SQL"
// @Param       solution_sql query string true "Solution SQL"
// @Success     200 {object} map[string]interface{}
// @Router      /ai/validate-solution [post]
func (h *AIHandler) ValidateSolution(c *gin.Context) {
    var body struct {
        SchemaSQL   string `json:"schema_sql" binding:"required"`
        SolutionSQL string `json:"solution_sql" binding:"required"`
    }
    if err := c.ShouldBindJSON(&body); err != nil {
        response.BadRequest(c, err.Error())
        return
    }

    result, err := h.orchestrator.ValidateTestCases(c.Request.Context(), body.SchemaSQL, body.SolutionSQL)
    if err != nil {
        response.InternalServerError(c, err.Error())
        return
    }
    response.Success(c, result)
}
