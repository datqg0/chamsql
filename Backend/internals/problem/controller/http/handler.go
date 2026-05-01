package http

import (
	"strconv"

	"backend/internals/problem/controller/dto"
	"backend/internals/problem/usecase"
	miniopkg "backend/pkgs/minio"
	"backend/pkgs/middlewares"
	"backend/pkgs/response"
	"time"

	"github.com/gin-gonic/gin"
)

type ProblemHandler struct {
	usecase usecase.IProblemUseCase
	storage miniopkg.IUploadService
}

func NewProblemHandler(uc usecase.IProblemUseCase, storage miniopkg.IUploadService) *ProblemHandler {
	return &ProblemHandler{
		usecase: uc,
		storage: storage,
	}
}

// List godoc
// @Summary     List problems
// @Tags        Problems
// @Produce     json
// @Param       topicId query int false "Filter by topic ID"
// @Param       difficulty query string false "Filter by difficulty (easy, medium, hard)"
// @Param       page query int false "Page number" default(1)
// @Param       pageSize query int false "Page size" default(20)
// @Success     200 {object} dto.ProblemListResponse
// @Router      /problems [get]
func (h *ProblemHandler) List(c *gin.Context) {
	var query dto.ProblemListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}

	role := ""
	if r, ok := middlewares.GetUserRole(c); ok {
		role = r
	}

	result, err := h.usecase.List(c.Request.Context(), role, &query)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// GetBySlug godoc
// @Summary     Get problem by slug
// @Tags        Problems
// @Produce     json
// @Param       slug path string true "Problem slug"
// @Success     200 {object} dto.ProblemResponse
// @Router      /problems/{slug} [get]
func (h *ProblemHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")

	// Check if user is authenticated
	var userID *int64
	if id, ok := middlewares.GetUserID(c); ok {
		userID = &id
	}

	result, err := h.usecase.GetBySlug(c.Request.Context(), slug, userID)
	if err != nil {
		if err == usecase.ErrProblemNotFound {
			response.NotFound(c, "Problem not found")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// Create godoc
// @Summary     Create a new problem
// @Tags        Problems
// @Accept      json
// @Produce     json
// @Param       request body dto.CreateProblemRequest true "Problem data"
// @Success     201 {object} dto.ProblemResponse
// @Router      /problems [post]
func (h *ProblemHandler) Create(c *gin.Context) {
	userID, ok := middlewares.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	var req dto.CreateProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.usecase.Create(c.Request.Context(), userID, &req)
	if err != nil {
		if err == usecase.ErrSlugExists {
			response.BadRequest(c, "Problem slug already exists")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	response.Created(c, result)
}

// Update godoc
// @Summary     Update a problem
// @Tags        Problems
// @Accept      json
// @Produce     json
// @Param       id path int true "Problem ID"
// @Param       request body dto.UpdateProblemRequest true "Problem data"
// @Success     200 {object} dto.ProblemResponse
// @Router      /problems/{id} [put]
func (h *ProblemHandler) Update(c *gin.Context) {
	userID, ok := middlewares.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid problem ID")
		return
	}

	var req dto.UpdateProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.usecase.Update(c.Request.Context(), userID, id, &req)
	if err != nil {
		if err == usecase.ErrProblemNotFound {
			response.NotFound(c, "Problem not found")
			return
		}
		if err == usecase.ErrForbidden {
			response.Forbidden(c, "You don't have permission to modify this problem")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// Delete godoc
// @Summary     Delete a problem
// @Tags        Problems
// @Produce     json
// @Param       id path int true "Problem ID"
// @Success     200 {object} response.Response
// @Router      /problems/{id} [delete]
func (h *ProblemHandler) Delete(c *gin.Context) {
	userID, ok := middlewares.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid problem ID")
		return
	}

	err = h.usecase.Delete(c.Request.Context(), userID, id)
	if err != nil {
		if err == usecase.ErrProblemNotFound {
			response.NotFound(c, "Problem not found")
			return
		}
		if err == usecase.ErrForbidden {
			response.Forbidden(c, "You don't have permission to delete this problem")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "Problem deleted successfully"})
}

// ListMyProblems godoc
// @Summary     List my problems (for lecturers)
// @Tags        Problems
// @Produce     json
// @Param       page query int false "Page number" default(1)
// @Param       pageSize query int false "Page size" default(20)
// @Success     200 {object} dto.ProblemListResponse
// @Router      /lecturer/problems/mine [get]
func (h *ProblemHandler) ListMine(c *gin.Context) {
	userID, ok := middlewares.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	result, err := h.usecase.ListMine(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// DownloadProblemPDF godoc
// @Summary     Download PDF source of a problem
// @Tags        Problems
// @Produce     json
// @Param       slug path string true "Problem slug"
// @Success     200 {object} map[string]interface{}
// @Router      /problems/{slug}/pdf [get]
func (h *ProblemHandler) DownloadProblemPDF(c *gin.Context) {
	slug := c.Param("slug")
	// Lấy userID nếu có để track progress/permission nếu cần trong tương lai
	var userID *int64
	if id, ok := middlewares.GetUserID(c); ok {
		userID = &id
	}

	problem, err := h.usecase.GetBySlug(c.Request.Context(), slug, userID)
	if err != nil {
		response.NotFound(c, "Problem not found")
		return
	}

	if problem.SourcePdfUrl == nil || *problem.SourcePdfUrl == "" {
		response.NotFound(c, "No PDF available for this problem")
		return
	}

	// Gen presigned URL (1h cho sinh viên)
	presignedURL, err := h.storage.GetPresignedURL(c.Request.Context(), *problem.SourcePdfUrl, 1*time.Hour)
	if err != nil {
		response.InternalServerError(c, "Failed to generate download URL: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"download_url": presignedURL,
		"expires_in":   "1h",
	})
}
