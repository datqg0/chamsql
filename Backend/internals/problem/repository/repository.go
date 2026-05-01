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
	ListAdmin(ctx context.Context, limit, offset int32) ([]models.ListProblemsAdminRow, error)
	ListByCreator(ctx context.Context, creatorID int64, limit, offset int32) ([]models.ListProblemsByCreatorRow, error)
	CountByCreator(ctx context.Context, creatorID int64) (int64, error)
	ListAdminByTopic(ctx context.Context, topicID int32, limit, offset int32) ([]models.ListProblemsByTopicAdminRow, error)
	ListAdminByDifficulty(ctx context.Context, difficulty string, limit, offset int32) ([]models.ListProblemsByDifficultyAdminRow, error)
	// Search
	Search(ctx context.Context, query string, limit, offset int32) ([]models.SearchProblemsRow, error)
	SearchAdmin(ctx context.Context, query string, limit, offset int32) ([]models.SearchProblemsAdminRow, error)
	CountSearch(ctx context.Context, query string) (int64, error)
	Update(ctx context.Context, params models.UpdateProblemParams) (*models.Problem, error)
	Delete(ctx context.Context, id int64) error
	Count(ctx context.Context) (int64, error)
	CountAdmin(ctx context.Context) (int64, error)
	// Test Case Management
	CreateTestCase(ctx context.Context, params models.CreateProblemTestCaseParams) (*models.ProblemTestCase, error)
	ListTestCases(ctx context.Context, problemID int64) ([]models.ProblemTestCase, error)
	DeleteAllTestCases(ctx context.Context, problemID int64) error
	// User Progress
	UpsertProgress(ctx context.Context, userID, problemID int64) error
	MarkProblemSolved(ctx context.Context, userID, problemID int64, bestTimeMs int32) error
	GetProgress(ctx context.Context, userID, problemID int64) (*models.UserProgress, error)
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

func (r *problemRepository) ListAdmin(ctx context.Context, limit, offset int32) ([]models.ListProblemsAdminRow, error) {
	return r.queries.ListProblemsAdmin(ctx, models.ListProblemsAdminParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *problemRepository) ListByCreator(ctx context.Context, creatorID int64, limit, offset int32) ([]models.ListProblemsByCreatorRow, error) {
	return r.queries.ListProblemsByCreator(ctx, models.ListProblemsByCreatorParams{
		CreatedBy: &creatorID,
		Limit:     limit,
		Offset:    offset,
	})
}

func (r *problemRepository) CountByCreator(ctx context.Context, creatorID int64) (int64, error) {
	return r.queries.CountProblemsByCreator(ctx, &creatorID)
}

func (r *problemRepository) ListAdminByTopic(ctx context.Context, topicID int32, limit, offset int32) ([]models.ListProblemsByTopicAdminRow, error) {
	return r.queries.ListProblemsByTopicAdmin(ctx, models.ListProblemsByTopicAdminParams{
		TopicID: &topicID,
		Limit:   limit,
		Offset:  offset,
	})
}

func (r *problemRepository) ListAdminByDifficulty(ctx context.Context, difficulty string, limit, offset int32) ([]models.ListProblemsByDifficultyAdminRow, error) {
	return r.queries.ListProblemsByDifficultyAdmin(ctx, models.ListProblemsByDifficultyAdminParams{
		Difficulty: difficulty,
		Limit:      limit,
		Offset:     offset,
	})
}

func (r *problemRepository) Search(ctx context.Context, query string, limit, offset int32) ([]models.SearchProblemsRow, error) {
	return r.queries.SearchProblems(ctx, models.SearchProblemsParams{
		SearchQuery: query,
		Limit:       limit,
		Offset:      offset,
	})
}

func (r *problemRepository) SearchAdmin(ctx context.Context, query string, limit, offset int32) ([]models.SearchProblemsAdminRow, error) {
	return r.queries.SearchProblemsAdmin(ctx, models.SearchProblemsAdminParams{
		SearchQuery: query,
		Limit:       limit,
		Offset:      offset,
	})
}

func (r *problemRepository) CountSearch(ctx context.Context, query string) (int64, error) {
	return r.queries.CountSearchProblems(ctx, query)
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

func (r *problemRepository) CountAdmin(ctx context.Context) (int64, error) {
	return r.queries.CountProblemsAdmin(ctx)
}

func (r *problemRepository) CreateTestCase(ctx context.Context, params models.CreateProblemTestCaseParams) (*models.ProblemTestCase, error) {
	tc, err := r.queries.CreateProblemTestCase(ctx, params)
	if err != nil {
		return nil, err
	}
	return &tc, nil
}

func (r *problemRepository) ListTestCases(ctx context.Context, problemID int64) ([]models.ProblemTestCase, error) {
	return r.queries.ListProblemTestCases(ctx, problemID)
}

func (r *problemRepository) DeleteAllTestCases(ctx context.Context, problemID int64) error {
	return r.queries.DeleteAllProblemTestCases(ctx, problemID)
}

func (r *problemRepository) UpsertProgress(ctx context.Context, userID, problemID int64) error {
	_, err := r.queries.UpsertProgress(ctx, models.UpsertProgressParams{
		UserID:    userID,
		ProblemID: problemID,
	})
	return err
}

func (r *problemRepository) MarkProblemSolved(ctx context.Context, userID, problemID int64, bestTimeMs int32) error {
	_, err := r.queries.MarkProblemSolved(ctx, models.MarkProblemSolvedParams{
		UserID:     userID,
		ProblemID:  problemID,
		BestTimeMs: &bestTimeMs,
	})
	return err
}

func (r *problemRepository) GetProgress(ctx context.Context, userID, problemID int64) (*models.UserProgress, error) {
	progress, err := r.queries.GetUserProblemProgress(ctx, models.GetUserProblemProgressParams{
		UserID:    userID,
		ProblemID: problemID,
	})
	if err != nil {
		return nil, err
	}
	return &progress, nil
}
