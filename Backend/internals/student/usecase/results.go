package usecase

import (
	"context"
	"fmt"
	"time"

	"backend/db"
	"backend/internals/student/controller/dto"
	"backend/sql/models"
)

type IStudentResultsUseCase interface {
	GetExamResults(ctx context.Context, userID int64) (*dto.ListExamResultsResponse, error)
	GetExamResultDetail(ctx context.Context, examID, userID int64) (*dto.ExamResultDetail, error)
	GetClassRanking(ctx context.Context, examID int64) (*dto.ClassRankingResponse, error)
	GetExamAnalytics(ctx context.Context, examID int64) (*dto.ExamAnalytics, error)
}

type studentResultsUseCase struct {
	db      *db.Database
	queries *models.Queries
}

func NewStudentResultsUseCase(database *db.Database) IStudentResultsUseCase {
	return &studentResultsUseCase{
		db:      database,
		queries: models.New(database.GetPool()),
	}
}

func (su *studentResultsUseCase) GetExamResults(ctx context.Context, userID int64) (*dto.ListExamResultsResponse, error) {
	rows := su.db.GetPool().QueryRow(ctx,
		`SELECT COUNT(*) FROM exam_participants
		 WHERE user_id = $1 AND submitted_at IS NOT NULL`,
		userID)

	var total int64
	if err := rows.Scan(&total); err != nil {
		total = 0
	}

	resultRows, err := su.db.GetPool().Query(ctx,
		`SELECT ep.exam_id, e.title, ep.total_score, ep.submitted_at, ep.status
		 FROM exam_participants ep
		 JOIN exams e ON e.id = ep.exam_id
		 WHERE ep.user_id = $1 AND ep.submitted_at IS NOT NULL
		 ORDER BY ep.submitted_at DESC`,
		userID)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch exam results: %w", err)
	}
	defer resultRows.Close()

	results := make([]dto.ExamResult, 0)
	for resultRows.Next() {
		var examID int64
		var title string
		var totalScore interface{}
		var submittedAt time.Time
		var status string

		if err := resultRows.Scan(&examID, &title, &totalScore, &submittedAt, &status); err != nil {
			continue
		}

		score := 0.0
		if ts, ok := totalScore.(float64); ok {
			score = ts
		}

		results = append(results, dto.ExamResult{
			ExamID:      examID,
			Title:       title,
			StudentID:   userID,
			TotalScore:  score,
			SubmittedAt: submittedAt.Format(time.RFC3339),
			Status:      status,
		})
	}

	return &dto.ListExamResultsResponse{
		Results: results,
		Total:   total,
	}, nil
}

func (su *studentResultsUseCase) GetExamResultDetail(ctx context.Context, examID, userID int64) (*dto.ExamResultDetail, error) {
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

	if !participant.SubmittedAt.Valid {
		return nil, fmt.Errorf("exam not submitted yet")
	}

	problems, err := su.queries.GetExamProblemsForStudent(ctx, examID)
	if err != nil {
		return nil, fmt.Errorf("failed to load problems: %w", err)
	}

	problemDetails := make([]dto.ProblemResultDetail, 0, len(problems))
	for _, p := range problems {
		submissions, err := su.queries.GetStudentSubmissionsForProblem(ctx, models.GetStudentSubmissionsForProblemParams{
			ExamID:        examID,
			ExamProblemID: p.ID,
			UserID:        userID,
		})

		score := 0.0
		isCorrect := false
		status := "pending"
		attempts := int32(0)
		var graderComment *string

		if err == nil {
			attempts = int32(len(submissions))
			if len(submissions) > 0 {
				lastSubmission := submissions[len(submissions)-1]
				score = convertNumericToFloat64(lastSubmission.Score)
				if lastSubmission.IsCorrect != nil {
					isCorrect = *lastSubmission.IsCorrect
				}
				status = lastSubmission.Status
				graderComment = lastSubmission.ErrorMessage
			}
		}

		problemDetails = append(problemDetails, dto.ProblemResultDetail{
			ExamProblemID: p.ID,
			ProblemID:     p.ProblemID,
			Title:         p.Title,
			Difficulty:    p.Difficulty,
			Points:        p.Points,
			StudentScore:  score,
			IsCorrect:     isCorrect,
			Status:        status,
			Attempts:      attempts,
			GraderComment: graderComment,
		})
	}

	description := ""
	if exam.Description != nil {
		description = *exam.Description
	}

	status := "submitted"
	if exam.Status != nil {
		status = *exam.Status
	}

	return &dto.ExamResultDetail{
		ExamID:       examID,
		Title:        exam.Title,
		Description:  description,
		TotalScore:   convertNumericToFloat64(participant.TotalScore),
		DurationMins: exam.DurationMinutes,
		StartTime:    exam.StartTime.Time.Format(time.RFC3339),
		EndTime:      exam.EndTime.Time.Format(time.RFC3339),
		SubmittedAt:  participant.SubmittedAt.Time.Format(time.RFC3339),
		Status:       status,
		Problems:     problemDetails,
	}, nil
}

