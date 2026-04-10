package domain

import (
	"encoding/json"
	"time"
)

// StudentSubmissionRequest is published when student submits their exam answer
type StudentSubmissionRequest struct {
	SubmissionID   int64          `json:"submission_id"`
	StudentID      int64          `json:"student_id"`
	ExamID         int64          `json:"exam_id"`
	ProblemID      int64          `json:"problem_id"`
	StudentSQL     string         `json:"student_sql"`
	SchemaSQL      string         `json:"schema_sql"`
	SolutionSQL    string         `json:"solution_sql"`
	TestCases      []TestCaseData `json:"test_cases"`
	SubmittedAt    time.Time      `json:"submitted_at"`
	TimeoutSeconds int            `json:"timeout_seconds"`
}

// TestCaseData represents test case for grading
type TestCaseData struct {
	ID             int64           `json:"id"`
	TestNumber     int             `json:"test_number"`
	Description    string          `json:"description"`
	TestDataSQL    string          `json:"test_data_sql"`
	ExpectedOutput json.RawMessage `json:"expected_output"`
	IsPublic       bool            `json:"is_public"`
}

// GradingResult is the result of grading a submission
type GradingResult struct {
	SubmissionID  int64        `json:"submission_id"`
	StudentID     int64        `json:"student_id"`
	ExamID        int64        `json:"exam_id"`
	ProblemID     int64        `json:"problem_id"`
	Score         int          `json:"score"`     // 0-100
	MaxScore      int          `json:"max_score"` // Usually 100
	PassedTests   int          `json:"passed_tests"`
	TotalTests    int          `json:"total_tests"`
	ExecutionTime int64        `json:"execution_time"` // Milliseconds
	TestResults   []TestResult `json:"test_results"`
	Errors        []string     `json:"errors,omitempty"`
	Status        string       `json:"status"` // passed, partial, failed
	GradedAt      time.Time    `json:"graded_at"`
	ErrorMessage  string       `json:"error_message,omitempty"`
}

// TestResult represents individual test case result
type TestResult struct {
	TestNumber     int             `json:"test_number"`
	Description    string          `json:"description"`
	Passed         bool            `json:"passed"`
	ExpectedOutput json.RawMessage `json:"expected_output"`
	ActualOutput   json.RawMessage `json:"actual_output"`
	Error          string          `json:"error,omitempty"`
	ExecutionTime  int64           `json:"execution_time"` // Milliseconds
	IsPublic       bool            `json:"is_public"`
}
