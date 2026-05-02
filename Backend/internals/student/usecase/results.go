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
	GetExamResults(ctx context.Context, userID int64, req *dto.ListExamResultsRequest) (*dto.ListExamResultsResponse, error)
	GetExamResultDetail(ctx context.Context, examID, userID int64) (*dto.ExamResultDetail, error)
	GetClassRanking(ctx context.Context, examID int64, req *dto.RankingRequest) (*dto.ClassRankingResponse, error)
	GetExamAnalytics(ctx context.Context, examID int64) (*dto.ExamAnalytics, error)
	// GetMySubmissions trả về lịch sử nộp bài luyện tập tổng hợp của sinh viên
	GetMySubmissions(ctx context.Context, userID int64, page, pageSize int) (*dto.MySubmissionsResponse, error)
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

func (su *studentResultsUseCase) GetExamResults(ctx context.Context, userID int64, req *dto.ListExamResultsRequest) (*dto.ListExamResultsResponse, error) {
	if req == nil {
		req = &dto.ListExamResultsRequest{
			Page:  1,
			Limit: 10,
		}
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	countQuery := `SELECT COUNT(*) FROM exam_participants ep
		 JOIN exams e ON e.id = ep.exam_id
		 WHERE ep.user_id = $1
		   AND (
				ep.submitted_at IS NOT NULL
				OR EXISTS (
					SELECT 1
					FROM exam_submissions es
					WHERE es.exam_id = ep.exam_id AND es.user_id = ep.user_id
				)
		   )`
	countParams := []interface{}{userID}

	if req.Status != "" {
		countQuery += ` AND ep.status = $2`
		countParams = append(countParams, req.Status)
	}

	if req.ScoreMin > 0 || req.ScoreMax > 0 {
		if req.ScoreMin > 0 && req.ScoreMax > 0 {
			countQuery += ` AND ep.total_score BETWEEN $` + fmt.Sprintf("%d AND $%d", len(countParams)+1, len(countParams)+2)
			countParams = append(countParams, req.ScoreMin, req.ScoreMax)
		} else if req.ScoreMin > 0 {
			countQuery += ` AND ep.total_score >= $` + fmt.Sprintf("%d", len(countParams)+1)
			countParams = append(countParams, req.ScoreMin)
		} else if req.ScoreMax > 0 {
			countQuery += ` AND ep.total_score <= $` + fmt.Sprintf("%d", len(countParams)+1)
			countParams = append(countParams, req.ScoreMax)
		}
	}

	if req.StartDate != "" {
		countQuery += ` AND ep.submitted_at >= $` + fmt.Sprintf("%d", len(countParams)+1)
		countParams = append(countParams, req.StartDate)
	}

	if req.EndDate != "" {
		countQuery += ` AND ep.submitted_at <= $` + fmt.Sprintf("%d", len(countParams)+1)
		countParams = append(countParams, req.EndDate)
	}

	rows := su.db.GetPool().QueryRow(ctx, countQuery, countParams...)

	var total int64
	if err := rows.Scan(&total); err != nil {
		total = 0
	}

	offset := (req.Page - 1) * req.Limit

	dataQuery := `SELECT
			ep.exam_id,
			e.title,
			ep.total_score,
			COALESCE(
				ep.submitted_at,
				(
					SELECT MAX(es.submitted_at)
					FROM exam_submissions es
					WHERE es.exam_id = ep.exam_id AND es.user_id = ep.user_id
				)
			) AS submitted_at,
			ep.status
		 FROM exam_participants ep
		 JOIN exams e ON e.id = ep.exam_id
		 WHERE ep.user_id = $1
		   AND (
				ep.submitted_at IS NOT NULL
				OR EXISTS (
					SELECT 1
					FROM exam_submissions es
					WHERE es.exam_id = ep.exam_id AND es.user_id = ep.user_id
				)
		   )`
	dataParams := []interface{}{userID}

	if req.Status != "" {
		dataQuery += ` AND ep.status = $` + fmt.Sprintf("%d", len(dataParams)+1)
		dataParams = append(dataParams, req.Status)
	}

	if req.ScoreMin > 0 || req.ScoreMax > 0 {
		if req.ScoreMin > 0 && req.ScoreMax > 0 {
			dataQuery += ` AND ep.total_score BETWEEN $` + fmt.Sprintf("%d AND $%d", len(dataParams)+1, len(dataParams)+2)
			dataParams = append(dataParams, req.ScoreMin, req.ScoreMax)
		} else if req.ScoreMin > 0 {
			dataQuery += ` AND ep.total_score >= $` + fmt.Sprintf("%d", len(dataParams)+1)
			dataParams = append(dataParams, req.ScoreMin)
		} else if req.ScoreMax > 0 {
			dataQuery += ` AND ep.total_score <= $` + fmt.Sprintf("%d", len(dataParams)+1)
			dataParams = append(dataParams, req.ScoreMax)
		}
	}

	if req.StartDate != "" {
		dataQuery += ` AND ep.submitted_at >= $` + fmt.Sprintf("%d", len(dataParams)+1)
		dataParams = append(dataParams, req.StartDate)
	}

	if req.EndDate != "" {
		dataQuery += ` AND ep.submitted_at <= $` + fmt.Sprintf("%d", len(dataParams)+1)
		dataParams = append(dataParams, req.EndDate)
	}

	dataQuery += ` ORDER BY ep.submitted_at DESC LIMIT $` + fmt.Sprintf("%d OFFSET $%d", len(dataParams)+1, len(dataParams)+2)
	dataParams = append(dataParams, req.Limit, offset)

	resultRows, err := su.db.GetPool().Query(ctx, dataQuery, dataParams...)

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
		Page:    req.Page,
		Limit:   req.Limit,
	}, nil
}

