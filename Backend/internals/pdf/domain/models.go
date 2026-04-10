package domain

import (
	"database/sql"
	"encoding/json"
	"time"
)

// PDFUpload represents a PDF file upload
type PDFUpload struct {
	ID               int64
	LecturerID       int64
	FilePath         string // MinIO path
	FileName         string
	OriginalFilename string
	Status           string // uploading, parsing, generating, completed, failed
	ExtractionResult json.RawMessage
	ErrorMessage     sql.NullString
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ExtractedProblem represents a problem extracted from PDF
type ExtractedProblem struct {
	ProblemNumber int
	Title         string
	Description   string
	Difficulty    string // easy, medium, hard
	SchemaSQL     string
	SolutionSQL   sql.NullString // May be empty if not in PDF
	TestCases     []TestCaseData
	SampleOutput  json.RawMessage
}

// TestCaseData represents a test case
type TestCaseData struct {
	TestNumber     int
	Description    string
	TestDataSQL    string
	ExpectedOutput json.RawMessage
	IsPublic       bool
}

// AIGeneratedContent stores AI-generated content
type AIGeneratedContent struct {
	ID                 int64
	PDFUploadID        int64
	ProblemNumber      int
	ContentType        string // solution, test_case, description
	OriginalContent    string
	AIGeneratedContent string
	ConfidenceScore    sql.NullFloat64
	AIProvider         string // pattern, huggingface
	IsApproved         bool
	LecturerNotes      sql.NullString
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// ProblemReviewQueue tracks pending problem reviews
type ProblemReviewQueue struct {
	ID            int64
	PDFUploadID   int64
	ProblemNumber int
	ProblemDraft  json.RawMessage
	Status        string // pending, approved, rejected, editing
	ReviewerID    sql.NullInt64
	ReviewNotes   sql.NullString
	EditsMade     json.RawMessage
	ReviewedAt    sql.NullTime
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ProblemDraft is the draft data for review
type ProblemDraft struct {
	Title         string          `json:"title"`
	Description   string          `json:"description"`
	Difficulty    string          `json:"difficulty"`
	TopicID       int             `json:"topic_id"`
	SolutionQuery string          `json:"solution_query"`
	InitScript    string          `json:"init_script"`
	TestCases     []TestCaseData  `json:"test_cases"`
	SampleOutput  json.RawMessage `json:"sample_output"`
	Hints         json.RawMessage `json:"hints"`
}

// ValidationResult represents test validation result
type ValidationResult struct {
	IsValid         bool     `json:"is_valid"`
	PassedCount     int      `json:"passed_count"`
	TotalCount      int      `json:"total_count"`
	Errors          []string `json:"errors,omitempty"`
	ExecutionTimeMS int      `json:"execution_time_ms"`
}
