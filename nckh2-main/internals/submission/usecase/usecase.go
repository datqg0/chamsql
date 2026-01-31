package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"backend/configs"
	"backend/internals/problem/repository"
	"backend/internals/submission/controller/dto"
	submissionRepo "backend/internals/submission/repository"
	"backend/pkgs/runner"
	"backend/sql/models"
)

var (
	ErrProblemNotFound    = errors.New("problem not found")
	ErrSubmissionNotFound = errors.New("submission not found")
	ErrUnsupportedDB      = errors.New("database type not supported for this problem")
)

type ISubmissionUseCase interface {
	Run(ctx context.Context, problemID int64, req *dto.RunQueryRequest) (*dto.RunQueryResponse, error)
	Submit(ctx context.Context, userID, problemID int64, req *dto.SubmitQueryRequest) (*dto.SubmitQueryResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.SubmissionResponse, error)
	ListByUser(ctx context.Context, userID int64, page, pageSize int) (*dto.SubmissionListResponse, error)
}

type submissionUseCase struct {
	submissionRepo submissionRepo.ISubmissionRepository
	problemRepo    repository.IProblemRepository
	runner         runner.Runner
	cfg            *configs.Config
}

func NewSubmissionUseCase(
	subRepo submissionRepo.ISubmissionRepository,
	probRepo repository.IProblemRepository,
	queryRunner runner.Runner,
	cfg *configs.Config,
) ISubmissionUseCase {
	return &submissionUseCase{
		submissionRepo: subRepo,
		problemRepo:    probRepo,
		runner:         queryRunner,
		cfg:            cfg,
	}
}

func (u *submissionUseCase) Run(ctx context.Context, problemID int64, req *dto.RunQueryRequest) (*dto.RunQueryResponse, error) {
	// Get problem
	problem, err := u.problemRepo.GetByID(ctx, problemID)
	if err != nil {
		return nil, ErrProblemNotFound
	}

	// Check if database type is supported
	if !containsDB(problem.SupportedDatabases, req.DatabaseType) {
		return nil, ErrUnsupportedDB
	}

	// Execute user query
	dbType := runner.DBType(req.DatabaseType)
	result, err := u.runner.ExecuteWithSetup(ctx, dbType, problem.InitScript, req.Code)

	response := &dto.RunQueryResponse{
		ExecutionMs: result.ExecutionMs,
	}

	if err != nil || result.Error != "" {
		response.Success = false
		response.Error = result.Error
		response.ErrorType = result.ErrorType
		return response, nil
	}

	response.Success = true
	response.Columns = result.Columns
	response.Rows = result.Rows
	response.RowCount = result.RowCount
	return response, nil
}

