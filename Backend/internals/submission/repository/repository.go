package repository

import (
	"context"

	"backend/db"
	"backend/sql/models"

	"github.com/jackc/pgx/v5/pgtype"
)

type ISubmissionRepository interface {
	Create(ctx context.Context, params models.CreateSubmissionParams) (*models.Submission, error)
	GetByID(ctx context.Context, id int64) (*models.GetSubmissionByIDRow, error)
	ListByUser(ctx context.Context, userID int64, limit, offset int32) ([]models.ListUserSubmissionsRow, error)
	ListByUserAndProblem(ctx context.Context, userID, problemID int64, limit int32) ([]models.Submission, error)
	CountByUser(ctx context.Context, userID int64) (int64, error)
	// Test Results
	CreateTestResult(ctx context.Context, params models.CreateSubmissionTestResultParams) (*models.SubmissionTestResult, error)
	ListTestResults(ctx context.Context, submissionID int64) ([]models.ListSubmissionTestResultsRow, error)
	UpdateScore(ctx context.Context, submissionID int64, score string, total, passed int32) error
}

type submissionRepository struct {
	db      *db.Database
	queries *models.Queries
}

func NewSubmissionRepository(database *db.Database) ISubmissionRepository {
	return &submissionRepository{
		db:      database,
		queries: models.New(database.GetPool()),
	}
}

func (r *submissionRepository) Create(ctx context.Context, params models.CreateSubmissionParams) (*models.Submission, error) {
	submission, err := r.queries.CreateSubmission(ctx, params)
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

func (r *submissionRepository) GetByID(ctx context.Context, id int64) (*models.GetSubmissionByIDRow, error) {
	submission, err := r.queries.GetSubmissionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

func (r *submissionRepository) ListByUser(ctx context.Context, userID int64, limit, offset int32) ([]models.ListUserSubmissionsRow, error) {
	return r.queries.ListUserSubmissions(ctx, models.ListUserSubmissionsParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
}

func (r *submissionRepository) ListByUserAndProblem(ctx context.Context, userID, problemID int64, limit int32) ([]models.Submission, error) {
	return r.queries.ListUserSubmissionsForProblem(ctx, models.ListUserSubmissionsForProblemParams{
		UserID:    userID,
		ProblemID: problemID,
		Limit:     limit,
	})
}

func (r *submissionRepository) CountByUser(ctx context.Context, userID int64) (int64, error) {
	return r.queries.CountUserSubmissions(ctx, userID)
}

func (r *submissionRepository) CreateTestResult(ctx context.Context, params models.CreateSubmissionTestResultParams) (*models.SubmissionTestResult, error) {
	tr, err := r.queries.CreateSubmissionTestResult(ctx, params)
	if err != nil {
		return nil, err
	}
	return &tr, nil
}

func (r *submissionRepository) ListTestResults(ctx context.Context, submissionID int64) ([]models.ListSubmissionTestResultsRow, error) {
	return r.queries.ListSubmissionTestResults(ctx, submissionID)
}

func (r *submissionRepository) UpdateScore(ctx context.Context, submissionID int64, score string, total, passed int32) error {
	var n pgtype.Numeric
	_ = n.Scan(score) // Assume valid score string from result calculation

	return r.queries.UpdateSubmissionScore(ctx, models.UpdateSubmissionScoreParams{
		ID:              submissionID,
		Score:           n,
		TotalTestCases:  &total,
		PassedTestCases: &passed,
	})
}