func (su *studentResultsUseCase) GetClassRanking(ctx context.Context, examID int64) (*dto.ClassRankingResponse, error) {
	exam, err := su.queries.GetExamForStudent(ctx, examID)
	if err != nil {
		return nil, fmt.Errorf("exam not found: %w", err)
	}

	rows, err := su.db.GetPool().Query(ctx,
		`SELECT ep.user_id, ep.total_score, 
		        ROW_NUMBER() OVER (ORDER BY ep.total_score DESC) as rank,
		        ROUND(100.0 * ROW_NUMBER() OVER (ORDER BY ep.total_score DESC) / 
		              COUNT(*) OVER ()) as percentile
		 FROM exam_participants ep
		 WHERE ep.exam_id = $1 AND ep.submitted_at IS NOT NULL
		 ORDER BY ep.total_score DESC`,
		examID)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch rankings: %w", err)
	}
	defer rows.Close()

	rankings := make([]dto.StudentRanking, 0)
	for rows.Next() {
		var userID int64
		var totalScore interface{}
		var rank, percentile int32

		if err := rows.Scan(&userID, &totalScore, &rank, &percentile); err != nil {
			continue
		}

		score := 0.0
		if ts, ok := totalScore.(float64); ok {
			score = ts
		}

		rankings = append(rankings, dto.StudentRanking{
			Rank:        rank,
			StudentID:   userID,
			StudentName: fmt.Sprintf("Student %d", userID),
			Score:       score,
			Percentile:  float64(percentile),
		})
	}

	return &dto.ClassRankingResponse{
		ExamID:    examID,
		ExamTitle: exam.Title,
		Rankings:  rankings,
	}, nil
}

func (su *studentResultsUseCase) GetExamAnalytics(ctx context.Context, examID int64) (*dto.ExamAnalytics, error) {
	exam, err := su.queries.GetExamForStudent(ctx, examID)
	if err != nil {
		return nil, fmt.Errorf("exam not found: %w", err)
	}

	statsRow := su.db.GetPool().QueryRow(ctx,
		`SELECT COUNT(*) as total_students,
		        COALESCE(AVG(ep.total_score), 0) as avg_score,
		        COALESCE(MAX(ep.total_score), 0) as highest_score,
		        COALESCE(MIN(ep.total_score), 0) as lowest_score,
		        COUNT(CASE WHEN ep.total_score >= 60 THEN 1 END) * 100.0 / COUNT(*) as pass_rate
		 FROM exam_participants ep
		 WHERE ep.exam_id = $1 AND ep.submitted_at IS NOT NULL`,
		examID)

	var totalStudents int64
	var avgScore, highestScore, lowestScore, passRate float64
	if err := statsRow.Scan(&totalStudents, &avgScore, &highestScore, &lowestScore, &passRate); err != nil {
		return nil, fmt.Errorf("failed to fetch analytics: %w", err)
	}

	problems, err := su.queries.GetExamProblemsForStudent(ctx, examID)
	if err != nil {
		return nil, fmt.Errorf("failed to load problems: %w", err)
	}

	problemStats := make([]dto.ProblemStat, 0, len(problems))
	for _, p := range problems {
		probRow := su.db.GetPool().QueryRow(ctx,
			`SELECT COALESCE(AVG(es.score), 0) as avg_score,
			        COUNT(CASE WHEN es.is_correct = true THEN 1 END) * 100.0 / COUNT(*) as correct_rate,
			        COALESCE(AVG(es.attempt_number), 0) as avg_attempts
			 FROM exam_submissions es
			 WHERE es.exam_id = $1 AND es.exam_problem_id = $2`,
			examID, p.ID)

		var probAvgScore, probCorrectRate, probAvgAttempts float64
		if err := probRow.Scan(&probAvgScore, &probCorrectRate, &probAvgAttempts); err != nil {
			probAvgScore = 0
			probCorrectRate = 0
			probAvgAttempts = 0
		}

		problemStats = append(problemStats, dto.ProblemStat{
			ProblemID:   p.ProblemID,
			Title:       p.Title,
			Points:      p.Points,
			AvgScore:    probAvgScore,
			CorrectRate: probCorrectRate,
			AvgAttempts: probAvgAttempts,
		})
	}

	return &dto.ExamAnalytics{
		ExamID:        examID,
		Title:         exam.Title,
		TotalStudents: totalStudents,
		AvgScore:      avgScore,
		HighestScore:  highestScore,
		LowestScore:   lowestScore,
		PassRate:      passRate,
		ProblemsStats: problemStats,
	}, nil
}

func convertNumericToFloat64(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case int64:
		return float64(v)
	default:
		return 0
	}
}
