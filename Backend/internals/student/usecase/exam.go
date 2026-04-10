package usecase

import (
	"context"
	"fmt"
	"time"

	"backend/db"
	"backend/internals/student/controller/dto"
	"backend/pkgs/redis"
	"backend/sql/models"
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

func NewStudentExamUseCase(database *db.Database, cache redis.IRedis) IStudentExamUseCase {
	return &studentExamUseCase{
		db:       database,
		queries:  models.New(database.GetPool()),
		executor: NewCodeExecutor(database),
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

	// NOTE: GetExamProblemsForStudent query not yet implemented
	// Using placeholder value for now
	problemCount := int64(0)
	// problems, err := su.queries.GetExamProblemsForStudent(ctx, examID)
	// if err == nil {
	// 	problemCount = int64(len(problems))
	// }

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
	// NOTE: GetExamProblemsForStudent query not yet implemented
	return nil, fmt.Errorf("exam problem loading not yet implemented")
}

func (su *studentExamUseCase) GetProblem(ctx context.Context, examID, examProblemID, userID int64) (*dto.GetProblemResponse, error) {
	// NOTE: GetExamProblemDetails query not yet implemented
	return nil, fmt.Errorf("exam problem loading not yet implemented")
}

func (su *studentExamUseCase) SubmitCode(ctx context.Context, examID, examProblemID, userID int64, req *dto.SubmitCodeRequest) (*dto.SubmitCodeResponse, error) {
	// NOTE: GetExamProblemDetails query not yet implemented
	return nil, fmt.Errorf("exam submission functionality not yet implemented")
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
		 WHERE es.exam_id = $1 AND es.user_id = $2 AND es.graded_by IS NOT NULL`,
		examID, userID)

	var totalScore float64
	if err := row.Scan(&totalScore); err != nil {
		totalScore = 0
	}

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
