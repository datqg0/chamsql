package dto

import "encoding/json"

type RunQueryRequest struct {
	Code         string `json:"code" binding:"required"`
	DatabaseType string `json:"databaseType" binding:"required,oneof=postgresql mysql sqlserver"`
}

type SubmitQueryRequest struct {
	Code         string `json:"code" binding:"required"`
	DatabaseType string `json:"databaseType" binding:"required,oneof=postgresql mysql sqlserver"`
}

type RunQueryResponse struct {
	Success     bool            `json:"success"`
	Columns     []string        `json:"columns,omitempty"`
	Rows        [][]interface{} `json:"rows,omitempty"`
	RowCount    int             `json:"rowCount"`
	ExecutionMs int64           `json:"executionMs"`
	Error       string          `json:"error,omitempty"`
	ErrorType   string          `json:"errorType,omitempty"`
}

type SubmitQueryResponse struct {
	ID              int64              `json:"id"`
	IsCorrect       bool               `json:"isCorrect"`
	Status          string             `json:"status"` // accepted, wrong_answer, error, timeout
	ExecutionMs     int64              `json:"executionMs"`
	Score           float64            `json:"score"`
	TotalTests      int                `json:"totalTests"`
	PassedTests     int                `json:"passedTests"`
	Message         string             `json:"message,omitempty"`
	ExpectedOutput  json.RawMessage    `json:"expectedOutput,omitempty"`
	ActualOutput    json.RawMessage    `json:"actualOutput,omitempty"`
	Error           string             `json:"error,omitempty"`
	TestResults     []TestResultResponse `json:"testResults,omitempty"`
}

type TestResultResponse struct {
	TestCaseID   int64           `json:"testCaseId"`
	TestCaseName string          `json:"testCaseName"`
	Status       string          `json:"status"`
	ExecutionMs  int64           `json:"executionMs"`
	IsCorrect    bool            `json:"isCorrect"`
	IsHidden     bool            `json:"isHidden"`
	ActualOutput json.RawMessage `json:"actualOutput,omitempty"`
	ErrorMessage string          `json:"errorMessage,omitempty"`
}

type SubmissionResponse struct {
	ID              int64                `json:"id"`
	ProblemID       int64                `json:"problemId"`
	ProblemTitle    string               `json:"problemTitle,omitempty"`
	ProblemSlug     string               `json:"problemSlug,omitempty"`
	Code            string               `json:"code"`
	DatabaseType    string               `json:"databaseType"`
	Status          string               `json:"status"`
	IsCorrect       bool                 `json:"isCorrect"`
	Score           float64              `json:"score"`
	TotalTests      int                  `json:"totalTests"`
	PassedTests     int                  `json:"passedTests"`
	ExecutionTimeMs *int                 `json:"executionTimeMs,omitempty"`
	ExpectedOutput  json.RawMessage      `json:"expectedOutput,omitempty"`
	ActualOutput    json.RawMessage      `json:"actualOutput,omitempty"`
	ErrorMessage    string               `json:"errorMessage,omitempty"`
	SubmittedAt     string               `json:"submittedAt"`
	TestResults     []TestResultResponse `json:"testResults,omitempty"`
}

type SubmissionListResponse struct {
	Submissions []SubmissionResponse `json:"submissions"`
	Total       int64                `json:"total"`
	Page        int                  `json:"page"`
	PageSize    int                  `json:"pageSize"`
}
