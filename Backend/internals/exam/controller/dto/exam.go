package dto

import "time"

// ============ CREATE/UPDATE ============

type CreateExamRequest struct {
	Title                 string    `json:"title" binding:"required,min=3,max=255"`
	Description           string    `json:"description" binding:"omitempty,max=2000"`
	StartTime             time.Time `json:"startTime" binding:"required"`
	EndTime               time.Time `json:"endTime" binding:"required,gtfield=StartTime"`
	DurationMinutes       int       `json:"durationMinutes" binding:"required,min=5,max=480"`
	AllowedDatabases      []string  `json:"allowedDatabases" binding:"required,min=1"`
	AllowAiAssistance     bool      `json:"allowAiAssistance"`
	ShuffleProblems       bool      `json:"shuffleProblems"`
	ShowResultImmediately bool      `json:"showResultImmediately"`
	MaxAttempts           int       `json:"maxAttempts" binding:"omitempty,min=1,max=10"`
	IsPublic              bool      `json:"isPublic"`
}

type UpdateExamRequest struct {
	Title                 *string    `json:"title" binding:"omitempty,min=3,max=255"`
	Description           *string    `json:"description" binding:"omitempty,max=2000"`
	StartTime             *time.Time `json:"startTime" binding:"omitempty"`
	EndTime               *time.Time `json:"endTime" binding:"omitempty"`
	DurationMinutes       *int       `json:"durationMinutes" binding:"omitempty,min=5,max=480"`
	AllowAiAssistance     *bool      `json:"allowAiAssistance"`
	ShuffleProblems       *bool      `json:"shuffleProblems"`
	ShowResultImmediately *bool      `json:"showResultImmediately"`
	MaxAttempts           *int       `json:"maxAttempts" binding:"omitempty,min=1,max=10"`
	IsPublic              *bool      `json:"isPublic"`
}

// ============ EXAM PROBLEMS ============

type AddProblemRequest struct {
	ProblemID int64 `json:"problemId" binding:"required"`
	Points    int   `json:"points" binding:"required,min=1,max=100"`
	SortOrder int   `json:"sortOrder" binding:"omitempty,min=0"`
}

type ExamProblemResponse struct {
	ID          int64  `json:"id"`
	ProblemID   int64  `json:"problemId"`
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Difficulty  string `json:"difficulty"`
	Description string `json:"description,omitempty"`
	Points      int    `json:"points"`
	SortOrder   int    `json:"sortOrder"`
}

// ============ PARTICIPANTS ============

type AddParticipantsRequest struct {
	UserIDs []int64 `json:"userIds" binding:"required,min=1"`
}

type ParticipantResponse struct {
	ID          int64   `json:"id"`
	UserID      int64   `json:"userId"`
	FullName    string  `json:"fullName"`
	Email       string  `json:"email"`
	StudentID   string  `json:"studentId,omitempty"`
	Status      string  `json:"status"`
	StartedAt   *string `json:"startedAt,omitempty"`
	SubmittedAt *string `json:"submittedAt,omitempty"`
	TotalScore  float64 `json:"totalScore"`
}

// ============ RESPONSES ============

type ExamResponse struct {
	ID                    int64                 `json:"id"`
	Title                 string                `json:"title"`
	Description           string                `json:"description,omitempty"`
	CreatedBy             int64                 `json:"createdBy"`
	CreatorName           string                `json:"creatorName,omitempty"`
	StartTime             string                `json:"startTime"`
	EndTime               string                `json:"endTime"`
	DurationMinutes       int                   `json:"durationMinutes"`
	AllowedDatabases      []string              `json:"allowedDatabases"`
	AllowAiAssistance     bool                  `json:"allowAiAssistance"`
	ShuffleProblems       bool                  `json:"shuffleProblems"`
	ShowResultImmediately bool                  `json:"showResultImmediately"`
	MaxAttempts           int                   `json:"maxAttempts"`
	IsPublic              bool                  `json:"isPublic"`
	Status                string                `json:"status"`
	ProblemCount          int64                 `json:"problemCount,omitempty"`
	ParticipantCount      int64                 `json:"participantCount,omitempty"`
	Problems              []ExamProblemResponse `json:"problems,omitempty"`
	CreatedAt             string                `json:"createdAt"`
}

type ExamListResponse struct {
	Exams    []ExamResponse `json:"exams"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
}

// ============ STUDENT EXAM ============

type StartExamResponse struct {
	ExamID          int64                 `json:"examId"`
	Title           string                `json:"title"`
	DurationMinutes int                   `json:"durationMinutes"`
	StartedAt       string                `json:"startedAt"`
	EndsAt          string                `json:"endsAt"`
	Problems        []ExamProblemResponse `json:"problems"`
}

type ExamSubmitRequest struct {
	ProblemID    int64  `json:"problemId" binding:"required"`
	Code         string `json:"code" binding:"required"`
	DatabaseType string `json:"databaseType" binding:"required,oneof=postgresql mysql sqlserver"`
}

type ExamSubmitResponse struct {
	IsCorrect     bool    `json:"isCorrect"`
	Score         float64 `json:"score"`
	MaxScore      int     `json:"maxScore"`
	ExecutionMs   int64   `json:"executionMs"`
	Message       string  `json:"message,omitempty"`
	Error         string  `json:"error,omitempty"`
	AttemptNumber int     `json:"attemptNumber"`
	MaxAttempts   int     `json:"maxAttempts"`
}

type ExamResultResponse struct {
	ExamID       int64                 `json:"examId"`
	Title        string                `json:"title"`
	TotalScore   float64               `json:"totalScore"`
	MaxScore     int                   `json:"maxScore"`
	Status       string                `json:"status"`
	StartedAt    string                `json:"startedAt,omitempty"`
	SubmittedAt  string                `json:"submittedAt,omitempty"`
	Participants []ParticipantResponse `json:"participants,omitempty"`
}
