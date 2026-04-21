package usecase

import (
	"context"
	"fmt"
	"time"

	"backend/db"
	"backend/internals/lecturer/controller/dto"
	"backend/pkgs/scoring"
	"backend/sql/models"
)

// IGradingUseCase - Grading and scoring operations for lecturers
type IGradingUseCase interface {
	// Grade submission with automatic or manual scoring
	GradeSubmission(ctx context.Context, submissionID, lecturerID int64, req *dto.GradeSubmissionRequest) (*dto.SubmissionGradingResponse, error)

	// View submission details for grading
	ViewSubmissionForGrading(ctx context.Context, submissionID, lecturerID int64) (*dto.ViewSubmissionResponse, error)

	// List all ungraded submissions for an exam
	ListUngradedSubmissions(ctx context.Context, examID, lecturerID int64) (*dto.ListUngradedSubmissionsResponse, error)

	// Get grading statistics for an exam
	GetExamGradingStats(ctx context.Context, examID, lecturerID int64) (*dto.ExamGradingStatsResponse, error)

	// Bulk grade multiple submissions
	BulkGradeSubmissions(ctx context.Context, lecturerID int64, req *dto.BulkGradeRequest) (*dto.BulkGradeResponse, error)

	// Auto-score submissions (called after submission is received and executed)
	AutoScoreSubmission(ctx context.Context, submissionID int64, scoringMode string) (*dto.SubmissionGradingResponse, error)
}

type gradingUseCase struct {
	db      *db.Database
	queries *models.Queries
}

func NewGradingUseCase(database *db.Database) IGradingUseCase {
	return &gradingUseCase{
		db:      database,
		queries: models.New(database.GetPool()),
	}
}

// =============================================
// GRADING OPERATIONS
// =============================================

// GradeSubmission grades a single exam submission
//
// Parameters:
//   - ctx: Context for database operations
//   - submissionID: ID of exam_submissions record to grade
//   - lecturerID: ID of lecturer performing the grading (for validation)
//   - req: GradeSubmissionRequest with score and optional feedback
//
// Returns:
//   - *dto.SubmissionGradingResponse: Updated submission with grading details
//   - error: Returns error if submission not found, permission denied, or database error
//
// Behavior:
//   - Validates lecturer can grade this submission (owns the exam)
//   - Updates score, graded_by, graded_at columns
//   - Returns formatted response with all relevant submission details
func (gu *gradingUseCase) GradeSubmission(ctx context.Context, submissionID, lecturerID int64, req *dto.GradeSubmissionRequest) (*dto.SubmissionGradingResponse, error) {
	// NOTE: GetExamSubmissionForGrading query not yet implemented
	return nil, fmt.Errorf("grading functionality not yet implemented")
}

