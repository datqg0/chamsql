package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"backend/internals/problem/repository"
	"backend/internals/submission/domain"
	submissionRepo "backend/internals/submission/repository"
	"backend/pkgs/runner"
)

var (
	ErrGradingFailed = errors.New("grading failed")
)

// IGradingService defines grading operations
type IGradingService interface {
	Grade(ctx context.Context, req *domain.StudentSubmissionRequest) (*domain.GradingResult, error)
}

// gradingService handles automatic grading of student submissions
type gradingService struct {
	submissionRepo submissionRepo.ISubmissionRepository
	problemRepo    repository.IProblemRepository
	runner         runner.Runner
}

// NewGradingService creates a new grading service
func NewGradingService(
	subRepo submissionRepo.ISubmissionRepository,
	probRepo repository.IProblemRepository,
	queryRunner runner.Runner,
) IGradingService {
	return &gradingService{
		submissionRepo: subRepo,
		problemRepo:    probRepo,
		runner:         queryRunner,
	}
}

// Grade grades a student submission against test cases
func (s *gradingService) Grade(ctx context.Context, req *domain.StudentSubmissionRequest) (*domain.GradingResult, error) {
	if err := validateSubmissionRequest(req); err != nil {
		return nil, err
	}

	result := &domain.GradingResult{
		SubmissionID: req.SubmissionID,
		StudentID:    req.StudentID,
		ExamID:       req.ExamID,
		ProblemID:    req.ProblemID,
		TestResults:  make([]domain.TestResult, 0, len(req.TestCases)),
		Errors:       make([]string, 0),
		GradedAt:     time.Now().UTC(),
		MaxScore:     100,
	}

	if len(req.TestCases) == 0 {
		result.Status = "error"
		result.ErrorMessage = "no test cases provided"
		result.Errors = append(result.Errors, "no test cases provided")
		return result, nil
	}

	// Execute tests and collect results
	passedTests := 0
	dbType := runner.DBType("postgresql") // Default to PostgreSQL

	for _, testCase := range req.TestCases {
		testResult := s.executeTestCase(ctx, req.StudentSQL, testCase, dbType)
		result.TestResults = append(result.TestResults, testResult)

		if testResult.Passed {
			passedTests++
		}

		// Collect errors for summary
		if testResult.Error != "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Test %d: %s", testCase.TestNumber, testResult.Error))
		}
	}

	// Calculate score and status
	result.TotalTests = len(req.TestCases)
	result.PassedTests = passedTests
	result.Score = calculateScore(passedTests, len(req.TestCases))
	result.Status = determineStatus(passedTests, len(req.TestCases))

	return result, nil
}

// executeTestCase executes a single test case
func (s *gradingService) executeTestCase(
	ctx context.Context,
	studentSQL string,
	testCase domain.TestCaseData,
	dbType runner.DBType,
) domain.TestResult {
	startTime := time.Now()

	result := domain.TestResult{
		TestNumber:     testCase.TestNumber,
		Description:    testCase.Description,
		IsPublic:       testCase.IsPublic,
		ExpectedOutput: testCase.ExpectedOutput,
	}

	// Execute student query with test data setup
	actualResult, err := s.runner.ExecuteWithSetup(ctx, dbType, testCase.TestDataSQL, studentSQL)
	executionTime := time.Since(startTime).Milliseconds()
	result.ExecutionTime = executionTime

	if err != nil || actualResult.Error != "" {
		result.Passed = false
		result.Error = actualResult.Error
		result.ActualOutput = nil
		return result
	}

	// Compare outputs
	actualOutput := marshalJSON(actualResult.Rows)
	result.ActualOutput = actualOutput

	// Perform comparison (exact match for JSON)
	isCorrect := compareOutputs(testCase.ExpectedOutput, actualOutput)
	result.Passed = isCorrect

	return result
}

// compareOutputs compares expected and actual output
func compareOutputs(expected, actual json.RawMessage) bool {
	if len(expected) == 0 && len(actual) == 0 {
		return true
	}

	// Both should be valid JSON arrays
	var expectedRows []interface{}
	var actualRows []interface{}

	if err := json.Unmarshal(expected, &expectedRows); err != nil {
		return false
	}

	if err := json.Unmarshal(actual, &actualRows); err != nil {
		return false
	}

	if len(expectedRows) != len(actualRows) {
		return false
	}

	// Compare each row
	for i, exp := range expectedRows {
		act := actualRows[i]
		if !deepEqual(exp, act) {
			return false
		}
	}

	return true
}

// deepEqual performs deep equality check on JSON values
func deepEqual(a, b interface{}) bool {
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}

// calculateScore calculates the score based on passed tests
func calculateScore(passed, total int) int {
	if total == 0 {
		return 0
	}
	return (passed * 100) / total
}

// determineStatus determines the grading status
func determineStatus(passed, total int) string {
	if passed == total {
		return "passed"
	}
	if passed > 0 {
		return "partial"
	}
	return "failed"
}

// validateSubmissionRequest validates the submission request
func validateSubmissionRequest(req *domain.StudentSubmissionRequest) error {
	if req == nil {
		return errors.New("submission request is nil")
	}
	if req.StudentSQL == "" {
		return errors.New("student SQL is empty")
	}
	if req.SubmissionID == 0 {
		return errors.New("submission ID is invalid")
	}
	return nil
}
