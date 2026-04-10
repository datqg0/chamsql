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
		 WHERE ep.user_id = $1 AND ep.submitted_at IS NOT NULL`
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

	dataQuery := `SELECT ep.exam_id, e.title, ep.total_score, ep.submitted_at, ep.status
		 FROM exam_participants ep
		 JOIN exams e ON e.id = ep.exam_id
		 WHERE ep.user_id = $1 AND ep.submitted_at IS NOT NULL`
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
	// NOTE: GetExamProblemsForStudent query not yet implemented
	return nil, fmt.Errorf("exam result detail loading not yet implemented")
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
		`SELECT ep.user_id, ep.total_score, 
		        ROW_NUMBER() OVER (ORDER BY ep.total_score DESC) as rank,
		        ROUND(100.0 * ROW_NUMBER() OVER (ORDER BY ep.total_score DESC) / 
		              COUNT(*) OVER ()) as percentile
		 FROM exam_participants ep
		 WHERE ep.exam_id = $1 AND ep.submitted_at IS NOT NULL
		 ORDER BY ep.total_score DESC
		 LIMIT $2 OFFSET $3`,
		examID, req.Limit, offset)

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
		Total:     total,
		Page:      req.Page,
		Limit:     req.Limit,
	}, nil
}

func (su *studentResultsUseCase) GetExamAnalytics(ctx context.Context, examID int64) (*dto.ExamAnalytics, error) {
	// NOTE: GetExamProblemsForStudent query not yet implemented
	return nil, fmt.Errorf("exam analytics loading not yet implemented")
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
