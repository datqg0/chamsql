package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"time"

	"backend/db"
	"backend/internals/student/controller/dto"
	"backend/pkgs/redis"
	"backend/pkgs/runner"
	"backend/sql/models"

	"github.com/jackc/pgx/v5/pgtype"
)

type IStudentExamUseCase interface {
	JoinExam(ctx context.Context, examID, userID int64) (*dto.JoinExamResponse, error)
	StartExam(ctx context.Context, examID, userID int64) (*dto.StartExamResponse, error)
	GetExam(ctx context.Context, examID, userID int64) (*dto.GetExamResponse, error)
	GetProblem(ctx context.Context, examID, examProblemID, userID int64) (*dto.GetProblemResponse, error)
	SubmitCode(ctx context.Context, examID, examProblemID, userID int64, req *dto.SubmitCodeRequest) (*dto.SubmitCodeResponse, error)
	SubmitExam(ctx context.Context, examID, userID int64) (*dto.SubmitExamResponse, error)
	GetTimeRemaining(ctx context.Context, examID, userID int64) (*dto.GetTimeRemainingResponse, error)
}

type studentExamUseCase struct {
	db       *db.Database
	queries  *models.Queries
	executor CodeExecutor
	cache    redis.IRedis
}

func numericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid || n.Int == nil {
		return 0
	}
	f, _ := new(big.Float).SetInt(n.Int).Float64()
	if n.Exp != 0 {
		f *= math.Pow(10, float64(n.Exp))
	}
	return f
}

func NewStudentExamUseCase(database *db.Database, cache redis.IRedis, queryRunner runner.Runner) IStudentExamUseCase {
	return &studentExamUseCase{
		db:       database,
		queries:  models.New(database.GetPool()),
		executor: NewCodeExecutor(queryRunner),
		cache:    cache,
	}
}