func (su *studentResultsUseCase) GetExamResultDetail(ctx context.Context, examID, userID int64) (*dto.ExamResultDetail, error) {
	// Lấy thông tin participant
	var status string
	var totalScore float64
	var submittedAt time.Time

	err := su.db.GetPool().QueryRow(ctx,
		`SELECT ep.status, COALESCE(ep.total_score, 0), COALESCE(ep.submitted_at, NOW())
         FROM exam_participants ep
         WHERE ep.exam_id = $1 AND ep.user_id = $2`,
		examID, userID,
	).Scan(&status, &totalScore, &submittedAt)
	if err != nil {
		return nil, fmt.Errorf("not registered for this exam")
	}

	// Lấy danh sách submissions của user trong exam này
	rows, err := su.db.GetPool().Query(ctx,
		`SELECT es.problem_id, p.title, p.slug,
                COALESCE(es.score, 0), es.is_correct,
                es.submitted_at
         FROM exam_submissions es
         JOIN problems p ON p.id = es.problem_id
         WHERE es.exam_id = $1 AND es.user_id = $2
         ORDER BY es.problem_id`,
		examID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get submissions: %w", err)
	}
	defer rows.Close()

	var submissions []dto.ExamSubmissionResult
	for rows.Next() {
		var s dto.ExamSubmissionResult
		var sat time.Time
		if err := rows.Scan(&s.ProblemID, &s.ProblemTitle, &s.ProblemSlug,
			&s.Score, &s.IsCorrect, &sat); err != nil {
			continue
		}
		s.SubmittedAt = sat.Format(time.RFC3339)
		submissions = append(submissions, s)
	}
	if submissions == nil {
		submissions = []dto.ExamSubmissionResult{}
	}

	return &dto.ExamResultDetail{
		ExamID:      examID,
		UserID:      userID,
		TotalScore:  totalScore,
		Status:      status,
		SubmittedAt: submittedAt.Format(time.RFC3339),
		Submissions: submissions,
	}, nil
}

