package repository

import (
	"context"

	"backend/db"
	"backend/sql/models"

	"github.com/jackc/pgx/v5/pgtype"
)

type IExamRepository interface {
	// Exam CRUD
	Create(ctx context.Context, params models.CreateExamParams) (*models.Exam, error)
	GetByID(ctx context.Context, id int64) (*models.GetExamByIDRow, error)
	List(ctx context.Context, limit, offset int32) ([]models.ListExamsRow, error)
	ListByLecturer(ctx context.Context, lecturerID int64, limit, offset int32) ([]models.ListExamsByLecturerRow, error)
	ListPublic(ctx context.Context, limit, offset int32) ([]models.ListPublicExamsRow, error)
	Update(ctx context.Context, params models.UpdateExamParams) (*models.Exam, error)
	UpdateStatus(ctx context.Context, id int64, status string) (*models.Exam, error)
	Delete(ctx context.Context, id int64) error

	// Exam Problems
	AddProblem(ctx context.Context, params models.AddProblemToExamParams) (*models.ExamProblem, error)
	ListProblems(ctx context.Context, examID int64) ([]models.ListExamProblemsRow, error)
	RemoveProblem(ctx context.Context, examID, problemID int64) error
	UpdateProblemPoints(ctx context.Context, examID, problemID int64, points int32) error

	// Participants
	AddParticipant(ctx context.Context, examID, userID int64) (*models.ExamParticipant, error)
	GetParticipant(ctx context.Context, examID, userID int64) (*models.GetParticipantRow, error)
	ListParticipants(ctx context.Context, examID int64) ([]models.ListExamParticipantsRow, error)
	StartExam(ctx context.Context, examID, userID int64) (*models.ExamParticipant, error)
	SubmitExam(ctx context.Context, examID, userID int64) (*models.ExamParticipant, error)
	UpdateScore(ctx context.Context, examID, userID int64, score float64) error
	RemoveParticipant(ctx context.Context, examID, userID int64) error

	// Student's exams
	ListUserExams(ctx context.Context, userID int64) ([]models.ListUserExamsRow, error)

	// Exam Submissions
	CreateExamSubmission(ctx context.Context, params models.CreateExamSubmissionParams) (*models.ExamSubmission, error)
	GetExamSubmission(ctx context.Context, examID, examProblemID, userID int64) (*models.ExamSubmission, error)
	CountExamSubmissions(ctx context.Context, examID, examProblemID, userID int64) (int64, error)
	GetExamResults(ctx context.Context, examID int64) ([]models.GetExamResultsRow, error)
}

type examRepository struct {
	db      *db.Database
	queries *models.Queries
}

func NewExamRepository(database *db.Database) IExamRepository {
	return &examRepository{
		db:      database,
		queries: models.New(database.GetPool()),
	}
}

func (r *examRepository) Create(ctx context.Context, params models.CreateExamParams) (*models.Exam, error) {
	exam, err := r.queries.CreateExam(ctx, params)
	if err != nil {
		return nil, err
	}
	return &exam, nil
}

func (r *examRepository) GetByID(ctx context.Context, id int64) (*models.GetExamByIDRow, error) {
	exam, err := r.queries.GetExamByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &exam, nil
}

