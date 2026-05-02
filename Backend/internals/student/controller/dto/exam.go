package dto

type JoinExamRequest struct {
	ExamID int64 `json:"exam_id" binding:"required"`
}

type JoinExamResponse struct {
	ParticipantID int64  `json:"participant_id"`
	ExamID        int64  `json:"exam_id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	DurationMins  int32  `json:"duration_minutes"`
	TotalProblems int64  `json:"total_problems"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

type StartExamRequest struct {
	ExamID int64 `json:"exam_id" binding:"required"`
}

type StartExamResponse struct {
	ParticipantID   int64  `json:"participant_id"`
	ExamID          int64  `json:"exam_id"`
	StartedAt       string `json:"started_at"`
	TimeRemainingMs int64  `json:"time_remaining_ms"`
	Status          string `json:"status"`
}

type GetExamResponse struct {
	ExamID            int64              `json:"exam_id"`
	Title             string             `json:"title"`
	Description       string             `json:"description"`
	StartTime         string             `json:"start_time"`
	EndTime           string             `json:"end_time"`
	DurationMins      int32              `json:"duration_minutes"`
	Status            string             `json:"status"`
	TimeRemainingMs   int64              `json:"time_remaining_ms"`
	ParticipantStatus string             `json:"participant_status"`
	Problems          []ExamProblemBrief `json:"problems"`
}

type ExamProblemBrief struct {
	ExamProblemID int64  `json:"exam_problem_id"`
	ProblemID     int64  `json:"problem_id"`
	Title         string `json:"title"`
	Difficulty    string `json:"difficulty"`
	Points        *int32 `json:"points"`
	SortOrder     *int32 `json:"sort_order"`
}

type GetProblemResponse struct {
	ExamProblemID   int64               `json:"exam_problem_id"`
	ProblemID       int64               `json:"problem_id"`
	Title           string              `json:"title"`
	Description     string              `json:"description"`
	Difficulty      string              `json:"difficulty"`
	Points          *int32              `json:"points"`
	SortOrder       *int32              `json:"sort_order"`
	ScoringMode     *string             `json:"scoring_mode"`
	ReferenceAnswer *string             `json:"reference_answer,omitempty"`
	InitScript      *string             `json:"init_script,omitempty"`
	AttemptNumber   int32               `json:"attempt_number"`
	Submissions     []StudentSubmission `json:"submissions"`
}

type StudentSubmission struct {
	SubmissionID    int64   `json:"submission_id"`
	Code            string  `json:"code"`
	Status          string  `json:"status"`
	Score           float64 `json:"score,omitempty"`
	IsCorrect       bool    `json:"is_correct"`
	AttemptNumber   int32   `json:"attempt_number"`
	ExecutionTimeMs *int32  `json:"execution_time_ms,omitempty"`
	ErrorMessage    *string `json:"error_message,omitempty"`
	SubmittedAt     string  `json:"submitted_at"`
}

type SubmitCodeRequest struct {
	Code         string `json:"code" binding:"required"`
	DatabaseType string `json:"database_type"`
}

type SubmitCodeResponse struct {
	SubmissionID    int64   `json:"submission_id"`
	ExamID          int64   `json:"exam_id"`
	ExamProblemID   int64   `json:"exam_problem_id"`
	Status          string  `json:"status"`
	Score           float64 `json:"score"`
	IsCorrect       bool    `json:"is_correct"`
	AttemptNumber   int32   `json:"attempt_number"`
	ExecutionTimeMs *int32  `json:"execution_time_ms,omitempty"`
	ErrorMessage    *string `json:"error_message,omitempty"`
	SubmittedAt     string  `json:"submitted_at"`
	ScoringMode     string  `json:"scoring_mode"`
}

type SubmitExamRequest struct {
	ExamID int64 `json:"exam_id" binding:"required"`
}

type SubmitExamResponse struct {
	ParticipantID int64   `json:"participant_id"`
	ExamID        int64   `json:"exam_id"`
	TotalScore    float64 `json:"total_score"`
	SubmittedAt   string  `json:"submitted_at"`
	Status        string  `json:"status"`
}

type GetTimeRemainingResponse struct {
	TimeRemainingMs int64  `json:"time_remaining_ms"`
	ExamID          int64  `json:"exam_id"`
	Status          string `json:"status"`
	Message         string `json:"message,omitempty"`
}
