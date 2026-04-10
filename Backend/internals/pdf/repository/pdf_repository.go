package repository

import (
	"context"
	"database/sql"
	"fmt"

	"backend/db"
	"backend/internals/pdf/domain"
	"backend/sql/models"
)

// IPDFRepository defines PDF upload data access operations
type IPDFRepository interface {
	// PDF Upload operations
	CreatePDFUpload(ctx context.Context, lecturerID int64, filePath, fileName, originalFilename string) (*domain.PDFUpload, error)
	GetPDFUploadByID(ctx context.Context, id int64) (*domain.PDFUpload, error)
	GetPDFUploadsByLecturer(ctx context.Context, lecturerID int64, limit, offset int32) ([]domain.PDFUpload, error)
	UpdatePDFUploadStatus(ctx context.Context, id int64, status string) (*domain.PDFUpload, error)
	UpdatePDFUploadWithExtraction(ctx context.Context, id int64, status string, extractionResult []byte) (*domain.PDFUpload, error)
	UpdatePDFUploadError(ctx context.Context, id int64, errorMessage string) (*domain.PDFUpload, error)

	// Problem Review Queue operations
	CreateProblemReviewQueue(ctx context.Context, pdfUploadID int64, problemNumber int, problemDraft []byte) (*domain.ProblemReviewQueue, error)
	GetProblemReviewQueueByID(ctx context.Context, id int64) (*domain.ProblemReviewQueue, error)
	GetProblemReviewQueueByPDF(ctx context.Context, pdfUploadID int64) ([]domain.ProblemReviewQueue, error)
	GetProblemReviewQueueByStatus(ctx context.Context, status string, limit, offset int32) ([]domain.ProblemReviewQueue, error)
	UpdateProblemReviewStatus(ctx context.Context, id int64, status string, reviewerID int64, reviewNotes string) (*domain.ProblemReviewQueue, error)
	UpdateProblemReviewDraft(ctx context.Context, id int64, problemDraft, editsMade []byte) (*domain.ProblemReviewQueue, error)
}

// pdfRepository implements IPDFRepository
type pdfRepository struct {
	db      *db.Database
	queries *models.Queries
}

// NewPDFRepository creates a new PDF repository
func NewPDFRepository(database *db.Database) IPDFRepository {
	return &pdfRepository{
		db:      database,
		queries: models.New(database.GetPool()),
	}
}

// CreatePDFUpload inserts a new PDF upload record
func (r *pdfRepository) CreatePDFUpload(ctx context.Context, lecturerID int64, filePath, fileName, originalFilename string) (*domain.PDFUpload, error) {
	result, err := r.queries.CreatePDFUpload(ctx, models.CreatePDFUploadParams{
		LecturerID:       lecturerID,
		FilePath:         filePath,
		FileName:         fileName,
		OriginalFilename: &originalFilename,
		Status:           "uploading",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF upload: %w", err)
	}

	return pdfUploadModelToDomain(result), nil
}

// GetPDFUploadByID retrieves a PDF upload by ID
func (r *pdfRepository) GetPDFUploadByID(ctx context.Context, id int64) (*domain.PDFUpload, error) {
	result, err := r.queries.GetPDFUploadByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get PDF upload: %w", err)
	}

	return pdfUploadModelToDomain(result), nil
}