func (r *examRepository) List(ctx context.Context, limit, offset int32) ([]models.ListExamsRow, error) {
	return r.queries.ListExams(ctx, models.ListExamsParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *examRepository) ListByLecturer(ctx context.Context, lecturerID int64, limit, offset int32) ([]models.ListExamsByLecturerRow, error) {
	return r.queries.ListExamsByLecturer(ctx, models.ListExamsByLecturerParams{
		CreatedBy: lecturerID,
		Limit:     limit,
		Offset:    offset,
	})
}

func (r *examRepository) ListPublic(ctx context.Context, limit, offset int32) ([]models.ListPublicExamsRow, error) {
	return r.queries.ListPublicExams(ctx, models.ListPublicExamsParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *examRepository) Update(ctx context.Context, params models.UpdateExamParams) (*models.Exam, error) {
	exam, err := r.queries.UpdateExam(ctx, params)
	if err != nil {
		return nil, err
	}
	return &exam, nil
}

func (r *examRepository) UpdateStatus(ctx context.Context, id int64, status string) (*models.Exam, error) {
	exam, err := r.queries.UpdateExamStatus(ctx, models.UpdateExamStatusParams{
		ID:     id,
		Status: &status,
	})
	if err != nil {
		return nil, err
	}
	return &exam, nil
}

func (r *examRepository) Delete(ctx context.Context, id int64) error {
	return r.queries.DeleteExam(ctx, id)
}

// Exam Problems
func (r *examRepository) AddProblem(ctx context.Context, params models.AddProblemToExamParams) (*models.ExamProblem, error) {
	ep, err := r.queries.AddProblemToExam(ctx, params)
	if err != nil {
		return nil, err
	}
	return &ep, nil
}

func (r *examRepository) ListProblems(ctx context.Context, examID int64) ([]models.ListExamProblemsRow, error) {
	return r.queries.ListExamProblems(ctx, examID)
}

func (r *examRepository) RemoveProblem(ctx context.Context, examID, problemID int64) error {
	return r.queries.RemoveProblemFromExam(ctx, models.RemoveProblemFromExamParams{
		ExamID:    examID,
		ProblemID: problemID,
	})
}

func (r *examRepository) UpdateProblemPoints(ctx context.Context, examID, problemID int64, points int32) error {
	pts := points
	_, err := r.queries.UpdateExamProblemPoints(ctx, models.UpdateExamProblemPointsParams{
		ExamID:    examID,
		ProblemID: problemID,
		Points:    &pts,
	})
	return err
}

// Participants
func (r *examRepository) AddParticipant(ctx context.Context, examID, userID int64) (*models.ExamParticipant, error) {
	p, err := r.queries.AddParticipant(ctx, models.AddParticipantParams{
		ExamID: examID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *examRepository) GetParticipant(ctx context.Context, examID, userID int64) (*models.GetParticipantRow, error) {
	p, err := r.queries.GetParticipant(ctx, models.GetParticipantParams{
		ExamID: examID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *examRepository) ListParticipants(ctx context.Context, examID int64) ([]models.ListExamParticipantsRow, error) {
	return r.queries.ListExamParticipants(ctx, examID)
}

func (r *examRepository) StartExam(ctx context.Context, examID, userID int64) (*models.ExamParticipant, error) {
	p, err := r.queries.StartExam(ctx, models.StartExamParams{
		ExamID: examID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *examRepository) SubmitExam(ctx context.Context, examID, userID int64) (*models.ExamParticipant, error) {
	p, err := r.queries.SubmitExam(ctx, models.SubmitExamParams{
		ExamID: examID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *examRepository) UpdateScore(ctx context.Context, examID, userID int64, score float64) error {
	_, err := r.queries.UpdateParticipantScore(ctx, models.UpdateParticipantScoreParams{
		ExamID:     examID,
		UserID:     userID,
		TotalScore: pgtype.Numeric{Valid: true},
	})
	return err
}

func (r *examRepository) RemoveParticipant(ctx context.Context, examID, userID int64) error {
	return r.queries.RemoveParticipant(ctx, models.RemoveParticipantParams{
		ExamID: examID,
		UserID: userID,
	})
}

func (r *examRepository) ListUserExams(ctx context.Context, userID int64) ([]models.ListUserExamsRow, error) {
	return r.queries.ListUserExams(ctx, userID)
}

// Exam Submissions
func (r *examRepository) CreateExamSubmission(ctx context.Context, params models.CreateExamSubmissionParams) (*models.ExamSubmission, error) {
	s, err := r.queries.CreateExamSubmission(ctx, params)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *examRepository) GetExamSubmission(ctx context.Context, examID, examProblemID, userID int64) (*models.ExamSubmission, error) {
	s, err := r.queries.GetExamSubmission(ctx, models.GetExamSubmissionParams{
		ExamID:        examID,
		ExamProblemID: examProblemID,
		UserID:        userID,
	})
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *examRepository) CountExamSubmissions(ctx context.Context, examID, examProblemID, userID int64) (int64, error) {
	return r.queries.CountUserExamSubmissions(ctx, models.CountUserExamSubmissionsParams{
		ExamID:        examID,
		ExamProblemID: examProblemID,
		UserID:        userID,
	})
}

func (r *examRepository) GetExamResults(ctx context.Context, examID int64) ([]models.GetExamResultsRow, error) {
	return r.queries.GetExamResults(ctx, examID)
}
