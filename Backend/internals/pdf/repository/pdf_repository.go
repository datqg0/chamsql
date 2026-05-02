package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

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
	ResetStuckPDFUploads(ctx context.Context, timeout time.Duration) error

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

func mapPdfUpload(m models.PdfUpload) *domain.PDFUpload {
	errMsg := sql.NullString{Valid: false}
	if m.ErrorMessage != nil {
		errMsg = sql.NullString{String: *m.ErrorMessage, Valid: true}
	}

	return &domain.PDFUpload{
		ID:               m.ID,
		LecturerID:       m.LecturerID,
		FilePath:         m.FilePath,
		FileName:         m.FileName,
		OriginalFilename: m.OriginalFilename,
		Status:           m.Status,
		ExtractionResult: m.ExtractionResult,
		ErrorMessage:     errMsg,
		CreatedAt:        m.CreatedAt.Time,
		UpdatedAt:        m.UpdatedAt.Time,
	}
}

func mapProblemReviewQueue(m models.ProblemReviewQueue) *domain.ProblemReviewQueue {
	reviewerID := sql.NullInt64{Valid: false}
	if m.ReviewerID != nil {
		reviewerID = sql.NullInt64{Int64: *m.ReviewerID, Valid: true}
	}

	reviewNotes := sql.NullString{Valid: false}
	if m.ReviewNotes != nil {
		reviewNotes = sql.NullString{String: *m.ReviewNotes, Valid: true}
	}

	reviewedAt := sql.NullTime{Valid: false}
	if m.ReviewedAt.Valid {
		reviewedAt = sql.NullTime{Time: m.ReviewedAt.Time, Valid: true}
	}

	return &domain.ProblemReviewQueue{
		ID:            m.ID,
		PDFUploadID:   m.PdfUploadID,
		ProblemNumber: int(m.ProblemNumber),
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

func (r *pdfRepository) CreatePDFUpload(ctx context.Context, lecturerID int64, filePath, fileName, originalFilename string) (*domain.PDFUpload, error) {
	res, err := r.queries.CreatePDFUpload(ctx, models.CreatePDFUploadParams{
		LecturerID:       lecturerID,
		FilePath:         filePath,
		FileName:         fileName,
		OriginalFilename: originalFilename,
		Status:           "uploading",
	})
	if err != nil {
		return nil, err
	}
	return mapPdfUpload(res), nil
}

func (r *pdfRepository) GetPDFUploadByID(ctx context.Context, id int64) (*domain.PDFUpload, error) {
	res, err := r.queries.GetPDFUploadByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapPdfUpload(res), nil
}

func (r *pdfRepository) GetPDFUploadsByLecturer(ctx context.Context, lecturerID int64, limit, offset int32) ([]domain.PDFUpload, error) {
	return []domain.PDFUpload{}, nil
}

func (r *pdfRepository) UpdatePDFUploadStatus(ctx context.Context, id int64, status string) (*domain.PDFUpload, error) {
	res, err := r.queries.UpdatePDFUploadStatus(ctx, models.UpdatePDFUploadStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return nil, err
	}
	return mapPdfUpload(res), nil
}

func (r *pdfRepository) UpdatePDFUploadWithExtraction(ctx context.Context, id int64, status string, extractionResult []byte) (*domain.PDFUpload, error) {
	res, err := r.queries.UpdatePDFUploadWithExtraction(ctx, models.UpdatePDFUploadWithExtractionParams{
		ID:               id,
		Status:           status,
		ExtractionResult: extractionResult,
	})
	if err != nil {
		return nil, err
	}
	return mapPdfUpload(res), nil
}

func (r *pdfRepository) UpdatePDFUploadError(ctx context.Context, id int64, errorMessage string) (*domain.PDFUpload, error) {
	errMsg := errorMessage
	res, err := r.queries.UpdatePDFUploadError(ctx, models.UpdatePDFUploadErrorParams{
		ID:           id,
		ErrorMessage: &errMsg,
	})
	if err != nil {
		return nil, err
	}
	return mapPdfUpload(res), nil
}

func (r *pdfRepository) ResetStuckPDFUploads(ctx context.Context, timeout time.Duration) error {
	cutoff := time.Now().Add(-timeout)
	return r.queries.ResetStuckPDFUploads(ctx, pgtype.Timestamptz{
		Time:  cutoff,
		Valid: true,
	})
}

func (r *pdfRepository) CreateProblemReviewQueue(ctx context.Context, pdfUploadID int64, problemNumber int, problemDraft []byte) (*domain.ProblemReviewQueue, error) {
	res, err := r.queries.CreateProblemReviewQueue(ctx, models.CreateProblemReviewQueueParams{
		PdfUploadID:   pdfUploadID,
		ProblemNumber: int32(problemNumber),
		ProblemDraft:  problemDraft,
		Status:        "pending",
	})
	if err != nil {
		return nil, err
	}
	return mapProblemReviewQueue(res), nil
}

func (r *pdfRepository) GetProblemReviewQueueByID(ctx context.Context, id int64) (*domain.ProblemReviewQueue, error) {
	res, err := r.queries.GetProblemReviewQueueByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapProblemReviewQueue(res), nil
}

func (r *pdfRepository) GetProblemReviewQueueByPDF(ctx context.Context, pdfUploadID int64) ([]domain.ProblemReviewQueue, error) {
	res, err := r.queries.GetProblemReviewQueueByPDF(ctx, pdfUploadID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.ProblemReviewQueue, len(res))
	for i, r := range res {
		result[i] = *mapProblemReviewQueue(r)
	}
	return result, nil
}

func (r *pdfRepository) GetProblemReviewQueueByStatus(ctx context.Context, status string, limit, offset int32) ([]domain.ProblemReviewQueue, error) {
	res, err := r.queries.GetProblemReviewQueueByStatus(ctx, models.GetProblemReviewQueueByStatusParams{
		Status: status,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	result := make([]domain.ProblemReviewQueue, len(res))
	for i, r := range res {
		result[i] = *mapProblemReviewQueue(r)
	}
	return result, nil
}

func (r *pdfRepository) UpdateProblemReviewStatus(ctx context.Context, id int64, status string, reviewerID int64, reviewNotes string) (*domain.ProblemReviewQueue, error) {
	return nil, fmt.Errorf("UpdateProblemReviewStatus not implemented in sqlc")
}

func (r *pdfRepository) UpdateProblemReviewDraft(ctx context.Context, id int64, problemDraft, editsMade []byte) (*domain.ProblemReviewQueue, error) {
	return nil, fmt.Errorf("UpdateProblemReviewDraft not implemented in sqlc")
}
