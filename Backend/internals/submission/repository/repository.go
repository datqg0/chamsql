package repository

import (
	"context"

	"backend/db"
	"backend/sql/models"
)

type ISubmissionRepository interface {
	Create(ctx context.Context, params models.CreateSubmissionParams) (*models.Submission, error)
	GetByID(ctx context.Context, id int64) (*models.GetSubmissionByIDRow, error)
	ListByUser(ctx context.Context, userID int64, limit, offset int32) ([]models.ListUserSubmissionsRow, error)
	ListByUserAndProblem(ctx context.Context, userID, problemID int64, limit int32) ([]models.Submission, error)
	CountByUser(ctx context.Context, userID int64) (int64, error)
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
