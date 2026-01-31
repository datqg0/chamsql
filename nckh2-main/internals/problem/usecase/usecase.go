package usecase

import (
	"context"
	"errors"

	"backend/internals/problem/controller/dto"
	"backend/internals/problem/repository"
	"backend/sql/models"
)

var (
	ErrProblemNotFound = errors.New("problem not found")
	ErrSlugExists      = errors.New("problem slug already exists")
)

type IProblemUseCase interface {
	Create(ctx context.Context, userID int64, req *dto.CreateProblemRequest) (*dto.ProblemResponse, error)
	GetBySlug(ctx context.Context, slug string, userID *int64) (*dto.ProblemResponse, error)
	List(ctx context.Context, query *dto.ProblemListQuery) (*dto.ProblemListResponse, error)
	Update(ctx context.Context, id int64, req *dto.UpdateProblemRequest) (*dto.ProblemResponse, error)
	Delete(ctx context.Context, id int64) error
}

type problemUseCase struct {
	repo repository.IProblemRepository
}

func NewProblemUseCase(repo repository.IProblemRepository) IProblemUseCase {
	return &problemUseCase{repo: repo}
}

func (u *problemUseCase) Create(ctx context.Context, userID int64, req *dto.CreateProblemRequest) (*dto.ProblemResponse, error) {
	// Check if slug exists
	_, err := u.repo.GetBySlug(ctx, req.Slug)
	if err == nil {
		return nil, ErrSlugExists
	}

	isPublic := req.IsPublic
	orderMatters := req.OrderMatters

	problem, err := u.repo.Create(ctx, models.CreateProblemParams{
		Title:              req.Title,
		Slug:               req.Slug,
		Description:        req.Description,
		Difficulty:         req.Difficulty,
		TopicID:            req.TopicID,
		CreatedBy:          &userID,
		InitScript:         req.InitScript,
		SolutionQuery:      req.SolutionQuery,
		SupportedDatabases: req.SupportedDatabases,
		OrderMatters:       &orderMatters,
		Hints:              req.Hints,
		SampleOutput:       req.SampleOutput,
		IsPublic:           &isPublic,
	})
	if err != nil {
		return nil, err
	}

	return toProblemResponse(problem), nil
}

func (u *problemUseCase) GetBySlug(ctx context.Context, slug string, userID *int64) (*dto.ProblemResponse, error) {
	if userID != nil {
		// Get with user progress
		problem, err := u.repo.GetWithUserProgress(ctx, slug, *userID)
		if err != nil {
			return nil, ErrProblemNotFound
		}
		return toProblemWithProgressResponse(problem), nil
	}

	problem, err := u.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, ErrProblemNotFound
	}
	return toProblemResponse(problem), nil
}

