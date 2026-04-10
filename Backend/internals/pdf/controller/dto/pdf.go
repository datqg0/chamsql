package dto

import (
	"encoding/json"
	"time"
)

// PDFUploadResponse is the response for PDF upload
type PDFUploadResponse struct {
	ID        int64     `json:"id"`
	Status    string    `json:"status"`
	FileName  string    `json:"file_name"`
	CreatedAt time.Time `json:"created_at"`
	Message   string    `json:"message"`
}

// PDFUploadStatusResponse is the response for upload status
type PDFUploadStatusResponse struct {
	ID               int64           `json:"id"`
	Status           string          `json:"status"`
	FileName         string          `json:"file_name"`
	ExtractionResult json.RawMessage `json:"extraction_result,omitempty"`
	ErrorMessage     string          `json:"error_message,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// ProblemsResponse is the response for problems
type ProblemsResponse struct {
	UploadID int64             `json:"upload_id"`
	Problems []ProblemDraftDTO `json:"problems"`
}

// ProblemDraftDTO is a problem draft
type ProblemDraftDTO struct {
	ID          int64  `json:"id"`
	ProblemNum  int    `json:"problem_number"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Difficulty  string `json:"difficulty"`
	Solution    string `json:"solution,omitempty"`
	TestCases   int    `json:"test_case_count"`
	Status      string `json:"status"`
}

// ProblemReviewResponse is the response for problem review
type ProblemReviewResponse struct {
	ID            int64           `json:"id"`
	ProblemNumber int             `json:"problem_number"`
	Title         string          `json:"title"`
	Description   string          `json:"description"`
	Difficulty    string          `json:"difficulty"`
	SolutionQuery string          `json:"solution_query"`
	InitScript    string          `json:"init_script"`
	TestCases     json.RawMessage `json:"test_cases"`
	Status        string          `json:"status"`
	ReviewerID    *int64          `json:"reviewer_id,omitempty"`
	ReviewNotes   string          `json:"review_notes,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// ApproveReviewRequest is the request to approve a problem review
type ApproveReviewRequest struct {
	Status      string `json:"status" binding:"required,oneof=approved rejected editing"`
	ReviewNotes string `json:"review_notes"`
}

// ValidateTestCasesRequest is the request to validate test cases
type ValidateTestCasesRequest struct {
	ProblemID   int64  `json:"problem_id" binding:"required"`
	SolutionSQL string `json:"solution_sql" binding:"required"`
	SchemaSQL   string `json:"schema_sql" binding:"required"`
}

// ValidateTestCasesResponse is the response from test case validation
type ValidateTestCasesResponse struct {
	IsValid     bool     `json:"is_valid"`
	PassedCount int      `json:"passed_count"`
	TotalCount  int      `json:"total_count"`
	Errors      []string `json:"errors,omitempty"`
	ExecutionMS int      `json:"execution_time_ms"`
}

// ExcelExportRequest is the request to export results to Excel
type ExcelExportRequest struct {
	ExamID     int64  `json:"exam_id" binding:"required"`
	ExportType string `json:"export_type" binding:"required,oneof=results analytics submissions"`
}

// ExcelExportResponse is the response for Excel export
type ExcelExportResponse struct {
	ID        int64     `json:"id"`
	ExamID    int64     `json:"exam_id"`
	Type      string    `json:"type"`
	FilePath  string    `json:"file_path"`
	RowCount  int       `json:"row_count"`
	CreatedAt time.Time `json:"created_at"`
}
