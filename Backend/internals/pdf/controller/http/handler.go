package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internals/pdf/controller/dto"
	"backend/internals/pdf/usecase"
	"backend/pkgs/middlewares"
	"backend/pkgs/response"

	"github.com/gin-gonic/gin"
)

// PDFHandler handles PDF upload operations
type PDFHandler struct {
	uploadManager usecase.IUploadManager
}

// NewPDFHandler creates a new PDF handler
func NewPDFHandler(uploadManager usecase.IUploadManager) *PDFHandler {
	return &PDFHandler{
		uploadManager: uploadManager,
	}
}

// Upload godoc
// @Summary     Upload PDF for problem generation
// @Tags        PDF Upload
// @Accept      multipart/form-data
// @Produce     json
// @Param       file formData file true "PDF file"
// @Success     202 {object} dto.PDFUploadResponse "Accepted - processing asynchronously"
// @Router      /lecturer/pdf/upload [post]
func (h *PDFHandler) Upload(c *gin.Context) {
	lecturerID, _ := middlewares.GetUserID(c)

	// Get file from multipart form
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "No file provided")
		return
	}

	// Validate file is PDF
	if file.Header.Get("Content-Type") != "application/pdf" {
		response.BadRequest(c, "File must be PDF format")
		return
	}

	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		response.BadRequest(c, "File size exceeds 10MB limit")
		return
	}

	// Save file (placeholder - in production use MinIO)
	// For now, assume file is saved to temp location
	filePath := "/tmp/" + file.Filename

	// Open file
	src, err := file.Open()
	if err != nil {
		response.InternalServerError(c, "Failed to open file")
		return
	}
	defer src.Close()

	// TODO: Save to MinIO or temp storage
	// For now just create the upload record

	// Create upload record
	upload, err := h.uploadManager.HandleUpload(c.Request.Context(), lecturerID, filePath, file.Filename, file.Filename)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	// Return 202 Accepted (processing asynchronously)
	c.JSON(http.StatusAccepted, dto.PDFUploadResponse{
		ID:        upload.ID,
		Status:    upload.Status,
		FileName:  upload.FileName,
		CreatedAt: upload.CreatedAt,
		Message:   "PDF upload accepted. Processing in background.",
	})

	// TODO: Queue extraction job to background worker (or call synchronously for now)
	// For MVP, call extraction synchronously
	_ = h.uploadManager.ProcessExtraction(c.Request.Context(), upload.ID)
	_ = h.uploadManager.GenerateAIContent(c.Request.Context(), upload.ID)
}

// GetStatus godoc
// @Summary     Get PDF upload status
// @Tags        PDF Upload
// @Produce     json
// @Param       id path int true "Upload ID"
// @Success     200 {object} dto.PDFUploadStatusResponse
// @Router      /lecturer/pdf/{id}/status [get]
func (h *PDFHandler) GetStatus(c *gin.Context) {
	uploadID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid upload ID")
		return
	}

	upload, err := h.uploadManager.GetUploadStatus(c.Request.Context(), uploadID)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, dto.PDFUploadStatusResponse{
		ID:               upload.ID,
		Status:           upload.Status,
		FileName:         upload.FileName,
		ExtractionResult: upload.ExtractionResult,
		ErrorMessage:     upload.ErrorMessage.String,
		CreatedAt:        upload.CreatedAt,
		UpdatedAt:        upload.UpdatedAt,
	})
}

// GetProblems godoc
// @Summary     Get extracted problems from PDF
// @Description Returns all problems extracted from the uploaded PDF.
// @Description The problems only contain descriptions (like Codeforces).
// @Description Instructors must manually add solution queries separately.
// @Tags        PDF Upload
// @Produce     json
// @Param       id path int true "Upload ID"
// @Success     200 {object} dto.ProblemsResponse
// @Router      /lecturer/pdf/{id}/problems [get]
func (h *PDFHandler) GetProblems(c *gin.Context) {
	uploadID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid upload ID")
		return
	}

	problems, err := h.uploadManager.GetExtractedProblems(c.Request.Context(), uploadID)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	// Convert domain problems to DTOs
	problemDTOs := make([]dto.ProblemReviewResponse, len(problems))
	for i, p := range problems {
		problemDTOs[i] = dto.ProblemReviewResponse{
			ID:            p.ID,
			ProblemNumber: p.ProblemNumber,
			Status:        p.Status,
			CreatedAt:     p.CreatedAt,
			UpdatedAt:     p.UpdatedAt,
		}
		// Parse the draft to get title, description, etc.
		var draft struct {
			Title         string `json:"title"`
			Description   string `json:"description"`
			Difficulty    string `json:"difficulty"`
			SolutionQuery string `json:"solution_query"`
			InitScript    string `json:"init_script"`
		}
		if err := json.Unmarshal(p.ProblemDraft, &draft); err == nil {
			problemDTOs[i].Title = draft.Title
			problemDTOs[i].Description = draft.Description
			problemDTOs[i].Difficulty = draft.Difficulty
			problemDTOs[i].SolutionQuery = draft.SolutionQuery
			problemDTOs[i].InitScript = draft.InitScript
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"upload_id": uploadID,
		"problems":  problemDTOs,
		"note":      "PDF contains problem descriptions only. Solution queries must be added manually.",
	})
}

// UpdateSolution godoc
// @Summary     Update problem solution query
// @Description Instructors manually input the solution query for a problem.
// @Description Since PDF only contains problem descriptions (like Codeforces),
// @Description this endpoint allows adding the correct answer/solution.
// @Tags        PDF Upload
// @Accept      json
// @Produce     json
// @Param       id path int true "Problem Queue ID"
// @Param       request body dto.UpdateSolutionRequest true "Solution update request"
// @Success     200 {object} dto.UpdateSolutionResponse
// @Router      /lecturer/pdf/problems/{id}/solution [put]
func (h *PDFHandler) UpdateSolution(c *gin.Context) {
	queueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid problem ID")
		return
	}

	var req dto.UpdateSolutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Update the solution
	if err := h.uploadManager.UpdateProblemSolution(c.Request.Context(), queueID, req.SolutionQuery); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	// Optionally confirm the problem if db_type is provided
	if req.DBType != "" {
		if err := h.uploadManager.ConfirmProblem(c.Request.Context(), queueID, req.SolutionQuery, req.DBType); err != nil {
			// Log but don't fail - solution was updated
			response.Success(c, dto.UpdateSolutionResponse{
				ID:            queueID,
				SolutionQuery: req.SolutionQuery,
				DBType:        req.DBType,
				Status:        "solution_updated",
				Message:       "Solution updated but not confirmed: " + err.Error(),
			})
			return
		}
	}

	response.Success(c, dto.UpdateSolutionResponse{
		ID:            queueID,
		SolutionQuery: req.SolutionQuery,
		DBType:        req.DBType,
		Status:        "confirmed",
		Message:       "Solution updated and problem confirmed successfully",
	})
}