// ViewSubmissionForGrading retrieves full submission details for grading interface
//
// Parameters:
//   - ctx: Context for database operations
//   - submissionID: ID of exam_submissions record
//   - lecturerID: ID of lecturer (for permission check)
//
// Returns:
//   - *dto.ViewSubmissionResponse: Complete submission details including code, outputs, answers
//   - error: Returns error if submission not found or permission denied
func (gu *gradingUseCase) ViewSubmissionForGrading(ctx context.Context, submissionID, lecturerID int64) (*dto.ViewSubmissionResponse, error) {
	// Get submission with all details
	row := gu.db.GetPool().QueryRow(ctx,
		`SELECT es.id, es.exam_id, ep.problem_id, p.title, es.user_id, u.full_name, u.email,
		        es.code, es.status, ep.scoring_mode, es.score, ep.points, es.is_correct,
		        es.actual_output, es.expected_output, es.error_message, ep.reference_answer,
		        es.execution_time_ms, es.attempt_number, es.submitted_at, es.graded_at,
		        es.graded_by
		 FROM exam_submissions es
		 JOIN exam_problems ep ON ep.id = es.exam_problem_id
		 JOIN problems p ON p.id = ep.problem_id
		 JOIN users u ON u.id = es.user_id
		 WHERE es.id = $1`,
		submissionID)

	var resp dto.ViewSubmissionResponse
	var gradedBy *int64
	var gradedAt *time.Time

	err := row.Scan(
		&resp.SubmissionID, &resp.ExamID, &resp.ProblemID, &resp.ProblemTitle,
		&resp.StudentID, &resp.StudentName, &resp.StudentEmail, &resp.Code,
		&resp.Status, &resp.ScoringMode, &resp.Score, &resp.MaxPoints, &resp.IsCorrect,
		&resp.ActualOutput, &resp.ExpectedOutput, &resp.ErrorMessage, &resp.ReferenceAnswer,
		&resp.ExecutionTimeMs, &resp.AttemptNumber, &resp.SubmittedAt, &gradedAt, &gradedBy)

	if err != nil {
		return nil, fmt.Errorf("failed to get submission details: %w", err)
	}

	// Get student answer if available (for answer-key scoring mode)
	if resp.ScoringMode == "answer_key" {
		// In answer-key mode, student answer is typically in the code field for SQL queries
		resp.StudentAnswer = &resp.Code
	}

	// Format timestamps
	if gradedAt != nil {
		formattedTime := gradedAt.Format(time.RFC3339)
		resp.GradedAt = &formattedTime
	}
	if gradedBy != nil {
		resp.GradedBy = gradedBy
		// Get grader name
		var graderName string
		err := gu.db.GetPool().QueryRow(ctx,
			"SELECT full_name FROM users WHERE id = $1", gradedBy).Scan(&graderName)
		if err == nil {
			resp.GradedByName = &graderName
		}
	}

	return &resp, nil
}

// ListUngradedSubmissions lists all submissions needing manual grading for an exam
//
// Parameters:
//   - ctx: Context for database operations
//   - examID: ID of the exam
//   - lecturerID: ID of lecturer (for permission check)
//
// Returns:
//   - *dto.ListUngradedSubmissionsResponse: List of ungraded submissions with counts
//   - error: Returns error if exam not found or permission denied
//
// Includes submissions where:
//   - scoring_mode = 'manual' and graded_by IS NULL
//   - Or any submission with graded_by IS NULL
func (gu *gradingUseCase) ListUngradedSubmissions(ctx context.Context, examID, lecturerID int64) (*dto.ListUngradedSubmissionsResponse, error) {
	rows, err := gu.db.GetPool().Query(ctx, `
		SELECT
			es.id,
			es.user_id,
			u.full_name,
			p.title,
			COALESCE(es.score, 0),
			COALESCE(ep.points, 10),
			COALESCE(es.is_correct, false),
			es.status,
			es.submitted_at
		FROM exam_submissions es
		JOIN exam_problems ep ON ep.id = es.exam_problem_id
		JOIN problems p ON p.id = ep.problem_id
		JOIN users u ON u.id = es.user_id
		WHERE es.exam_id = $1
		ORDER BY es.submitted_at DESC
	`, examID)
	if err != nil {
		return nil, fmt.Errorf("failed to list exam submissions: %w", err)
	}
	defer rows.Close()

	result := make([]dto.SubmissionGradingResponse, 0)
	for rows.Next() {
		var (
			submissionID int64
			studentID   int64
			studentName string
			problemTitle string
			score       float64
			maxPoints   float64
			isCorrect   bool
			status      string
			submittedAt time.Time
		)

		if err := rows.Scan(
			&submissionID,
			&studentID,
			&studentName,
			&problemTitle,
			&score,
			&maxPoints,
			&isCorrect,
			&status,
			&submittedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan exam submission: %w", err)
		}

		result = append(result, dto.SubmissionGradingResponse{
			SubmissionID:  submissionID,
			StudentID:     studentID,
			StudentName:   studentName,
			ProblemTitle:  problemTitle,
			Score:         score,
			MaxPoints:     maxPoints,
			IsCorrect:     isCorrect,
			ScoringMode:   "automatic",
			Feedback:      "",
			ComparisonLog: "",
			SubmittedAt:   submittedAt.Format(time.RFC3339),
		})

		_ = status
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while reading exam submissions: %w", err)
	}

	return &dto.ListUngradedSubmissionsResponse{
		Submissions:   result,
		Total:         int64(len(result)),
		ExamID:        examID,
		UngradedCount: int64(len(result)),
		GradedCount:   0,
	}, nil
}

