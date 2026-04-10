package usecase

import (
	"context"
	"errors"
	"time"

	"backend/internals/problem/controller/dto"
	"backend/internals/problem/repository"
	"backend/pkgs/redis"
	"backend/sql/models"
)

var (
	ErrProblemNotFound = errors.New("problem not found")
	ErrSlugExists      = errors.New("problem slug already exists")
	ErrForbidden       = errors.New("you don't have permission to modify this problem")
)

type IProblemUseCase interface {
	Create(ctx context.Context, userID int64, req *dto.CreateProblemRequest) (*dto.ProblemResponse, error)
	GetBySlug(ctx context.Context, slug string, userID *int64) (*dto.ProblemResponse, error)
	List(ctx context.Context, role string, query *dto.ProblemListQuery) (*dto.ProblemListResponse, error)
	Update(ctx context.Context, userID int64, id int64, req *dto.UpdateProblemRequest) (*dto.ProblemResponse, error)
	Delete(ctx context.Context, userID int64, id int64) error
}

type problemUseCase struct {
	repo  repository.IProblemRepository
	cache redis.IRedis
}

func NewProblemUseCase(repo repository.IProblemRepository, cache redis.IRedis) IProblemUseCase {
	return &problemUseCase{
		repo:  repo,
		cache: cache,
	}
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

	// Create test cases if provided
	var testCases []models.ProblemTestCase
	if len(req.TestCases) > 0 {
		testCases = make([]models.ProblemTestCase, 0, len(req.TestCases))
		for _, tcReq := range req.TestCases {
			tc, err := u.repo.CreateTestCase(ctx, models.CreateProblemTestCaseParams{
				ProblemID:     problem.ID,
				Name:          &tcReq.Name,
				Description:   &tcReq.Description,
				InitScript:    tcReq.InitScript,
				SolutionQuery: tcReq.SolutionQuery,
				Weight:        &tcReq.Weight,
				IsHidden:      &tcReq.IsHidden,
			})
			if err != nil {
				return nil, err
			}
			testCases = append(testCases, *tc)
		}
	}

	return toProblemResponse(problem, testCases), nil
}

func (u *problemUseCase) GetBySlug(ctx context.Context, slug string, userID *int64) (*dto.ProblemResponse, error) {
	if userID != nil {
		// Get with user progress
		problem, err := u.repo.GetWithUserProgress(ctx, slug, *userID)
		if err != nil {
			return nil, ErrProblemNotFound
		}

		testCases, _ := u.repo.ListTestCases(ctx, problem.ID)

		return toProblemWithProgressResponse(problem, testCases), nil
	}

	// Try to get from cache first (cache public problems)
	cacheKey := "problem:" + slug
	if u.cache != nil {
		var cached dto.ProblemResponse
		if err := u.cache.Get(cacheKey, &cached); err == nil {
			return &cached, nil
		}
	}

	problem, err := u.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, ErrProblemNotFound
	}

	// Get test cases
	testCases, _ := u.repo.ListTestCases(ctx, problem.ID)

	response := toProblemResponse(problem, testCases)

	// Cache for 24 hours
	if u.cache != nil && (problem.IsPublic == nil || *problem.IsPublic) {
		u.cache.SetWithExpiration(cacheKey, response, 24*time.Hour)
	}

	return response, nil
}