func (u *submissionUseCase) Submit(ctx context.Context, userID, problemID int64, req *dto.SubmitQueryRequest) (*dto.SubmitQueryResponse, error) {
	// Get problem
	problem, err := u.problemRepo.GetByID(ctx, problemID)
	if err != nil {
		return nil, ErrProblemNotFound
	}

	// Check if database type is supported
	if !containsDB(problem.SupportedDatabases, req.DatabaseType) {
		return nil, ErrUnsupportedDB
	}

	dbType := runner.DBType(req.DatabaseType)

	// Execute expected query (solution)
	expectedResult, err := u.runner.ExecuteWithSetup(ctx, dbType, problem.InitScript, problem.SolutionQuery)
	if err != nil {
		// Provide more context if it's a runner error
		if expectedResult != nil && expectedResult.Error != "" {
			return nil, fmt.Errorf("solution query failed: %s (SQL Error: %s)", err.Error(), expectedResult.Error)
		}
		return nil, fmt.Errorf("failed to execute solution query: %w", err)
	}

	// Execute user query
	actualResult, err := u.runner.ExecuteWithSetup(ctx, dbType, problem.InitScript, req.Code)

	// Compare results
	orderMatters := ptrToBool(problem.OrderMatters)
	compareResult := u.runner.Compare(expectedResult, actualResult, orderMatters)

	// Determine status
	status := "wrong_answer"
	if actualResult.Error != "" {
		if actualResult.ErrorType == "timeout" {
			status = "timeout"
		} else {
			status = "error"
		}
	} else if compareResult.IsCorrect {
		status = "accepted"
	}

	// Convert results to JSON
	expectedJSON, _ := json.Marshal(expectedResult.Rows)
	actualJSON, _ := json.Marshal(actualResult.Rows)
	execTimeMs := int32(actualResult.ExecutionMs)

	// Save submission
	submission, err := u.submissionRepo.Create(ctx, models.CreateSubmissionParams{
		UserID:          userID,
		ProblemID:       problemID,
		Code:            req.Code,
		DatabaseType:    req.DatabaseType,
		Status:          status,
		ExecutionTimeMs: &execTimeMs,
		ExpectedOutput:  expectedJSON,
		ActualOutput:    actualJSON,
		ErrorMessage:    strPtr(actualResult.Error),
		IsCorrect:       &compareResult.IsCorrect,
	})
	if err != nil {
		return nil, err
	}

	return &dto.SubmitQueryResponse{
		ID:             submission.ID,
		IsCorrect:      compareResult.IsCorrect,
		Status:         status,
		ExecutionMs:    actualResult.ExecutionMs,
		Message:        compareResult.Message,
		ExpectedRows:   compareResult.ExpectedRows,
		ActualRows:     compareResult.ActualRows,
		ExpectedOutput: expectedJSON,
		ActualOutput:   actualJSON,
		Error:          actualResult.Error,
	}, nil
}

func (u *submissionUseCase) GetByID(ctx context.Context, id int64) (*dto.SubmissionResponse, error) {
	submission, err := u.submissionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrSubmissionNotFound
	}

	var execTime *int
	if submission.ExecutionTimeMs != nil {
		e := int(*submission.ExecutionTimeMs)
		execTime = &e
	}

	return &dto.SubmissionResponse{
		ID:              submission.ID,
		ProblemID:       submission.ProblemID,
		ProblemTitle:    submission.ProblemTitle,
		ProblemSlug:     submission.ProblemSlug,
		Code:            submission.Code,
		DatabaseType:    submission.DatabaseType,
		Status:          submission.Status,
		IsCorrect:       ptrToBool(submission.IsCorrect),
		ExecutionTimeMs: execTime,
		ExpectedOutput:  submission.ExpectedOutput,
		ActualOutput:    submission.ActualOutput,
		ErrorMessage:    ptrToStr(submission.ErrorMessage),
		SubmittedAt:     submission.SubmittedAt.Time.Format("2006-01-02T15:04:05Z"),
	}, nil
}

func (u *submissionUseCase) ListByUser(ctx context.Context, userID int64, page, pageSize int) (*dto.SubmissionListResponse, error) {
	offset := int32((page - 1) * pageSize)
	limit := int32(pageSize)

	submissions, err := u.submissionRepo.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	result := make([]dto.SubmissionResponse, len(submissions))
	for i, s := range submissions {
		var execTime *int
		if s.ExecutionTimeMs != nil {
			e := int(*s.ExecutionTimeMs)
			execTime = &e
		}
		result[i] = dto.SubmissionResponse{
			ID:              s.ID,
			ProblemID:       s.ProblemID,
			ProblemTitle:    s.ProblemTitle,
			ProblemSlug:     s.ProblemSlug,
			Code:            s.Code,
			DatabaseType:    s.DatabaseType,
			Status:          s.Status,
			IsCorrect:       ptrToBool(s.IsCorrect),
			ExecutionTimeMs: execTime,
			SubmittedAt:     s.SubmittedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
	}

	total, _ := u.submissionRepo.CountByUser(ctx, userID)

	return &dto.SubmissionListResponse{
		Submissions: result,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}

// Helper functions
func containsDB(dbs []string, db string) bool {
	for _, d := range dbs {
		if d == db {
			return true
		}
	}
	return false
}

func ptrToBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func ptrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