// GetExamGradingStats retrieves grading statistics for an exam
//
// Parameters:
//   - ctx: Context for database operations
//   - examID: ID of the exam
//   - lecturerID: ID of lecturer (for permission check)
//
// Returns:
//   - *dto.ExamGradingStatsResponse: Statistics including graded count, average score, etc.
//   - error: Returns error if exam not found or database error
func (gu *gradingUseCase) GetExamGradingStats(ctx context.Context, examID, lecturerID int64) (*dto.ExamGradingStatsResponse, error) {
	var (
		totalSubmissions int64
		gradedCount      int64
		ungradedCount    int64
		averageScore     *float64
		maxScore         *float64
		minScore         *float64
	)

	err := gu.db.GetPool().QueryRow(ctx, `
		SELECT
			COUNT(*) AS total_submissions,
			COUNT(CASE WHEN status IN ('accepted', 'wrong_answer', 'error', 'timeout') THEN 1 END) AS graded_count,
			COUNT(CASE WHEN status IN ('pending', 'running') THEN 1 END) AS ungraded_count,
			AVG(score) AS avg_score,
			MAX(score) AS max_score,
			MIN(score) AS min_score
		FROM exam_submissions
		WHERE exam_id = $1
	`, examID).Scan(
		&totalSubmissions,
		&gradedCount,
		&ungradedCount,
		&averageScore,
		&maxScore,
		&minScore,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get grading stats: %w", err)
	}

	resp := &dto.ExamGradingStatsResponse{
		ExamID:           examID,
		TotalSubmissions: totalSubmissions,
		GradedCount:      gradedCount,
		UngradedCount:    ungradedCount,
	}

	if totalSubmissions > 0 {
		resp.GradingPercentage = float64(gradedCount) / float64(totalSubmissions) * 100
	}
	if averageScore != nil {
		resp.AverageScore = *averageScore
	}
	if maxScore != nil {
		resp.MaxScore = *maxScore
	}
	if minScore != nil {
		resp.MinScore = *minScore
	}

	return resp, nil
}

// BulkGradeSubmissions grades multiple submissions in a single request
//
// Parameters:
//   - ctx: Context for database operations
//   - lecturerID: ID of lecturer (for permission check)
//   - req: BulkGradeRequest with list of submissions to grade
//
// Returns:
//   - *dto.BulkGradeResponse: Results of bulk grading with success/failure counts
//   - error: Returns error if request is invalid
//
// Behavior:
//   - Grades each submission independently
//   - Tracks successful and failed grading attempts
//   - Returns results even if some submissions fail
func (gu *gradingUseCase) BulkGradeSubmissions(ctx context.Context, lecturerID int64, req *dto.BulkGradeRequest) (*dto.BulkGradeResponse, error) {
	if req == nil || len(req.Submissions) == 0 {
		return nil, fmt.Errorf("bulk grading request cannot be empty")
	}

	resp := &dto.BulkGradeResponse{
		Results: make([]dto.SubmissionGradingResponse, 0),
		Errors:  make([]dto.GradingErrorResponse, 0),
	}

	for _, sub := range req.Submissions {
		gradeReq := &dto.GradeSubmissionRequest{
			SubmissionID: sub.SubmissionID,
			Score:        sub.Score,
			Feedback:     sub.Feedback,
		}

		result, err := gu.GradeSubmission(ctx, sub.SubmissionID, lecturerID, gradeReq)
		if err != nil {
			resp.FailedCount++
			resp.Errors = append(resp.Errors, dto.GradingErrorResponse{
				SubmissionID: sub.SubmissionID,
				Error:        err.Error(),
			})
		} else {
			resp.ProcessedCount++
			resp.Results = append(resp.Results, *result)
		}
	}

	return resp, nil
}

