package dto

// =============================================
// GRADING / SCORING DTOs
// =============================================

// GradeSubmissionRequest - Request to grade an exam submission
type GradeSubmissionRequest struct {
	SubmissionID  int64   `json:"submissionId" binding:"required"`
	Score         float64 `json:"score" binding:"required,min=0"`
	Feedback      string  `json:"feedback" binding:"omitempty,max=2000"`
	IsCorrect     *bool   `json:"isCorrect" binding:"omitempty"`
	ComparisonLog string  `json:"comparisonLog" binding:"omitempty,max=5000"`
}

// SubmissionGradingResponse - Response containing grading details
type SubmissionGradingResponse struct {
	SubmissionID    int64   `json:"submissionId"`
	StudentID       int64   `json:"studentId"`
	StudentName     string  `json:"studentName"`
	ProblemTitle    string  `json:"problemTitle"`
	Score           float64 `json:"score"`
	MaxPoints       float64 `json:"maxPoints"`
	IsCorrect       bool    `json:"isCorrect"`
	ScoringMode     string  `json:"scoringMode"`
	GradedBy        *int64  `json:"gradedBy,omitempty"`
	GradedByName    *string `json:"gradedByName,omitempty"`
	GradedAt        *string `json:"gradedAt,omitempty"`
	Feedback        string  `json:"feedback"`
	ComparisonLog   string  `json:"comparisonLog"`
	SubmittedAt     string  `json:"submittedAt"`
	StudentAnswer   *string `json:"studentAnswer,omitempty"`
	ReferenceAnswer *string `json:"referenceAnswer,omitempty"`
}

// ListUngradedSubmissionsResponse - List submissions needing grading
type ListUngradedSubmissionsResponse struct {
	Submissions   []SubmissionGradingResponse `json:"submissions"`
	Total         int64                       `json:"total"`
	ExamID        int64                       `json:"examId"`
	UngradedCount int64                       `json:"ungradedCount"`
	GradedCount   int64                       `json:"gradedCount"`
}

// ExamGradingStatsResponse - Statistics on exam grading progress
type ExamGradingStatsResponse struct {
	ExamID            int64   `json:"examId"`
	TotalSubmissions  int64   `json:"totalSubmissions"`
	GradedCount       int64   `json:"gradedCount"`
	UngradedCount     int64   `json:"ungradedCount"`
	GradingPercentage float64 `json:"gradingPercentage"`
	AverageScore      float64 `json:"averageScore"`
	MaxScore          float64 `json:"maxScore"`
	MinScore          float64 `json:"minScore"`
}

// ViewSubmissionResponse - Full submission details for grading
type ViewSubmissionResponse struct {
	SubmissionID    int64       `json:"submissionId"`
	ExamID          int64       `json:"examId"`
	ProblemID       int64       `json:"problemId"`
	ProblemTitle    string      `json:"problemTitle"`
	StudentID       int64       `json:"studentId"`
	StudentName     string      `json:"studentName"`
	StudentEmail    string      `json:"studentEmail"`
	Code            string      `json:"code"`
	Status          string      `json:"status"`
	ScoringMode     string      `json:"scoringMode"`
	Score           float64     `json:"score"`
	MaxPoints       float64     `json:"maxPoints"`
	IsCorrect       bool        `json:"isCorrect"`
	ActualOutput    interface{} `json:"actualOutput,omitempty"`
	ExpectedOutput  interface{} `json:"expectedOutput,omitempty"`
	ErrorMessage    *string     `json:"errorMessage,omitempty"`
	StudentAnswer   *string     `json:"studentAnswer,omitempty"`
	ReferenceAnswer *string     `json:"referenceAnswer,omitempty"`
	ExecutionTimeMs *int        `json:"executionTimeMs,omitempty"`
	AttemptNumber   int         `json:"attemptNumber"`
	SubmittedAt     string      `json:"submittedAt"`
	GradedAt        *string     `json:"gradedAt,omitempty"`
	GradedBy        *int64      `json:"gradedBy,omitempty"`
	GradedByName    *string     `json:"gradedByName,omitempty"`
	Feedback        string      `json:"feedback"`
}

// BulkGradeRequest - Request to grade multiple submissions
type BulkGradeRequest struct {
	Submissions []struct {
		SubmissionID int64   `json:"submissionId" binding:"required"`
		Score        float64 `json:"score" binding:"required,min=0"`
		Feedback     string  `json:"feedback" binding:"omitempty"`
	} `json:"submissions" binding:"required,min=1"`
}

// GradingErrorResponse represents an error from grading
type GradingErrorResponse struct {
	SubmissionID int64  `json:"submissionId"`
	Error        string `json:"error"`
}

// BulkGradeResponse - Response for bulk grading
type BulkGradeResponse struct {
	ProcessedCount int                         `json:"processedCount"`
	FailedCount    int                         `json:"failedCount"`
	Results        []SubmissionGradingResponse `json:"results"`
	Errors         []GradingErrorResponse      `json:"errors"`
}
