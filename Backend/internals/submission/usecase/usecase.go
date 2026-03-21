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

	// Get test cases
	testCases, _ := u.problemRepo.ListTestCases(ctx, problemID)

	var (
		totalWeight      int32
		passedWeight     int32
		passedTests      int
		finalStatus      = "accepted"
		totalExecTime    int64
		testResults      []dto.TestResultResponse
		firstError       string
	)

	// If no test cases, use problem's default init_script and solution_query as the single test case
	if len(testCases) == 0 {
		testCases = []models.ProblemTestCase{
			{
				ID:            0,
				ProblemID:     problemID,
				InitScript:    problem.InitScript,
				SolutionQuery: problem.SolutionQuery,
				Weight:        ptrToInt32Ptr(1),
			},
		}
	}

	testResults = make([]dto.TestResultResponse, 0, len(testCases))
	for _, tc := range testCases {
		weight := ptrToInt32Val(tc.Weight)
		totalWeight += weight

		// Execute expected query
		expectedResult, err := u.runner.ExecuteWithSetup(ctx, dbType, tc.InitScript, tc.SolutionQuery)
		if err != nil {
			// This is a system/problem error
			continue
		}

		// Execute user query
		actualResult, err := u.runner.ExecuteWithSetup(ctx, dbType, tc.InitScript, req.Code)
		totalExecTime += actualResult.ExecutionMs

		// Compare
		orderMatters := ptrToBool(problem.OrderMatters)
		compareResult := u.runner.Compare(expectedResult, actualResult, orderMatters)

		trStatus := "wrong_answer"
		if actualResult.Error != "" {
			if actualResult.ErrorType == "timeout" {
				trStatus = "timeout"
				if finalStatus == "accepted" {
					finalStatus = "timeout"
				}
			} else {
				trStatus = "error"
				if finalStatus == "accepted" {
					finalStatus = "error"
				}
			}
			if firstError == "" {
				firstError = actualResult.Error
			}
		} else if compareResult.IsCorrect {
			trStatus = "accepted"
			passedWeight += weight
			passedTests++
		} else {
			if finalStatus == "accepted" {
				finalStatus = "wrong_answer"
			}
		}

		testResults = append(testResults, dto.TestResultResponse{
			TestCaseID:   tc.ID,
			TestCaseName: ptrToStr(tc.Name),
			Status:       trStatus,
			ExecutionMs:  actualResult.ExecutionMs,
			IsCorrect:    compareResult.IsCorrect,
			IsHidden:     ptrToBool(tc.IsHidden),
			ActualOutput: marshalJSON(actualResult.Rows),
			ErrorMessage: actualResult.Error,
		})

	}

	if passedTests < len(testCases) && finalStatus == "accepted" {
		finalStatus = "wrong_answer"
	}
	
	score := 0.0
	if totalWeight > 0 {
		score = (float64(passedWeight) / float64(totalWeight)) * 10.0 // Scale to 10
	}

	execTimeMs := int32(totalExecTime)
	isCorrectFinal := passedTests == len(testCases)

	// Save main submission
	submission, err := u.submissionRepo.Create(ctx, models.CreateSubmissionParams{
		UserID:          userID,
		ProblemID:       problemID,
		Code:            req.Code,
		DatabaseType:    req.DatabaseType,
		Status:          finalStatus,
		ExecutionTimeMs: &execTimeMs,
		ErrorMessage:    strPtr(firstError),
		IsCorrect:       &isCorrectFinal,
	})
	if err != nil {
		return nil, err
	}

	// Update score and totals
	scoreStr := fmt.Sprintf("%.2f", score)
	_ = u.submissionRepo.UpdateScore(ctx, submission.ID, scoreStr, int32(len(testCases)), int32(passedTests))

	// Save individual test results
	for _, tr := range testResults {
		_, _ = u.submissionRepo.CreateTestResult(ctx, models.CreateSubmissionTestResultParams{
			SubmissionID:    submission.ID,
			TestCaseID:      tr.TestCaseID,
			Status:          tr.Status,
			ExecutionTimeMs: ptrToInt32Ptr(int32(tr.ExecutionMs)),
			ActualOutput:    tr.ActualOutput,
			ErrorMessage:    strPtr(tr.ErrorMessage),
			IsCorrect:       &tr.IsCorrect,
		})
	}

	return &dto.SubmitQueryResponse{
		ID:             submission.ID,
		IsCorrect:      isCorrectFinal,
		Status:         finalStatus,
		ExecutionMs:    totalExecTime,
		Score:          score,
		TotalTests:     len(testCases),
		PassedTests:    passedTests,
		Message:        fmt.Sprintf("Passed %d/%d test cases", passedTests, len(testCases)),
		Error:          firstError,
		TestResults:    testResults,
	}, nil
}