func (u *problemUseCase) List(ctx context.Context, role string, query *dto.ProblemListQuery) (*dto.ProblemListResponse, error) {
	offset := int32((query.Page - 1) * query.PageSize)
	limit := int32(query.PageSize)

	var problems []dto.ProblemResponse
	var total int64

	isAdmin := role == "admin" || role == "lecturer"

	if query.TopicID != nil {
		var rows []models.ListProblemsByTopicRow
		var err error
		if isAdmin {
			rows, err = u.repo.ListAdminByTopic(ctx, *query.TopicID, limit, offset)
		} else {
			rows, err = u.repo.ListByTopic(ctx, *query.TopicID, limit, offset)
		}
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
				IsPublic:           ptrToBool(row.IsPublic),
			}
		}
	} else if query.Difficulty != nil {
		var rows []models.ListProblemsByDifficultyRow
		var err error
		if isAdmin {
			rows, err = u.repo.ListAdminByDifficulty(ctx, *query.Difficulty, limit, offset)
		} else {
			rows, err = u.repo.ListByDifficulty(ctx, *query.Difficulty, limit, offset)
		}
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
				IsPublic:           ptrToBool(row.IsPublic),
			}
		}
	} else {
		var rows []models.ListProblemsRow
		var err error
		if isAdmin {
			rows, err = u.repo.ListAdmin(ctx, limit, offset)
		} else {
			rows, err = u.repo.List(ctx, limit, offset)
		}
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
				IsPublic:           ptrToBool(row.IsPublic),
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

func (u *problemUseCase) Update(ctx context.Context, userID int64, id int64, req *dto.UpdateProblemRequest) (*dto.ProblemResponse, error) {
	// Check if problem exists
	problem, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrProblemNotFound
	}

	// Check ownership
	if problem.CreatedBy == nil || *problem.CreatedBy != userID {
		return nil, ErrForbidden
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

	updatedProblem, err := u.repo.Update(ctx, params)
	if err != nil {
		return nil, err
	}

	// Update test cases if provided (replace all)
	if req.TestCases != nil {
		_ = u.repo.DeleteAllTestCases(ctx, updatedProblem.ID)
		for _, tcReq := range req.TestCases {
			_, _ = u.repo.CreateTestCase(ctx, models.CreateProblemTestCaseParams{
				ProblemID:     updatedProblem.ID,
				Name:          &tcReq.Name,
				Description:   &tcReq.Description,
				InitScript:    tcReq.InitScript,
				SolutionQuery: tcReq.SolutionQuery,
				Weight:        &tcReq.Weight,
				IsHidden:      &tcReq.IsHidden,
			})
		}
	}

	// Fetch the full updated problem to ensure we have the slug and other fields
	finalProblem, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get final test cases
	testCases, _ := u.repo.ListTestCases(ctx, finalProblem.ID)

	// Invalidate cache with old slug
	if u.cache != nil && problem.Slug != "" {
		u.cache.Remove("problem:" + problem.Slug)
	}

	return toProblemResponse(finalProblem, testCases), nil
}

func (u *problemUseCase) Delete(ctx context.Context, userID int64, id int64) error {
	problem, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return ErrProblemNotFound
	}

	// Check ownership
	if problem.CreatedBy == nil || *problem.CreatedBy != userID {
		return ErrForbidden
	}

	// Invalidate cache before deletion
	if u.cache != nil && problem.Slug != "" {
		u.cache.Remove("problem:" + problem.Slug)
	}

	return u.repo.Delete(ctx, id)
}

// Helper functions
func toProblemResponse(p *models.Problem, testCases []models.ProblemTestCase) *dto.ProblemResponse {
	tcResponses := make([]dto.TestCaseResponse, len(testCases))
	for i, tc := range testCases {
		tcResponses[i] = dto.TestCaseResponse{
			ID:            tc.ID,
			Name:          ptrToStr(tc.Name),
			Description:   ptrToStr(tc.Description),
			InitScript:    tc.InitScript,
			SolutionQuery: tc.SolutionQuery,
			Weight:        ptrToInt32(tc.Weight),
			IsHidden:      ptrToBool(tc.IsHidden),
		}
	}

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
		TestCases:          tcResponses,
	}
}

func toProblemWithProgressResponse(p *models.GetProblemWithUserProgressRow, testCases []models.ProblemTestCase) *dto.ProblemResponse {
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

	tcResponses := make([]dto.TestCaseResponse, len(testCases))
	for i, tc := range testCases {
		tcResponses[i] = dto.TestCaseResponse{
			ID:            tc.ID,
			Name:          ptrToStr(tc.Name),
			Description:   ptrToStr(tc.Description),
			InitScript:    tc.InitScript,
			SolutionQuery: tc.SolutionQuery,
			Weight:        ptrToInt32(tc.Weight),
			IsHidden:      ptrToBool(tc.IsHidden),
		}
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
		TestCases:          tcResponses,
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

func ptrToInt32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}