// GetPDFUploadsByLecturer retrieves all PDF uploads for a lecturer
func (r *pdfRepository) GetPDFUploadsByLecturer(ctx context.Context, lecturerID int64, limit, offset int32) ([]domain.PDFUpload, error) {
	results, err := r.queries.GetPDFUploadsByLecturer(ctx, models.GetPDFUploadsByLecturerParams{
		LecturerID: lecturerID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get PDF uploads: %w", err)
	}

	uploads := make([]domain.PDFUpload, len(results))
	for i, result := range results {
		uploads[i] = *pdfUploadModelToDomain(result)
	}

	return uploads, nil
}

// UpdatePDFUploadStatus updates the status of a PDF upload
func (r *pdfRepository) UpdatePDFUploadStatus(ctx context.Context, id int64, status string) (*domain.PDFUpload, error) {
	result, err := r.queries.UpdatePDFUploadStatus(ctx, models.UpdatePDFUploadStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update PDF upload status: %w", err)
	}

	return pdfUploadModelToDomain(result), nil
}

// UpdatePDFUploadWithExtraction updates PDF upload with extraction results
func (r *pdfRepository) UpdatePDFUploadWithExtraction(ctx context.Context, id int64, status string, extractionResult []byte) (*domain.PDFUpload, error) {
	result, err := r.queries.UpdatePDFUploadWithExtraction(ctx, models.UpdatePDFUploadWithExtractionParams{
		ID:               id,
		Status:           status,
		ExtractionResult: extractionResult,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update PDF upload with extraction: %w", err)
	}

	return pdfUploadModelToDomain(result), nil
}

// UpdatePDFUploadError updates PDF upload with error status
func (r *pdfRepository) UpdatePDFUploadError(ctx context.Context, id int64, errorMessage string) (*domain.PDFUpload, error) {
	result, err := r.queries.UpdatePDFUploadError(ctx, models.UpdatePDFUploadErrorParams{
		ID:           id,
		ErrorMessage: &errorMessage,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update PDF upload error: %w", err)
	}

	return pdfUploadModelToDomain(result), nil
}

// CreateProblemReviewQueue creates a new problem review queue entry
func (r *pdfRepository) CreateProblemReviewQueue(ctx context.Context, pdfUploadID int64, problemNumber int, problemDraft []byte) (*domain.ProblemReviewQueue, error) {
	pdfIDPtr := &pdfUploadID
	probNumVal := int32(problemNumber)
	probNumPtr := &probNumVal
	result, err := r.queries.CreateProblemReviewQueue(ctx, models.CreateProblemReviewQueueParams{
		PdfUploadID:   pdfIDPtr,
		ProblemNumber: probNumPtr,
		ProblemDraft:  problemDraft,
		Status:        "pending",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create problem review queue: %w", err)
	}

	return reviewQueueModelToDomain(result), nil
}

// GetProblemReviewQueueByID retrieves a problem review queue entry by ID
func (r *pdfRepository) GetProblemReviewQueueByID(ctx context.Context, id int64) (*domain.ProblemReviewQueue, error) {
	result, err := r.queries.GetProblemReviewQueueByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get problem review queue: %w", err)
	}

	return reviewQueueModelToDomain(result), nil
}

// GetProblemReviewQueueByPDF retrieves all review queue entries for a PDF upload
func (r *pdfRepository) GetProblemReviewQueueByPDF(ctx context.Context, pdfUploadID int64) ([]domain.ProblemReviewQueue, error) {
	pdfID := pdfUploadID
	results, err := r.queries.GetProblemReviewQueueByPDF(ctx, &pdfID)
	if err != nil {
		return nil, fmt.Errorf("failed to get problem review queue by PDF: %w", err)
	}

	reviews := make([]domain.ProblemReviewQueue, len(results))
	for i, result := range results {
		reviews[i] = *reviewQueueModelToDomain(result)
	}

	return reviews, nil
}

// GetProblemReviewQueueByStatus retrieves review queue entries by status
func (r *pdfRepository) GetProblemReviewQueueByStatus(ctx context.Context, status string, limit, offset int32) ([]domain.ProblemReviewQueue, error) {
	results, err := r.queries.GetProblemReviewQueueByStatus(ctx, models.GetProblemReviewQueueByStatusParams{
		Status: status,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get problem review queue by status: %w", err)
	}

	reviews := make([]domain.ProblemReviewQueue, len(results))
	for i, result := range results {
		reviews[i] = *reviewQueueModelToDomain(result)
	}

	return reviews, nil
}

// UpdateProblemReviewStatus updates the status of a problem review
func (r *pdfRepository) UpdateProblemReviewStatus(ctx context.Context, id int64, status string, reviewerID int64, reviewNotes string) (*domain.ProblemReviewQueue, error) {
	result, err := r.queries.UpdateProblemReviewStatus(ctx, models.UpdateProblemReviewStatusParams{
		ID:          id,
		Status:      status,
		ReviewerID:  &reviewerID,
		ReviewNotes: &reviewNotes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update problem review status: %w", err)
	}

	return reviewQueueModelToDomain(result), nil
}

// UpdateProblemReviewDraft updates the draft and edits of a problem review
func (r *pdfRepository) UpdateProblemReviewDraft(ctx context.Context, id int64, problemDraft, editsMade []byte) (*domain.ProblemReviewQueue, error) {
	result, err := r.queries.UpdateProblemReviewDraft(ctx, models.UpdateProblemReviewDraftParams{
		ID:           id,
		ProblemDraft: problemDraft,
		EditsMade:    editsMade,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update problem review draft: %w", err)
	}

	return reviewQueueModelToDomain(result), nil
}

// Helper: pdfUploadModelToDomain converts database model to domain model
func pdfUploadModelToDomain(m models.PdfUpload) *domain.PDFUpload {
	var originalFilename string
	if m.OriginalFilename != nil {
		originalFilename = *m.OriginalFilename
	}

	var errorMsg sql.NullString
	if m.ErrorMessage != nil {
		errorMsg = sql.NullString{String: *m.ErrorMessage, Valid: true}
	}

	return &domain.PDFUpload{
		ID:               m.ID,
		LecturerID:       m.LecturerID,
		FilePath:         m.FilePath,
		FileName:         m.FileName,
		OriginalFilename: originalFilename,
		Status:           m.Status,
		ExtractionResult: m.ExtractionResult,
		ErrorMessage:     errorMsg,
		CreatedAt:        m.CreatedAt.Time,
		UpdatedAt:        m.UpdatedAt.Time,
	}
}

// Helper: reviewQueueModelToDomain converts database model to domain model
func reviewQueueModelToDomain(m models.ProblemReviewQueue) *domain.ProblemReviewQueue {
	var pdfUploadID int64
	if m.PdfUploadID != nil {
		pdfUploadID = *m.PdfUploadID
	}

	var problemNumber int
	if m.ProblemNumber != nil {
		problemNumber = int(*m.ProblemNumber)
	}

	var reviewerID sql.NullInt64
	if m.ReviewerID != nil {
		reviewerID = sql.NullInt64{Int64: *m.ReviewerID, Valid: true}
	}

	var reviewNotes sql.NullString
	if m.ReviewNotes != nil {
		reviewNotes = sql.NullString{String: *m.ReviewNotes, Valid: true}
	}

	var reviewedAt sql.NullTime
	if m.ReviewedAt.Valid {
		reviewedAt = sql.NullTime{Time: m.ReviewedAt.Time, Valid: true}
	}

	return &domain.ProblemReviewQueue{
		ID:            m.ID,
		PDFUploadID:   pdfUploadID,
		ProblemNumber: problemNumber,
		ProblemDraft:  m.ProblemDraft,
		Status:        m.Status,
		ReviewerID:    reviewerID,
		ReviewNotes:   reviewNotes,
		EditsMade:     m.EditsMade,
		ReviewedAt:    reviewedAt,
		CreatedAt:     m.CreatedAt.Time,
		UpdatedAt:     m.UpdatedAt.Time,
	}
}
