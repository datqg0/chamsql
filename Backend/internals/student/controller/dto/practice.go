package dto

// =============================================
// PUBLIC PROBLEMS DTOs
// =============================================

// ListPublicProblemsRequest - List public problems for practice
type ListPublicProblemsRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PageSize   int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	Difficulty string `form:"difficulty" binding:"omitempty,oneof=easy medium hard"`
	Topic      string `form:"topic" binding:"omitempty"`
}

// PublicProblemBrief - Brief info about public problem
type PublicProblemBrief struct {
	ProblemID   int64   `json:"problem_id"`
	Title       string  `json:"title"`
	Slug        string  `json:"slug"`
	Description string  `json:"description"`
	Difficulty  string  `json:"difficulty"`
	TopicID     *int32  `json:"topic_id,omitempty"`
	TopicName   string  `json:"topic_name,omitempty"`
	TopicSlug   string  `json:"topic_slug,omitempty"`
	Hints       *string `json:"hints,omitempty"`
	CreatedBy   int64   `json:"created_by"`
	CreatedAt   string  `json:"created_at"`
}

// ListPublicProblemsResponse - Paginated list of public problems
type ListPublicProblemsResponse struct {
	Problems []PublicProblemBrief `json:"problems"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"pageSize"`
}

// GetPublicProblemResponse - Full details of public problem for practice
type GetPublicProblemResponse struct {
	ProblemID          int64                `json:"problem_id"`
	Title              string               `json:"title"`
	Slug               string               `json:"slug"`
	Description        string               `json:"description"`
	Difficulty         string               `json:"difficulty"`
	TopicID            *int32               `json:"topic_id,omitempty"`
	TopicName          string               `json:"topic_name,omitempty"`
	TopicSlug          string               `json:"topic_slug,omitempty"`
	InitScript         *string              `json:"init_script,omitempty"`
	SolutionQuery      *string              `json:"solution_query,omitempty"`
	SampleOutput       *string              `json:"sample_output,omitempty"`
	Hints              *string              `json:"hints,omitempty"`
	OrderMatters       *bool                `json:"order_matters,omitempty"`
	SupportedDatabases []string             `json:"supported_databases"`
	PracticeStats      *PracticeStats       `json:"practice_stats,omitempty"` // User's practice stats
	LatestSubmissions  []PracticeSubmission `json:"latest_submissions,omitempty"`
}

// PracticeStats - User's practice statistics for a problem
type PracticeStats struct {
	TotalAttempts   int64   `json:"total_attempts"`
	CorrectAttempts int64   `json:"correct_attempts"`
	IsSolved        bool    `json:"is_solved"`
	BestTimeMs      *int32  `json:"best_time_ms,omitempty"`
	LastSubmittedAt *string `json:"last_submitted_at,omitempty"`
}

// =============================================
// PRACTICE SUBMISSION DTOs
// =============================================

// PracticeSubmitCodeRequest - Submit code for practice problem
type PracticeSubmitCodeRequest struct {
	Code         string `json:"code" binding:"required"`
	DatabaseType string `json:"database_type" binding:"omitempty"`
}

// PracticeSubmitCodeResponse - Result of practice code submission
type PracticeSubmitCodeResponse struct {
	SubmissionID    int64   `json:"submission_id"`
	ProblemID       int64   `json:"problem_id"`
	Status          string  `json:"status"` // accepted, wrong_answer, error, timeout
	IsCorrect       bool    `json:"is_correct"`
	ExecutionTimeMs *int32  `json:"execution_time_ms,omitempty"`
	ErrorMessage    *string `json:"error_message,omitempty"`
	ActualOutput    *string `json:"actual_output,omitempty"`
	ExpectedOutput  *string `json:"expected_output,omitempty"`
	SubmittedAt     string  `json:"submitted_at"`
	AttemptNumber   int32   `json:"attempt_number"`
	TotalAttempts   int64   `json:"total_attempts"`   // Total attempts for this problem
	CorrectAttempts int64   `json:"correct_attempts"` // Total correct attempts
}

// PracticeSubmission - Single practice submission record
type PracticeSubmission struct {
	SubmissionID    int64   `json:"submission_id"`
	Code            string  `json:"code"`
	Status          string  `json:"status"`
	IsCorrect       bool    `json:"is_correct"`
	AttemptNumber   int32   `json:"attempt_number"`
	ExecutionTimeMs *int32  `json:"execution_time_ms,omitempty"`
	ErrorMessage    *string `json:"error_message,omitempty"`
	SubmittedAt     string  `json:"submitted_at"`
}

// ListPracticeSubmissionsResponse - List practice submissions for a problem
type ListPracticeSubmissionsResponse struct {
	Submissions []PracticeSubmission `json:"submissions"`
	Total       int64                `json:"total"`
	Page        int                  `json:"page"`
	PageSize    int                  `json:"pageSize"`
}
