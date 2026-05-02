package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"backend/internals/pdf/controller/dto"
	"backend/internals/pdf/usecase"
	miniopkg "backend/pkgs/minio"
	"backend/pkgs/middlewares"
	"backend/pkgs/response"

	"github.com/gin-gonic/gin"
)

// PDFHandler handles PDF upload operations
type PDFHandler struct {
	uploadManager usecase.IUploadManager
	storage       miniopkg.IUploadService
	appCtx        context.Context
}

// NewPDFHandler creates a new PDF handler
func NewPDFHandler(uploadManager usecase.IUploadManager, storage miniopkg.IUploadService, appCtx context.Context) *PDFHandler {
	return &PDFHandler{
		uploadManager: uploadManager,
		storage:       storage,
		appCtx:        appCtx,
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

	// Save uploaded file to OS temp directory for extraction
	tmpDir := filepath.Join(os.TempDir(), "chamsql", "pdf_uploads")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		response.InternalServerError(c, "Failed to prepare upload directory")
		return
	}
	storedFileName := strconv.FormatInt(time.Now().UnixNano(), 10) + "_" + filepath.Base(file.Filename)
	filePath := filepath.Join(tmpDir, storedFileName)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		response.InternalServerError(c, "Failed to save uploaded file")
		return
	}

	// Upload lên MinIO — lưu vĩnh viễn (bucket: pdf-files)
	minioURL, err := h.storage.UploadFileFromPath(c.Request.Context(), filePath, "pdf-files", "application/pdf")
	if err != nil {
		_ = os.Remove(filePath)
		response.InternalServerError(c, "Failed to upload to storage: "+err.Error())
		return
	}

	// Create upload record với MinIO URL
	upload, err := h.uploadManager.HandleUpload(c.Request.Context(), lecturerID, minioURL, file.Filename, file.Filename)
	if err != nil {
		_ = os.Remove(filePath)
		response.InternalServerError(c, err.Error())
		return
	}

	// Return 202 Accepted ngay lập tức
	c.JSON(http.StatusAccepted, dto.PDFUploadResponse{
		ID:        upload.ID,
		Status:    upload.Status,
		FileName:  upload.FileName,
		CreatedAt: upload.CreatedAt,
		Message:   "PDF upload accepted. Processing in background.",
	})

	// Chạy extraction + AI generation trong goroutine riêng.
	go func(uploadID int64, localPath string) {
		bgCtx := h.appCtx

		if err := h.uploadManager.ProcessExtraction(bgCtx, uploadID); err != nil {
			fmt.Printf("PDF extraction failed for upload %d: %v\n", uploadID, err)
			// Cập nhật status = "failed" trong DB để frontend biết
			if updateErr := h.uploadManager.MarkUploadFailed(bgCtx, uploadID, err.Error()); updateErr != nil {
				fmt.Printf("Failed to mark upload as failed: %v\n", updateErr)
			}
			_ = os.Remove(localPath)
			return
		}

		if err := h.uploadManager.GenerateAIContent(bgCtx, uploadID); err != nil {
			fmt.Printf("AI content generation failed for upload %d: %v\n", uploadID, err)
			// AI thất bại không critical — status vẫn là "extracted" để lecturer review thủ công
		}

		// Xóa local temp sau khi extract xong
		_ = os.Remove(localPath)
	}(upload.ID, filePath)
}

// DownloadPDF godoc
// @Summary     Get presigned download URL for original PDF
// @Tags        PDF Upload
// @Produce     json
// @Param       id path int true "Upload ID"
// @Success     200 {object} map[string]interface{}
// @Router      /lecturer/pdf/{id}/download [get]
func (h *PDFHandler) DownloadPDF(c *gin.Context) {
	uploadID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid upload ID")
		return
	}

	upload, err := h.uploadManager.GetUploadStatus(c.Request.Context(), uploadID)
	if err != nil {
		response.NotFound(c, "Upload record not found")
		return
	}

	// Gen presigned URL (24h)
	presignedURL, err := h.storage.GetPresignedURL(c.Request.Context(), upload.FilePath, 24*time.Hour)
	if err != nil {
		response.InternalServerError(c, "Failed to generate download URL: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"download_url": presignedURL,
		"file_name":    upload.FileName,
		"expires_in":   "24h",
	})
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
			Title               string `json:"title"`
			Description         string `json:"description"`
			Difficulty          string `json:"difficulty"`
			SolutionQuery       string `json:"solution_query"`
			InitScript          string `json:"init_script"`
			AIValidationWarning string `json:"ai_validation_warning"`
			AIValidationPassed  bool   `json:"ai_validation_passed"`
		}
		if err := json.Unmarshal(p.ProblemDraft, &draft); err == nil {
			problemDTOs[i].Title = draft.Title
			problemDTOs[i].Description = draft.Description
			problemDTOs[i].Difficulty = draft.Difficulty
			problemDTOs[i].SolutionQuery = draft.SolutionQuery
			problemDTOs[i].InitScript = draft.InitScript
			problemDTOs[i].AIValidationWarning = draft.AIValidationWarning
			problemDTOs[i].AIValidationPassed = draft.AIValidationPassed
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
		lecturerID, _ := middlewares.GetUserID(c)
		if err := h.uploadManager.ConfirmProblem(c.Request.Context(), queueID, lecturerID, req.SolutionQuery, req.DBType); err != nil {
			// Log nhưng không fail — solution đã được update
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

// ConfirmProblem godoc
// @Summary     Confirm extracted problem and save to main problems table
// @Description Giảng viên xác nhận bài toán đã extract từ PDF và lưu vào DB chính thức.
// @Description Có thể cung cấp hoặc override solution_query tại bước này.
// @Tags        PDF Upload
// @Accept      json
// @Produce     json
// @Param       id path int true "Problem Queue ID"
// @Param       request body dto.ConfirmProblemRequest true "Confirm request"
// @Success     200 {object} dto.ConfirmProblemResponse
// @Router      /lecturer/pdf/problems/{id}/confirm [post]
func (h *PDFHandler) ConfirmProblem(c *gin.Context) {
	lecturerID, _ := middlewares.GetUserID(c)

	queueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid problem queue ID")
		return
	}

	var req dto.ConfirmProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.uploadManager.ConfirmProblem(c.Request.Context(), queueID, lecturerID, req.SolutionQuery, req.DBType); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, dto.ConfirmProblemResponse{
		Message: "Problem confirmed and saved to problems table successfully",
	})
}
