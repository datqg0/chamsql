package usecase

import (
	"context"
	"fmt"
	"time"

	"backend/db"
	"backend/internals/student/controller/dto"
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
	db      *db.Database
	queries *models.Queries
}

func NewStudentExamUseCase(database *db.Database) IStudentExamUseCase {
	return &studentExamUseCase{
		db:      database,
		queries: models.New(database.GetPool()),
	}
}

func (su *studentExamUseCase) JoinExam(ctx context.Context, examID, userID int64) (*dto.JoinExamResponse, error) {
	exam, err := su.queries.GetExamForStudent(ctx, examID)
	if err != nil {
		return nil, fmt.Errorf("exam not found or not published: %w", err)
	}

	// Check if user already joined
	_, err = su.queries.GetParticipantStatus(ctx, models.GetParticipantStatusParams{
		ExamID: examID,
		UserID: userID,
	})
	if err == nil {
		return nil, fmt.Errorf("already joined this exam")
	}

	// Add participant
	newParticipant, err := su.queries.AddParticipant(ctx, models.AddParticipantParams{
		ExamID: examID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to join exam: %w", err)
	}

	// Count problems in exam
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

	// Verify exam is within time window
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

	// Start exam
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
	exam, err := su.queries.GetExamForStudent(ctx, examID)
	if err != nil {
		return nil, fmt.Errorf("exam not found: %w", err)
	}

	participant, err := su.queries.GetParticipantStatus(ctx, models.GetParticipantStatusParams{
		ExamID: examID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("not registered for this exam: %w", err)
	}

	problems, err := su.queries.GetExamProblemsForStudent(ctx, examID)
	if err != nil {
		return nil, fmt.Errorf("failed to load problems: %w", err)
	}

	problemBriefs := make([]dto.ExamProblemBrief, 0, len(problems))
	for _, p := range problems {
		problemBriefs = append(problemBriefs, dto.ExamProblemBrief{
			ExamProblemID: p.ID,
			ProblemID:     p.ProblemID,
			Title:         p.Title,
			Difficulty:    p.Difficulty,
			Points:        p.Points,
			SortOrder:     p.SortOrder,
		})
	}

	timeRemaining := int64(0)
	if participant.StartedAt.Valid {
		timeRemaining = calculateTimeRemaining(participant.StartedAt.Time, exam.EndTime.Time)
	}

	description := ""
	if exam.Description != nil {
		description = *exam.Description
	}

	examStatus := "draft"
	if exam.Status != nil {
		examStatus = *exam.Status
	}

	participantStatus := "registered"
	if participant.Status != nil {
		participantStatus = *participant.Status
	}

	return &dto.GetExamResponse{
		ExamID:            examID,
		Title:             exam.Title,
		Description:       description,
		StartTime:         exam.StartTime.Time.Format(time.RFC3339),
		EndTime:           exam.EndTime.Time.Format(time.RFC3339),
		DurationMins:      exam.DurationMinutes,
		Status:            examStatus,
		TimeRemainingMs:   timeRemaining,
		ParticipantStatus: participantStatus,
		Problems:          problemBriefs,
	}, nil
}

func (su *studentExamUseCase) GetProblem(ctx context.Context, examID, examProblemID, userID int64) (*dto.GetProblemResponse, error) {
	// Verify student is registered for exam
	_, err := su.queries.GetParticipantStatus(ctx, models.GetParticipantStatusParams{
		ExamID: examID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("not registered for this exam: %w", err)
	}

	// Get problem details
	problem, err := su.queries.GetExamProblemDetails(ctx, models.GetExamProblemDetailsParams{
		ExamID: examID,
		ID:     examProblemID,
	})
	if err != nil {
		return nil, fmt.Errorf("problem not found: %w", err)
	}

	// Get student's submissions
	submissions, err := su.queries.GetStudentSubmissionsForProblem(ctx, models.GetStudentSubmissionsForProblemParams{
		ExamID:        examID,
		ExamProblemID: examProblemID,
		UserID:        userID,
	})

	var studentSubs []dto.StudentSubmission
	attemptNumber := int32(1)
	if err == nil {
		attemptNumber = int32(len(submissions)) + 1
		for _, sub := range submissions {
			score := 0.0
			// Convert pgtype.Numeric to float64
			if sub.Score.Valid && sub.Score.Int != nil {
				// Use Int64Value to extract the value
				if i64, err := sub.Score.Int64Value(); err == nil {
					score = float64(i64.Int64)
				}
			}

			isCorrect := false
			if sub.IsCorrect != nil {
				isCorrect = *sub.IsCorrect
			}

			attemptNum := int32(1)
			if sub.AttemptNumber != nil {
				attemptNum = *sub.AttemptNumber
			}

			studentSubs = append(studentSubs, dto.StudentSubmission{
				SubmissionID:    sub.ID,
				Code:            sub.Code,
				Status:          sub.Status,
				Score:           score,
				IsCorrect:       isCorrect,
				AttemptNumber:   attemptNum,
				ExecutionTimeMs: sub.ExecutionTimeMs,
				ErrorMessage:    sub.ErrorMessage,
				SubmittedAt:     sub.SubmittedAt.Time.Format(time.RFC3339),
			})
		}
	}

	return &dto.GetProblemResponse{
		ExamProblemID:   problem.ID,
		ProblemID:       problem.ProblemID,
		Title:           problem.Title,
		Description:     problem.Description,
		Difficulty:      problem.Difficulty,
		Points:          problem.Points,
		SortOrder:       problem.SortOrder,
		ScoringMode:     problem.ScoringMode,
		ReferenceAnswer: problem.ReferenceAnswer,
		InitScript:      &problem.InitScript,
		SolutionQuery:   &problem.SolutionQuery,
		AttemptNumber:   attemptNumber,
		Submissions:     studentSubs,
	}, nil
}

func (su *studentExamUseCase) SubmitCode(ctx context.Context, examID, examProblemID, userID int64, req *dto.SubmitCodeRequest) (*dto.SubmitCodeResponse, error) {
	if req == nil || req.Code == "" {
		return nil, fmt.Errorf("code cannot be empty")
	}

	// Verify student is in exam and exam is active
	participant, err := su.queries.GetParticipantStatus(ctx, models.GetParticipantStatusParams{
		ExamID: examID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("not registered for this exam: %w", err)
	}

	participantStatus := "registered"
	if participant.Status != nil {
		participantStatus = *participant.Status
	}

	if participantStatus != "in_progress" {
		return nil, fmt.Errorf("exam not in progress")
	}

	// Create submission
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

	// Get problem details for scoring mode
	problem, err := su.queries.GetExamProblemDetails(ctx, models.GetExamProblemDetailsParams{
		ExamID: examID,
		ID:     examProblemID,
	})
	if err != nil {
		return nil, fmt.Errorf("problem not found: %w", err)
	}

	// TODO: Execute code against database (will be in separate service)
	// For now, return pending status with score 0
	score := 0.0
	scoringMode := "manual"
	if problem.ScoringMode != nil {
		scoringMode = *problem.ScoringMode
	}

	submissionStatus := "pending"
	submissionErrorMsg := submission.ErrorMessage

	attemptNum := int32(1)
	if submission.AttemptNumber != nil {
		attemptNum = *submission.AttemptNumber
	}

	return &dto.SubmitCodeResponse{
		SubmissionID:    submission.ID,
		ExamID:          examID,
		ExamProblemID:   examProblemID,
		Status:          submissionStatus,
		Score:           score,
		IsCorrect:       false,
		AttemptNumber:   attemptNum,
		ExecutionTimeMs: submission.ExecutionTimeMs,
		ErrorMessage:    submissionErrorMsg,
		SubmittedAt:     submission.SubmittedAt.Time.Format(time.RFC3339),
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

	// Submit exam
	updated, err := su.queries.SubmitExamParticipant(ctx, models.SubmitExamParticipantParams{
		ExamID: examID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to submit exam: %w", err)
	}

	// Calculate total score from submissions
	row := su.db.GetPool().QueryRow(ctx,
		`SELECT COALESCE(SUM(es.score), 0) FROM exam_submissions es
		 WHERE es.exam_id = $1 AND es.user_id = $2 AND es.graded_by IS NOT NULL`,
		examID, userID)

	var totalScore float64
	if err := row.Scan(&totalScore); err != nil {
		totalScore = 0
	}

	// Update total score (will be calculated separately)
	// _, err = su.queries.UpdateParticipantScore(ctx, models.UpdateParticipantScoreParams{
	// 	ExamID:     examID,
	// 	UserID:     userID,
	// 	TotalScore: totalScore,
	// })
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to update score: %w", err)
	// }

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
