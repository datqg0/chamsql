package http

import (
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

	// TODO: Implement getting problems from review queue
	c.JSON(http.StatusOK, gin.H{
		"message":   "Get problems endpoint",
		"upload_id": uploadID,
	})
}
