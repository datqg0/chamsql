package http

import (
	"strconv"

	"backend/internals/submission/controller/dto"
	"backend/internals/submission/usecase"
	"backend/pkgs/middlewares"
	"backend/pkgs/response"

	"github.com/gin-gonic/gin"
)

type SubmissionHandler struct {
	usecase usecase.ISubmissionUseCase
}

func NewSubmissionHandler(uc usecase.ISubmissionUseCase) *SubmissionHandler {
	return &SubmissionHandler{usecase: uc}
}

// Run godoc
// @Summary     Run SQL query without submitting
// @Tags        Submissions
// @Accept      json
// @Produce     json
// @Param       problemId path int true "Problem ID"
// @Param       request body dto.RunQueryRequest true "Query data"
// @Success     200 {object} dto.RunQueryResponse
// @Router      /problems/{problemId}/run [post]
func (h *SubmissionHandler) Run(c *gin.Context) {
	problemID, err := strconv.ParseInt(c.Param("problemId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid problem ID")
		return
	}

	var req dto.RunQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.usecase.Run(c.Request.Context(), problemID, &req)
	if err != nil {
		if err == usecase.ErrProblemNotFound {
			response.NotFound(c, "Problem not found")
			return
		}
		if err == usecase.ErrUnsupportedDB {
			response.BadRequest(c, "Database type not supported for this problem")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// Submit godoc
// @Summary     Submit SQL solution
// @Tags        Submissions
// @Accept      json
// @Produce     json
// @Param       problemId path int true "Problem ID"
// @Param       request body dto.SubmitQueryRequest true "Query data"
// @Success     200 {object} dto.SubmitQueryResponse
// @Router      /problems/{problemId}/submit [post]
func (h *SubmissionHandler) Submit(c *gin.Context) {
	userID, ok := middlewares.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	problemID, err := strconv.ParseInt(c.Param("problemId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid problem ID")
		return
	}

	var req dto.SubmitQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.usecase.Submit(c.Request.Context(), userID, problemID, &req)
	if err != nil {
		if err == usecase.ErrProblemNotFound {
			response.NotFound(c, "Problem not found")
			return
		}
		if err == usecase.ErrUnsupportedDB {
			response.BadRequest(c, "Database type not supported for this problem")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// GetByID godoc
// @Summary     Get submission by ID
// @Tags        Submissions
// @Produce     json
// @Param       id path int true "Submission ID"
// @Success     200 {object} dto.SubmissionResponse
// @Router      /submissions/{id} [get]
func (h *SubmissionHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid submission ID")
		return
	}

	result, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrSubmissionNotFound {
			response.NotFound(c, "Submission not found")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// List godoc
// @Summary     List my submissions
// @Tags        Submissions
// @Produce     json
// @Param       page query int false "Page number" default(1)
// @Param       pageSize query int false "Page size" default(20)
// @Success     200 {object} dto.SubmissionListResponse
// @Router      /submissions [get]
func (h *SubmissionHandler) List(c *gin.Context) {
	userID, ok := middlewares.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	result, err := h.usecase.ListByUser(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}
