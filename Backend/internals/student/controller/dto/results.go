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
	ExamID       int64                 `json:"exam_id"`
	Title        string                `json:"title"`
	Description  string                `json:"description"`
	TotalScore   float64               `json:"total_score"`
	DurationMins int32                 `json:"duration_minutes"`
	StartTime    string                `json:"start_time"`
	EndTime      string                `json:"end_time"`
	SubmittedAt  string                `json:"submitted_at"`
	Status       string                `json:"status"`
	Problems     []ProblemResultDetail `json:"problems"`
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
}

type ClassRankingResponse struct {
	ExamID    int64            `json:"exam_id"`
	ExamTitle string           `json:"exam_title"`
	Rankings  []StudentRanking `json:"rankings"`
}

type StudentRanking struct {
	Rank        int32   `json:"rank"`
	StudentID   int64   `json:"student_id"`
	StudentName string  `json:"student_name"`
	Score       float64 `json:"score"`
	Percentile  float64 `json:"percentile"`
}

type ExamAnalytics struct {
	ExamID        int64         `json:"exam_id"`
	Title         string        `json:"title"`
	TotalStudents int64         `json:"total_students"`
	AvgScore      float64       `json:"avg_score"`
	HighestScore  float64       `json:"highest_score"`
	LowestScore   float64       `json:"lowest_score"`
	PassRate      float64       `json:"pass_rate"`
	ProblemsStats []ProblemStat `json:"problems_stats"`
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