func (su *studentResultsUseCase) GetClassRanking(ctx context.Context, examID int64, req *dto.RankingRequest) (*dto.ClassRankingResponse, error) {
	if req == nil {
		req = &dto.RankingRequest{
			Page:  1,
			Limit: 50,
		}
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 50
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	exam, err := su.queries.GetExamForStudent(ctx, examID)
	if err != nil {
		return nil, fmt.Errorf("exam not found: %w", err)
	}

	countRow := su.db.GetPool().QueryRow(ctx,
		`SELECT COUNT(*) FROM exam_participants ep
		 WHERE ep.exam_id = $1 AND ep.submitted_at IS NOT NULL`,
		examID)

	var total int64
	if err := countRow.Scan(&total); err != nil {
		total = 0
	}

	offset := (req.Page - 1) * req.Limit

	rows, err := su.db.GetPool().Query(ctx,
		`SELECT ep.user_id, u.full_name,
            COALESCE(ep.total_score, 0),
            ROW_NUMBER() OVER (ORDER BY ep.total_score DESC NULLS LAST) as rank,
            ROUND(100.0 * ROW_NUMBER() OVER (ORDER BY ep.total_score DESC NULLS LAST) /
                  NULLIF(COUNT(*) OVER (), 0)) as percentile
     FROM exam_participants ep
     JOIN users u ON u.id = ep.user_id
     WHERE ep.exam_id = $1 AND ep.submitted_at IS NOT NULL
     ORDER BY ep.total_score DESC NULLS LAST
     LIMIT $2 OFFSET $3`,
		examID, req.Limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch rankings: %w", err)
	}
	defer rows.Close()

	rankings := make([]dto.StudentRanking, 0)
	for rows.Next() {
		var userID int64
		var fullName string
		var totalScore interface{}
		var rank, percentile int32

		if err := rows.Scan(&userID, &fullName, &totalScore, &rank, &percentile); err != nil {
			continue
		}

		score := 0.0
		if ts, ok := totalScore.(float64); ok {
			score = ts
		}

		rankings = append(rankings, dto.StudentRanking{
			Rank:        rank,
			StudentID:   userID,
			StudentName: fullName,
			Score:       score,
			Percentile:  float64(percentile),
		})
	}

	return &dto.ClassRankingResponse{
		ExamID:    examID,
		ExamTitle: exam.Title,
		Rankings:  rankings,
		Total:     total,
		Page:      req.Page,
		Limit:     req.Limit,
	}, nil
}

func (su *studentResultsUseCase) GetExamAnalytics(ctx context.Context, examID int64) (*dto.ExamAnalytics, error) {
	return &dto.ExamAnalytics{
		ExamID:  examID,
		Message: "Analytics feature coming soon.",
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

// GetMySubmissions trả về toàn bộ lịch sử nộp bài luyện tập của một sinh viên
func (su *studentResultsUseCase) GetMySubmissions(ctx context.Context, userID int64, page, pageSize int) (*dto.MySubmissionsResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Total count
	var total int64
	_ = su.db.GetPool().QueryRow(ctx,
		"SELECT COUNT(*) FROM submissions WHERE user_id = $1",
		userID,
	).Scan(&total)

	rows, err := su.db.GetPool().Query(ctx,
		`SELECT s.id, s.problem_id, p.title, p.slug, s.code, s.status,
		        s.is_correct, s.execution_time_ms, s.error_message, s.submitted_at
		 FROM submissions s
		 JOIN problems p ON p.id = s.problem_id
		 WHERE s.user_id = $1
		 ORDER BY s.submitted_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, pageSize, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list submissions: %w", err)
	}
	defer rows.Close()

	var items []dto.SubmissionRecord
	for rows.Next() {
		var r dto.SubmissionRecord
		var submittedAt time.Time
		if err := rows.Scan(
			&r.SubmissionID, &r.ProblemID, &r.ProblemTitle, &r.ProblemSlug,
			&r.Code, &r.Status, &r.IsCorrect, &r.ExecutionTimeMs,
			&r.ErrorMessage, &submittedAt,
		); err != nil {
			continue
		}
		r.SubmittedAt = submittedAt.Format(time.RFC3339)
		items = append(items, r)
	}

	return &dto.MySubmissionsResponse{
		Submissions: items,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}
