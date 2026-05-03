package dto

type ExamResult struct {
	ExamID      int64   `json:"exam_id"`
	Title       string  `json:"title"`
	StudentID   int64   `json:"student_id"`
	TotalScore  float64 `json:"total_score"`
	SubmittedAt string  `json:"submitted_at"`
	Status      string  `json:"status"`
}

type ExamResultDetail struct {
	ExamID      int64                  `json:"examId"`
	UserID      int64                  `json:"userId"`
	TotalScore  float64                `json:"totalScore"`
	Status      string                 `json:"status"`
	SubmittedAt string                 `json:"submittedAt"`
	Submissions []ExamSubmissionResult `json:"submissions"`
}

type ExamSubmissionResult struct {
	ProblemID    int64   `json:"problemId"`
	ProblemTitle string  `json:"problemTitle"`
	ProblemSlug  string  `json:"problemSlug"`
	Score        float64 `json:"score"`
	IsCorrect    bool    `json:"isCorrect"`
	ExecutionTimeMs *int32  `json:"executionTimeMs,omitempty"`
	SubmittedAt  string  `json:"submittedAt"`
}

type ProblemResultDetail struct {
	ExamProblemID int64   `json:"exam_problem_id"`
	ProblemID     int64   `json:"problem_id"`
	Title         string  `json:"title"`
	Difficulty    string  `json:"difficulty"`
	Points        *int32  `json:"points"`
	StudentScore  float64 `json:"student_score"`
	IsCorrect     bool    `json:"is_correct"`
	Status        string  `json:"status"`
	Attempts      int32   `json:"attempts"`
	GraderComment *string `json:"grader_comment,omitempty"`
}

type ListExamResultsResponse struct {
	Results []ExamResult `json:"results"`
	Total   int64        `json:"total"`
	Page    int32        `json:"page"`
	Limit   int32        `json:"limit"`
}

type ClassRankingResponse struct {
	ExamID    int64            `json:"exam_id"`
	ExamTitle string           `json:"exam_title"`
	Rankings  []StudentRanking `json:"rankings"`
	Total     int64            `json:"total"`
	Page      int32            `json:"page"`
	Limit     int32            `json:"limit"`
}

type StudentRanking struct {
	Rank        int32   `json:"rank"`
	StudentID   int64   `json:"student_id"`
	StudentCode string  `json:"student_code"`
	StudentName string  `json:"student_name"`
	Score       float64 `json:"score"`
	Percentile  float64 `json:"percentile"`
}

type ExamAnalytics struct {
	ExamID  int64  `json:"examId"`
	Message string `json:"message"`
}

type ProblemStat struct {
	ProblemID   int64   `json:"problem_id"`
	Title       string  `json:"title"`
	Points      *int32  `json:"points"`
	AvgScore    float64 `json:"avg_score"`
	CorrectRate float64 `json:"correct_rate"`
	AvgAttempts float64 `json:"avg_attempts"`
}

type ProblemDifficultyAnalytics struct {
	Difficulty   string  `json:"difficulty"`
	AvgScore     float64 `json:"avg_score"`
	CorrectRate  float64 `json:"correct_rate"`
	StudentCount int64   `json:"student_count"`
}

type ListExamResultsRequest struct {
	Page      int32   `form:"page" binding:"min=1"`
	Limit     int32   `form:"limit" binding:"min=1,max=100"`
	Status    string  `form:"status"`
	ScoreMin  float64 `form:"score_min"`
	ScoreMax  float64 `form:"score_max"`
	StartDate string  `form:"start_date"`
	EndDate   string  `form:"end_date"`
}

type RankingRequest struct {
	Page  int32 `form:"page" binding:"min=1"`
	Limit int32 `form:"limit" binding:"min=1,max=100"`
}

// SubmissionRecord - một lần nộp bài luyện tập
type SubmissionRecord struct {
	SubmissionID    int64   `json:"submissionId"`
	ProblemID       int64   `json:"problemId"`
	ProblemTitle    string  `json:"problemTitle"`
	ProblemSlug     string  `json:"problemSlug"`
	Code            string  `json:"code"`
	Status          string  `json:"status"`
	IsCorrect       *bool   `json:"isCorrect"`
	ExecutionTimeMs *int32  `json:"executionTimeMs"`
	ErrorMessage    *string `json:"errorMessage,omitempty"`
	SubmittedAt     string  `json:"submittedAt"`
}

// MySubmissionsResponse - lịch sử nộp bài luyện tập của sinh viên
type MySubmissionsResponse struct {
	Submissions []SubmissionRecord `json:"submissions"`
	Total       int64              `json:"total"`
	Page        int                `json:"page"`
	PageSize    int                `json:"pageSize"`
}