func (u *problemUseCase) List(ctx context.Context, query *dto.ProblemListQuery) (*dto.ProblemListResponse, error) {
	offset := int32((query.Page - 1) * query.PageSize)
	limit := int32(query.PageSize)

	var problems []dto.ProblemResponse
	var total int64

	if query.TopicID != nil {
		rows, err := u.repo.ListByTopic(ctx, *query.TopicID, limit, offset)
		if err != nil {
			return nil, err
		}
		problems = make([]dto.ProblemResponse, len(rows))
		for i, row := range rows {
			problems[i] = dto.ProblemResponse{
				ID:                 row.ID,
				Title:              row.Title,
				Slug:               row.Slug,
				Description:        row.Description,
				Difficulty:         row.Difficulty,
				SupportedDatabases: row.SupportedDatabases,
				TopicName:          ptrToStr(row.TopicName),
				TopicSlug:          ptrToStr(row.TopicSlug),
			}
		}
	} else if query.Difficulty != nil {
		rows, err := u.repo.ListByDifficulty(ctx, *query.Difficulty, limit, offset)
		if err != nil {
			return nil, err
		}
		problems = make([]dto.ProblemResponse, len(rows))
		for i, row := range rows {
			problems[i] = dto.ProblemResponse{
				ID:                 row.ID,
				Title:              row.Title,
				Slug:               row.Slug,
				Description:        row.Description,
				Difficulty:         row.Difficulty,
				SupportedDatabases: row.SupportedDatabases,
				TopicName:          ptrToStr(row.TopicName),
				TopicSlug:          ptrToStr(row.TopicSlug),
			}
		}
	} else {
		rows, err := u.repo.List(ctx, limit, offset)
		if err != nil {
			return nil, err
		}
		problems = make([]dto.ProblemResponse, len(rows))
		for i, row := range rows {
			problems[i] = dto.ProblemResponse{
				ID:                 row.ID,
				Title:              row.Title,
				Slug:               row.Slug,
				Description:        row.Description,
				Difficulty:         row.Difficulty,
				SupportedDatabases: row.SupportedDatabases,
				TopicName:          ptrToStr(row.TopicName),
				TopicSlug:          ptrToStr(row.TopicSlug),
			}
		}
	}

	total, _ = u.repo.Count(ctx)

	return &dto.ProblemListResponse{
		Problems: problems,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

func (u *problemUseCase) Update(ctx context.Context, id int64, req *dto.UpdateProblemRequest) (*dto.ProblemResponse, error) {
	// Check if problem exists
	_, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrProblemNotFound
	}

	params := models.UpdateProblemParams{ID: id}
	if req.Title != nil {
		params.Title = req.Title
	}
	if req.Description != nil {
		params.Description = req.Description
	}
	if req.Difficulty != nil {
		params.Difficulty = req.Difficulty
	}
	if req.TopicID != nil {
		params.TopicID = req.TopicID
	}
	if req.InitScript != nil {
		params.InitScript = req.InitScript
	}
	if req.SolutionQuery != nil {
		params.SolutionQuery = req.SolutionQuery
	}
	if req.OrderMatters != nil {
		params.OrderMatters = req.OrderMatters
	}
	if req.Hints != nil {
		params.Hints = req.Hints
	}
	if req.SampleOutput != nil {
		params.SampleOutput = req.SampleOutput
	}
	if req.IsPublic != nil {
		params.IsPublic = req.IsPublic
	}

	problem, err := u.repo.Update(ctx, params)
	if err != nil {
		return nil, err
	}

	return toProblemResponse(problem), nil
}

func (u *problemUseCase) Delete(ctx context.Context, id int64) error {
	_, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return ErrProblemNotFound
	}
	return u.repo.Delete(ctx, id)
}

// Helper functions
func toProblemResponse(p *models.Problem) *dto.ProblemResponse {
	return &dto.ProblemResponse{
		ID:                 p.ID,
		Title:              p.Title,
		Slug:               p.Slug,
		Description:        p.Description,
		Difficulty:         p.Difficulty,
		TopicID:            p.TopicID,
		InitScript:         p.InitScript,
		SolutionQuery:      p.SolutionQuery,
		SupportedDatabases: p.SupportedDatabases,
		OrderMatters:       ptrToBool(p.OrderMatters),
		Hints:              p.Hints,
		SampleOutput:       p.SampleOutput,
		IsPublic:           ptrToBool(p.IsPublic),
		CreatedBy:          p.CreatedBy,
	}
}

func toProblemWithProgressResponse(p *models.GetProblemWithUserProgressRow) *dto.ProblemResponse {
	var attempts *int
	if p.Attempts != nil {
		a := int(*p.Attempts)
		attempts = &a
	}
	var bestTime *int
	if p.BestTimeMs != nil {
		b := int(*p.BestTimeMs)
		bestTime = &b
	}

	return &dto.ProblemResponse{
		ID:                 p.ID,
		Title:              p.Title,
		Slug:               p.Slug,
		Description:        p.Description,
		Difficulty:         p.Difficulty,
		TopicID:            p.TopicID,
		TopicName:          ptrToStr(p.TopicName),
		TopicSlug:          ptrToStr(p.TopicSlug),
		InitScript:         p.InitScript,
		SolutionQuery:      p.SolutionQuery,
		SupportedDatabases: p.SupportedDatabases,
		OrderMatters:       ptrToBool(p.OrderMatters),
		Hints:              p.Hints,
		SampleOutput:       p.SampleOutput,
		IsPublic:           ptrToBool(p.IsPublic),
		CreatedBy:          p.CreatedBy,
		IsSolved:           p.IsSolved,
		Attempts:           attempts,
		BestTimeMs:         bestTime,
	}
}

func ptrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptrToBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}
