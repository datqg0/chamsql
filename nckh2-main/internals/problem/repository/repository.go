package repository

import (
	"context"

	"backend/db"
	"backend/sql/models"
)

type IProblemRepository interface {
	Create(ctx context.Context, params models.CreateProblemParams) (*models.Problem, error)
	GetByID(ctx context.Context, id int64) (*models.Problem, error)
	GetBySlug(ctx context.Context, slug string) (*models.Problem, error)
	GetWithUserProgress(ctx context.Context, slug string, userID int64) (*models.GetProblemWithUserProgressRow, error)
	List(ctx context.Context, limit, offset int32) ([]models.ListProblemsRow, error)
	ListByTopic(ctx context.Context, topicID int32, limit, offset int32) ([]models.ListProblemsByTopicRow, error)
	ListByDifficulty(ctx context.Context, difficulty string, limit, offset int32) ([]models.ListProblemsByDifficultyRow, error)
	Update(ctx context.Context, params models.UpdateProblemParams) (*models.Problem, error)
	Delete(ctx context.Context, id int64) error
	Count(ctx context.Context) (int64, error)
}

type problemRepository struct {
	db      *db.Database
	queries *models.Queries
}

func NewProblemRepository(database *db.Database) IProblemRepository {
	return &problemRepository{
		db:      database,
		queries: models.New(database.GetPool()),
	}
}

func (r *problemRepository) Create(ctx context.Context, params models.CreateProblemParams) (*models.Problem, error) {
	problem, err := r.queries.CreateProblem(ctx, params)
	if err != nil {
		return nil, err
	}
	return &problem, nil
}

func (r *problemRepository) GetByID(ctx context.Context, id int64) (*models.Problem, error) {
	problem, err := r.queries.GetProblemByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &problem, nil
}

func (r *problemRepository) GetBySlug(ctx context.Context, slug string) (*models.Problem, error) {
	problem, err := r.queries.GetProblemBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return &problem, nil
}

func (r *problemRepository) GetWithUserProgress(ctx context.Context, slug string, userID int64) (*models.GetProblemWithUserProgressRow, error) {
	problem, err := r.queries.GetProblemWithUserProgress(ctx, models.GetProblemWithUserProgressParams{
		Slug:   slug,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return &problem, nil
}

func (r *problemRepository) List(ctx context.Context, limit, offset int32) ([]models.ListProblemsRow, error) {
	return r.queries.ListProblems(ctx, models.ListProblemsParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *problemRepository) ListByTopic(ctx context.Context, topicID int32, limit, offset int32) ([]models.ListProblemsByTopicRow, error) {
	return r.queries.ListProblemsByTopic(ctx, models.ListProblemsByTopicParams{
		TopicID: &topicID,
		Limit:   limit,
		Offset:  offset,
	})
}

func (r *problemRepository) ListByDifficulty(ctx context.Context, difficulty string, limit, offset int32) ([]models.ListProblemsByDifficultyRow, error) {
	return r.queries.ListProblemsByDifficulty(ctx, models.ListProblemsByDifficultyParams{
		Difficulty: difficulty,
		Limit:      limit,
		Offset:     offset,
	})
}

func (r *problemRepository) Update(ctx context.Context, params models.UpdateProblemParams) (*models.Problem, error) {
	problem, err := r.queries.UpdateProblem(ctx, params)
	if err != nil {
		return nil, err
	}
	return &problem, nil
}

func (r *problemRepository) Delete(ctx context.Context, id int64) error {
	return r.queries.DeleteProblem(ctx, id)
}

func (r *problemRepository) Count(ctx context.Context) (int64, error) {
	return r.queries.CountProblems(ctx)
}
