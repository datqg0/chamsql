package domain

import (
	"database/sql"
	"time"
)

// ExcelExport represents an exported Excel file
type ExcelExport struct {
	ID         int64
	ExamID     int64
	ExportType string // results, analytics, submissions
	FilePath   string
	FileName   string
	CreatedBy  int64
	RowCount   sql.NullInt64
	CreatedAt  time.Time
}

// ExamResult represents exam results for export
type ExamResult struct {
	StudentID      int64
	StudentName    string
	StudentEmail   string
	TotalProblems  int
	PassedProblems int
	TotalScore     float64
	MaxScore       float64
	Percentage     float64
	SubmittedAt    time.Time
	Rank           int
}

// ProblemAnalytics represents analytics for a problem
type ProblemAnalytics struct {
	ProblemID      int64
	ProblemTitle   string
	TotalAttempts  int
	PassedAttempts int
	FailedAttempts int
	PassPercentage float64
	AverageScore   float64
	AverageTime    int // milliseconds
}

// SubmissionDetail represents a detailed submission
type SubmissionDetail struct {
	StudentID     int64
	StudentName   string
	ProblemID     int64
	ProblemTitle  string
	AttemptNumber int
	SubmittedCode string
	Status        string // accepted, wrong_answer, error
	Score         float64
	ExecutionTime int // milliseconds
	SubmittedAt   time.Time
}