func (u *submissionUseCase) GetByID(ctx context.Context, id int64) (*dto.SubmissionResponse, error) {
	submission, err := u.submissionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrSubmissionNotFound
	}

	// Get test results
	testResults, _ := u.submissionRepo.ListTestResults(ctx, id)

	return toSubmissionResponse(submission, testResults), nil
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
		
		var score float64
		_ = s.Score.Scan(&score)

		result[i] = dto.SubmissionResponse{
			ID:              s.ID,
			ProblemID:       s.ProblemID,
			ProblemTitle:    s.ProblemTitle,
			ProblemSlug:     s.ProblemSlug,
			Code:            s.Code,
			DatabaseType:    s.DatabaseType,
			Status:          s.Status,
			IsCorrect:       ptrToBool(s.IsCorrect),
			Score:           score,
			TotalTests:      int(ptrToInt32Val(s.TotalTestCases)),
			PassedTests:     int(ptrToInt32Val(s.PassedTestCases)),
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
func toSubmissionResponse(s *models.GetSubmissionByIDRow, testResults []models.ListSubmissionTestResultsRow) *dto.SubmissionResponse {
	var execTime *int
	if s.ExecutionTimeMs != nil {
		e := int(*s.ExecutionTimeMs)
		execTime = &e
	}

	var score float64
	_ = s.Score.Scan(&score)

	trResponses := make([]dto.TestResultResponse, len(testResults))
	for i, tr := range testResults {
		trResponses[i] = dto.TestResultResponse{
			TestCaseID:   tr.TestCaseID,
			TestCaseName: ptrToStr(tr.TestCaseName),
			Status:       tr.Status,
			ExecutionMs:  int64(ptrToInt32Val(tr.ExecutionTimeMs)),
			IsCorrect:    ptrToBool(tr.IsCorrect),
			IsHidden:     ptrToBool(tr.IsHidden),
			ActualOutput: tr.ActualOutput,
			ErrorMessage: ptrToStr(tr.ErrorMessage),
		}
	}

	return &dto.SubmissionResponse{
		ID:              s.ID,
		ProblemID:       s.ProblemID,
		ProblemTitle:    s.ProblemTitle,
		ProblemSlug:     s.ProblemSlug,
		Code:            s.Code,
		DatabaseType:    s.DatabaseType,
		Status:          s.Status,
		IsCorrect:       ptrToBool(s.IsCorrect),
		Score:           score,
		TotalTests:      int(ptrToInt32Val(s.TotalTestCases)),
		PassedTests:     int(ptrToInt32Val(s.PassedTestCases)),
		ExecutionTimeMs: execTime,
		ErrorMessage:    ptrToStr(s.ErrorMessage),
		SubmittedAt:     s.SubmittedAt.Time.Format("2006-01-02T15:04:05Z"),
		TestResults:     trResponses,
	}
}

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

func ptrToInt32Val(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

func ptrToInt32Ptr(i int32) *int32 {
	return &i
}

func marshalJSON(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