func (su *studentExamUseCase) JoinExam(ctx context.Context, examID, userID int64) (*dto.JoinExamResponse, error) {
	exam, err := su.queries.GetExamForStudent(ctx, examID)
	if err != nil {
		return nil, fmt.Errorf("exam not found or not published: %w", err)
	}

	_, err = su.queries.GetParticipantStatus(ctx, models.GetParticipantStatusParams{
		ExamID: examID,
		UserID: userID,
	})
	if err == nil {
		return nil, fmt.Errorf("already joined this exam")
	}

	newParticipant, err := su.queries.AddParticipant(ctx, models.AddParticipantParams{
		ExamID: examID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to join exam: %w", err)
	}

	// Get problem count for this exam
	problemCount := int64(0)
	problems, err := su.queries.GetExamProblemsForStudent(ctx, examID)
	if err == nil {
		problemCount = int64(len(problems))
	}

	status := "registered"
	if newParticipant.Status != nil {
		status = *newParticipant.Status
	}

	description := ""
	if exam.Description != nil {
		description = *exam.Description
	}

	return &dto.JoinExamResponse{
		ParticipantID: newParticipant.ID,
		ExamID:        exam.ID,
		Title:         exam.Title,
		Description:   description,
		StartTime:     exam.StartTime.Time.Format(time.RFC3339),
		EndTime:       exam.EndTime.Time.Format(time.RFC3339),
		DurationMins:  exam.DurationMinutes,
		TotalProblems: problemCount,
		Status:        status,
		CreatedAt:     newParticipant.CreatedAt.Time.Format(time.RFC3339),
	}, nil
}

func (su *studentExamUseCase) StartExam(ctx context.Context, examID, userID int64) (*dto.StartExamResponse, error) {
	participant, err := su.queries.GetParticipantStatus(ctx, models.GetParticipantStatusParams{
		ExamID: examID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("not registered for this exam: %w", err)
	}

	status := "registered"
	if participant.Status != nil {
		status = *participant.Status
	}

	if status != "registered" {
		return nil, fmt.Errorf("exam already started or completed")
	}

	exam, err := su.queries.GetExamForStudent(ctx, examID)
	if err != nil {
		return nil, fmt.Errorf("exam not found: %w", err)
	}

	now := time.Now()
	if now.Before(exam.StartTime.Time) {
		return nil, fmt.Errorf("exam has not started yet")
	}
	if now.After(exam.EndTime.Time) {
		return nil, fmt.Errorf("exam has ended")
	}

	updated, err := su.queries.StartExamParticipant(ctx, models.StartExamParticipantParams{
		ExamID: examID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start exam: %w", err)
	}

	timeRemaining := calculateTimeRemaining(updated.StartedAt.Time, exam.EndTime.Time)

	updatedStatus := "in_progress"
	if updated.Status != nil {
		updatedStatus = *updated.Status
	}

	return &dto.StartExamResponse{
		ParticipantID:   updated.ID,
		ExamID:          examID,
		StartedAt:       updated.StartedAt.Time.Format(time.RFC3339),
		TimeRemainingMs: timeRemaining,
		Status:          updatedStatus,
	}, nil
}

func (su *studentExamUseCase) GetExam(ctx context.Context, examID, userID int64) (*dto.GetExamResponse, error) {
	// 1. Verify participant is registered
	participant, err := su.queries.GetParticipantStatus(ctx, models.GetParticipantStatusParams{
		ExamID: examID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("not registered for this exam: %w", err)
	}

	// 2. Get exam metadata
	exam, err := su.queries.GetExamForStudent(ctx, examID)
	if err != nil {
		return nil, fmt.Errorf("exam not found or not published: %w", err)
	}

	// 3. Get all problems in exam
	problemRows, err := su.queries.GetExamProblemsForStudent(ctx, examID)
	if err != nil {
		return nil, fmt.Errorf("failed to load exam problems: %w", err)
	}

	// 4. Convert to DTOs
	problems := make([]dto.ExamProblemBrief, len(problemRows))
	for i, p := range problemRows {
		problems[i] = dto.ExamProblemBrief{
			ExamProblemID: p.ID,
			ProblemID:     p.ProblemID,
			Title:         p.Title,
			Difficulty:    p.Difficulty,
			Points:        p.Points,
			SortOrder:     p.SortOrder,
		}
	}

	// 5. Calculate time remaining
	timeRemaining := calculateTimeRemaining(time.Now(), exam.EndTime.Time)
	if timeRemaining < 0 {
		timeRemaining = 0
	}

	// 6. Get participant status
	status := "registered"
	if participant.Status != nil {
		status = *participant.Status
	}

	description := ""
	if exam.Description != nil {
		description = *exam.Description
	}

	return &dto.GetExamResponse{
		ExamID:            exam.ID,
		Title:             exam.Title,
		Description:       description,
		StartTime:         exam.StartTime.Time.Format(time.RFC3339),
		EndTime:           exam.EndTime.Time.Format(time.RFC3339),
		DurationMins:      exam.DurationMinutes,
		Status:            status,
		TimeRemainingMs:   timeRemaining,
		ParticipantStatus: status,
		Problems:          problems,
	}, nil
}

func (su *studentExamUseCase) GetProblem(ctx context.Context, examID, examProblemID, userID int64) (*dto.GetProblemResponse, error) {
	// 1. Verify participant is registered
	_, err := su.queries.GetParticipantStatus(ctx, models.GetParticipantStatusParams{
		ExamID: examID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("not registered for this exam: %w", err)
	}

	// 2. Get problem details
	problem, err := su.queries.GetExamProblemDetails(ctx, models.GetExamProblemDetailsParams{
		ExamID: examID,
		ID:     examProblemID,
	})
	if err != nil {
		return nil, fmt.Errorf("problem not found: %w", err)
	}

	// 3. Get all submissions for this problem
	submissionRows, err := su.queries.GetStudentSubmissionsForProblem(ctx, models.GetStudentSubmissionsForProblemParams{
		ExamID:        examID,
		ExamProblemID: examProblemID,
		UserID:        userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load submissions: %w", err)
	}

	// 4. Convert submissions to DTOs
	submissions := make([]dto.StudentSubmission, len(submissionRows))
	var attemptNumber int32 = 1
	for i, s := range submissionRows {
		status := s.Status
		score := numericToFloat64(s.Score)
		isCorrect := false
		if s.IsCorrect != nil {
			isCorrect = *s.IsCorrect
		}

		submittedAt := ""
		if s.SubmittedAt.Valid {
			submittedAt = s.SubmittedAt.Time.Format(time.RFC3339)
		}

		submissions[i] = dto.StudentSubmission{
			SubmissionID:    s.ID,
			Code:            s.Code,
			Status:          status,
			Score:           score,
			IsCorrect:       isCorrect,
			AttemptNumber:   *s.AttemptNumber,
			ExecutionTimeMs: s.ExecutionTimeMs,
			ErrorMessage:    s.ErrorMessage,
			SubmittedAt:     submittedAt,
		}
		if s.AttemptNumber != nil {
			attemptNumber = *s.AttemptNumber
		}
	}

	// 5. Build response
	initScript := problem.InitScript
	return &dto.GetProblemResponse{
		ExamProblemID: problem.ID,
		ProblemID:     problem.ProblemID,
		Title:         problem.Title,
		Description:   problem.Description,
		Difficulty:    problem.Difficulty,
		Points:        problem.Points,
		SortOrder:     problem.SortOrder,
		InitScript:    &initScript,
		AttemptNumber: attemptNumber,
		Submissions:   submissions,
	}, nil
}

func (su *studentExamUseCase) SubmitCode(ctx context.Context, examID, examProblemID, userID int64, req *dto.SubmitCodeRequest) (*dto.SubmitCodeResponse, error) {
	// 1. Verify participant is registered and in progress
	participant, err := su.queries.GetParticipantStatus(ctx, models.GetParticipantStatusParams{
		ExamID: examID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("not registered for this exam: %w", err)
	}

	status := "registered"
	if participant.Status != nil {
		status = *participant.Status
	}

	if status != "in_progress" {
		return nil, fmt.Errorf("exam not in progress")
	}

	// Kiểm tra thời hạn thi — không chấp nhận nộp bài sau khi hết giờ
	examInfo, err := su.queries.GetExamForStudent(ctx, examID)
	if err != nil {
		return nil, fmt.Errorf("exam not found: %w", err)
	}
	if examInfo.EndTime.Valid && time.Now().UTC().After(examInfo.EndTime.Time) {
		return nil, fmt.Errorf("exam time has expired, cannot submit")
	}

	// 2. Get problem details (includes init_script and solution_query)
	problem, err := su.queries.GetExamProblemDetails(ctx, models.GetExamProblemDetailsParams{
		ExamID: examID,
		ID:     examProblemID,
	})
	if err != nil {
		return nil, fmt.Errorf("problem not found: %w", err)
	}

	// 3. Check max attempts
	examFull, err := su.queries.GetExamByID(ctx, examID)
	if err != nil {
		return nil, fmt.Errorf("exam not found: %w", err)
	}

	submissionCount, err := su.queries.CountUserExamSubmissions(ctx, models.CountUserExamSubmissionsParams{
		ExamID:        examID,
		ExamProblemID: examProblemID,
		UserID:        userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check submission count: %w", err)
	}

	maxAttempts := int64(100) // Default: unlimited attempts
	if examFull.MaxAttempts != nil {
		maxAttempts = int64(*examFull.MaxAttempts)
	}

	if submissionCount >= maxAttempts {
		return nil, fmt.Errorf("max attempts exceeded")
	}

	// 4. Create submission record
	submission, err := su.queries.CreateExamSubmissionForStudent(ctx, models.CreateExamSubmissionForStudentParams{
		ExamID:        examID,
		ExamProblemID: examProblemID,
		UserID:        userID,
		Code:          req.Code,
		DatabaseType:  req.DatabaseType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create submission: %w", err)
	}

	// 5. Execute code
	timeout := 30 * time.Second
	execResult, err := su.executor.ExecuteCode(ctx, req.Code, problem.InitScript, problem.SolutionQuery, req.DatabaseType, timeout)
	if err != nil {
		return nil, fmt.Errorf("code execution failed: %w", err)
	}

	// 6. Convert output to JSON bytes
	actualOutput, _ := json.Marshal(execResult.Output)
	expectedOutput, _ := json.Marshal(execResult.ExpectedOutput)

	// 7. Update submission with results
	statusStr := "accepted"
	if !execResult.Success {
		statusStr = "error"
	} else if !execResult.IsCorrect {
		statusStr = "wrong_answer"
	}

	score := pgtype.Numeric{}
	if execResult.IsCorrect && problem.Points != nil {
		score.Int = big.NewInt(int64(*problem.Points))
		score.Valid = true
	}

	executionTimeMs := int32(execResult.ExecutionTime)

	updatedSubmission, err := su.queries.UpdateExamSubmissionWithResult(ctx, models.UpdateExamSubmissionWithResultParams{
		ID:              submission.ID,
		Status:          statusStr,
		ActualOutput:    actualOutput,
		ExpectedOutput:  expectedOutput,
		ErrorMessage:    &execResult.ErrorMessage,
		ExecutionTimeMs: &executionTimeMs,
		IsCorrect:       &execResult.IsCorrect,
		Score:           score,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update submission: %w", err)
	}

	// 8. Build response
	resultScore := numericToFloat64(updatedSubmission.Score)

	errorMsg := ""
	if updatedSubmission.ErrorMessage != nil {
		errorMsg = *updatedSubmission.ErrorMessage
	}

	submittedAtStr := ""
	if updatedSubmission.SubmittedAt.Valid {
		submittedAtStr = updatedSubmission.SubmittedAt.Time.Format(time.RFC3339)
	}

	scoringMode := "automatic"

	return &dto.SubmitCodeResponse{
		SubmissionID:    updatedSubmission.ID,
		ExamID:          examID,
		ExamProblemID:   examProblemID,
		Status:          statusStr,
		Score:           resultScore,
		IsCorrect:       execResult.IsCorrect,
		AttemptNumber:   *updatedSubmission.AttemptNumber,
		ExecutionTimeMs: &executionTimeMs,
		ErrorMessage:    &errorMsg,
		SubmittedAt:     submittedAtStr,
		ScoringMode:     scoringMode,
	}, nil
}

func (su *studentExamUseCase) SubmitExam(ctx context.Context, examID, userID int64) (*dto.SubmitExamResponse, error) {
	participant, err := su.queries.GetParticipantStatus(ctx, models.GetParticipantStatusParams{
		ExamID: examID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("not registered for this exam: %w", err)
	}

	status := "registered"
	if participant.Status != nil {
		status = *participant.Status
	}

	if status == "submitted" || status == "graded" {
		return nil, fmt.Errorf("exam already submitted")
	}

	updated, err := su.queries.SubmitExamParticipant(ctx, models.SubmitExamParticipantParams{
		ExamID: examID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to submit exam: %w", err)
	}

	row := su.db.GetPool().QueryRow(ctx,
		`SELECT COALESCE(SUM(es.score), 0) FROM exam_submissions es
		 WHERE es.exam_id = $1 AND es.user_id = $2 AND es.score IS NOT NULL`,
		examID, userID)

	var totalScore float64
	if err := row.Scan(&totalScore); err != nil {
		totalScore = 0
	}

	// Bug 2 Fix: Ghi lại điểm vào DB
	_, _ = su.db.GetPool().Exec(ctx,
		"UPDATE exam_participants SET total_score = $1 WHERE id = $2",
		totalScore, updated.ID,
	)

	updatedStatus := "submitted"
	if updated.Status != nil {
		updatedStatus = *updated.Status
	}

	return &dto.SubmitExamResponse{
		ParticipantID: updated.ID,
		ExamID:        examID,
		TotalScore:    totalScore,
		SubmittedAt:   updated.SubmittedAt.Time.Format(time.RFC3339),
		Status:        updatedStatus,
	}, nil
}

func (su *studentExamUseCase) GetTimeRemaining(ctx context.Context, examID, userID int64) (*dto.GetTimeRemainingResponse, error) {
	participant, err := su.queries.GetParticipantStatus(ctx, models.GetParticipantStatusParams{
		ExamID: examID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("not registered for this exam: %w", err)
	}

	exam, err := su.queries.GetExamForStudent(ctx, examID)
	if err != nil {
		return nil, fmt.Errorf("exam not found: %w", err)
	}

	status := "registered"
	if participant.Status != nil {
		status = *participant.Status
	}

	if status == "submitted" || status == "graded" {
		return &dto.GetTimeRemainingResponse{
			TimeRemainingMs: 0,
			ExamID:          examID,
			Status:          status,
			Message:         "Exam already submitted",
		}, nil
	}

	timeRemaining := calculateTimeRemaining(time.Now(), exam.EndTime.Time)
	if timeRemaining < 0 {
		timeRemaining = 0
	}

	participantStatus := "not_started"
	if participant.StartedAt.Valid {
		participantStatus = "in_progress"
	}

	return &dto.GetTimeRemainingResponse{
		TimeRemainingMs: timeRemaining,
		ExamID:          examID,
		Status:          participantStatus,
	}, nil
}

func calculateTimeRemaining(from, to time.Time) int64 {
	remaining := to.Sub(from)
	if remaining < 0 {
		return 0
	}
	return remaining.Milliseconds()
}