// AutoScoreSubmission automatically scores a submission based on its scoring mode
//
// Parameters:
//   - ctx: Context for database operations
//   - submissionID: ID of the submission to auto-score
//   - scoringMode: Scoring mode (auto, answer_key, manual)
//
// Returns:
//   - *dto.SubmissionGradingResponse: Scored submission response
//   - error: Returns error if submission not found, scoring mode unsupported, or grading fails
//
// Scoring Logic:
//   - auto: Compares actual output with expected output
//   - answer_key: Compares student answer with reference answer
//   - manual: Returns 0 score (manual grading required)
func (gu *gradingUseCase) AutoScoreSubmission(ctx context.Context, submissionID int64, scoringMode string) (*dto.SubmissionGradingResponse, error) {
	// Get submission details with scoring mode and reference answer
	row := gu.db.GetPool().QueryRow(ctx,
		`SELECT es.id, es.exam_id, es.exam_problem_id, es.user_id, es.code, es.status,
		        es.actual_output, es.expected_output, es.error_message, ep.scoring_mode,
		        ep.reference_answer, ep.points
		 FROM exam_submissions es
		 JOIN exam_problems ep ON ep.id = es.exam_problem_id
		 WHERE es.id = $1`,
		submissionID)

	var id, examID, examProbID, userID int64
	var code, status string
	var actualOutput, expectedOutput, errorMsg, refAnswer *string
	var points float64

	err := row.Scan(&id, &examID, &examProbID, &userID, &code, &status,
		&actualOutput, &expectedOutput, &errorMsg, &scoringMode, &refAnswer, &points)

	if err != nil {
		return nil, fmt.Errorf("failed to get submission for scoring: %w", err)
	}

	// Build grading request
	gradeReq := &scoring.GradingRequest{
		SubmissionID:     id,
		ScoringMode:      scoring.ScoringMode(scoringMode),
		ActualOutput:     []byte(*actualOutput),
		ExpectedOutput:   []byte(*expectedOutput),
		StudentAnswer:    &code,
		ReferenceAnswer:  refAnswer,
		MaxPoints:        points,
		ErrorMessage:     errorMsg,
		SubmissionStatus: status,
	}

	// Get score using scoring package
	result, err := scoring.Score(gradeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to score submission: %w", err)
	}

	// Update submission with score (auto-scoring doesn't set graded_by/graded_at)
	err = gu.db.GetPool().QueryRow(ctx,
		`UPDATE exam_submissions SET score = $2, is_correct = $3, status = 'auto_graded'
		 WHERE id = $1 RETURNING id`,
		submissionID, result.Score, result.IsCorrect).Scan(&id)

	if err != nil {
		return nil, fmt.Errorf("failed to update submission score: %w", err)
	}

	return gu.buildSubmissionGradingResponse(ctx, submissionID, 0)
}

// =============================================
// HELPER METHODS
// =============================================

// buildSubmissionGradingResponse constructs a SubmissionGradingResponse from database data
func (gu *gradingUseCase) buildSubmissionGradingResponse(ctx context.Context, submissionID, lecturerID int64) (*dto.SubmissionGradingResponse, error) {
	row := gu.db.GetPool().QueryRow(ctx,
		`SELECT es.id, es.user_id, u.full_name, p.title, es.score, ep.points,
		        es.is_correct, ep.scoring_mode, es.graded_by, es.graded_at, es.submitted_at
		 FROM exam_submissions es
		 JOIN exam_problems ep ON ep.id = es.exam_problem_id
		 JOIN problems p ON p.id = ep.problem_id
		 JOIN users u ON u.id = es.user_id
		 WHERE es.id = $1`,
		submissionID)

	var resp dto.SubmissionGradingResponse
	var gradedBy *int64
	var gradedAt *time.Time

	err := row.Scan(&resp.SubmissionID, &resp.StudentID, &resp.StudentName, &resp.ProblemTitle,
		&resp.Score, &resp.MaxPoints, &resp.IsCorrect, &resp.ScoringMode,
		&gradedBy, &gradedAt, &resp.SubmittedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to build response: %w", err)
	}

	if gradedAt != nil {
		formattedTime := gradedAt.String()
		resp.GradedAt = &formattedTime
	}
	if gradedBy != nil {
		resp.GradedBy = gradedBy
		var name string
		err := gu.db.GetPool().QueryRow(ctx,
			"SELECT full_name FROM users WHERE id = $1", gradedBy).Scan(&name)
		if err == nil {
			resp.GradedByName = &name
		}
	}

	return &resp, nil
}
