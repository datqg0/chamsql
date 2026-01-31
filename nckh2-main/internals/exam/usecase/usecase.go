package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"backend/configs"
	"backend/internals/exam/controller/dto"
	examRepo "backend/internals/exam/repository"
	problemRepo "backend/internals/problem/repository"
	"backend/pkgs/runner"
	"backend/sql/models"

	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrExamNotFound       = errors.New("exam not found")
	ErrNotParticipant     = errors.New("you are not a participant")
	ErrExamNotStarted     = errors.New("exam has not started yet")
	ErrExamEnded          = errors.New("exam has ended")
	ErrAlreadyStarted     = errors.New("you have already started this exam")
	ErrAlreadySubmitted   = errors.New("you have already submitted this exam")
	ErrMaxAttemptsReached = errors.New("maximum attempts reached")
	ErrTimeExpired        = errors.New("exam time has expired")
	ErrProblemNotInExam   = errors.New("problem not in this exam")
	ErrUnauthorized       = errors.New("unauthorized to perform this action")
)

type IExamUseCase interface {
	// Lecturer/Admin CRUD
	Create(ctx context.Context, userID int64, req *dto.CreateExamRequest) (*dto.ExamResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.ExamResponse, error)
	List(ctx context.Context, userID int64, role string, page, pageSize int) (*dto.ExamListResponse, error)
	Update(ctx context.Context, userID int64, examID int64, req *dto.UpdateExamRequest) (*dto.ExamResponse, error)
	Delete(ctx context.Context, userID int64, examID int64) error

	// Problem management
	AddProblem(ctx context.Context, userID int64, examID int64, req *dto.AddProblemRequest) error
	RemoveProblem(ctx context.Context, userID int64, examID, problemID int64) error
	ListProblems(ctx context.Context, examID int64) ([]dto.ExamProblemResponse, error)

	// Participant management
	AddParticipants(ctx context.Context, userID int64, examID int64, req *dto.AddParticipantsRequest) error
	RemoveParticipant(ctx context.Context, userID int64, examID, participantID int64) error
	ListParticipants(ctx context.Context, examID int64) ([]dto.ParticipantResponse, error)

	// Student actions
	StartExam(ctx context.Context, userID int64, examID int64) (*dto.StartExamResponse, error)
	SubmitAnswer(ctx context.Context, userID, examID int64, req *dto.ExamSubmitRequest) (*dto.ExamSubmitResponse, error)
	FinishExam(ctx context.Context, userID int64, examID int64) (*dto.ExamResultResponse, error)
	GetMyExams(ctx context.Context, userID int64) ([]dto.ExamResponse, error)
}

type examUseCase struct {
	examRepo    examRepo.IExamRepository
	problemRepo problemRepo.IProblemRepository
	runner      runner.Runner
	cfg         *configs.Config
}

func NewExamUseCase(
	examRepo examRepo.IExamRepository,
	problemRepo problemRepo.IProblemRepository,
	queryRunner runner.Runner,
	cfg *configs.Config,
) IExamUseCase {
	return &examUseCase{
		examRepo:    examRepo,
		problemRepo: problemRepo,
		runner:      queryRunner,
		cfg:         cfg,
	}
}

func (u *examUseCase) Create(ctx context.Context, userID int64, req *dto.CreateExamRequest) (*dto.ExamResponse, error) {
	maxAttempts := int32(req.MaxAttempts)
	if maxAttempts == 0 {
		maxAttempts = 1
	}

	exam, err := u.examRepo.Create(ctx, models.CreateExamParams{
		Title:                 req.Title,
		Description:           strPtr(req.Description),
		CreatedBy:             userID,
		StartTime:             timeToPg(req.StartTime),
		EndTime:               timeToPg(req.EndTime),
		DurationMinutes:       int32(req.DurationMinutes),
		AllowedDatabases:      req.AllowedDatabases,
		AllowAiAssistance:     &req.AllowAiAssistance,
		ShuffleProblems:       &req.ShuffleProblems,
		ShowResultImmediately: &req.ShowResultImmediately,
		MaxAttempts:           &maxAttempts,
		IsPublic:              &req.IsPublic,
	})
	if err != nil {
		return nil, err
	}

	return toExamResponseFromModel(exam), nil
}

func (u *examUseCase) GetByID(ctx context.Context, id int64) (*dto.ExamResponse, error) {
	exam, err := u.examRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrExamNotFound
	}

	problems, _ := u.examRepo.ListProblems(ctx, id)
	problemResponses := make([]dto.ExamProblemResponse, len(problems))
	for i, p := range problems {
		problemResponses[i] = dto.ExamProblemResponse{
			ID:         p.ID, // This is the exam_problems.id
			ProblemID:  p.ProblemID,
			Title:      p.Title,
			Slug:       p.Slug,
			Difficulty: p.Difficulty,
			Points:     int(ptrToInt32(p.Points)),
			SortOrder:  int(ptrToInt32(p.SortOrder)),
		}
	}

	resp := &dto.ExamResponse{
		ID:                    exam.ID,
		Title:                 exam.Title,
		Description:           ptrToStr(exam.Description),
		CreatedBy:             exam.CreatedBy,
		CreatorName:           exam.CreatorName, // string, not pointer
		StartTime:             pgToTime(exam.StartTime),
		EndTime:               pgToTime(exam.EndTime),
		DurationMinutes:       int(exam.DurationMinutes),
		AllowedDatabases:      exam.AllowedDatabases,
		AllowAiAssistance:     ptrToBool(exam.AllowAiAssistance),
		ShuffleProblems:       ptrToBool(exam.ShuffleProblems),
		ShowResultImmediately: ptrToBool(exam.ShowResultImmediately),
		MaxAttempts:           int(ptrToInt32(exam.MaxAttempts)),
		IsPublic:              ptrToBool(exam.IsPublic),
		Status:                ptrToStr(exam.Status),
		ProblemCount:          int64(len(problems)),
		Problems:              problemResponses,
		CreatedAt:             pgToTime(exam.CreatedAt),
	}

	return resp, nil
}

func (u *examUseCase) List(ctx context.Context, userID int64, role string, page, pageSize int) (*dto.ExamListResponse, error) {
	offset := int32((page - 1) * pageSize)
	limit := int32(pageSize)

	var exams []dto.ExamResponse

	if role == "admin" {
		rows, err := u.examRepo.List(ctx, limit, offset)
		if err != nil {
			return nil, err
		}
		exams = make([]dto.ExamResponse, len(rows))
		for i, e := range rows {
			exams[i] = dto.ExamResponse{
				ID:               e.ID,
				Title:            e.Title,
				StartTime:        pgToTime(e.StartTime),
				EndTime:          pgToTime(e.EndTime),
				DurationMinutes:  int(e.DurationMinutes),
				Status:           ptrToStr(e.Status),
				ProblemCount:     e.ProblemCount,
				ParticipantCount: e.ParticipantCount,
				CreatedAt:        pgToTime(e.CreatedAt),
			}
		}
	} else if role == "lecturer" {
		rows, err := u.examRepo.ListByLecturer(ctx, userID, limit, offset)
		if err != nil {
			return nil, err
		}
		exams = make([]dto.ExamResponse, len(rows))
		for i, e := range rows {
			exams[i] = dto.ExamResponse{
				ID:               e.ID,
				Title:            e.Title,
				StartTime:        pgToTime(e.StartTime),
				EndTime:          pgToTime(e.EndTime),
				DurationMinutes:  int(e.DurationMinutes),
				Status:           ptrToStr(e.Status),
				ProblemCount:     e.ProblemCount,
				ParticipantCount: e.ParticipantCount,
				CreatedAt:        pgToTime(e.CreatedAt),
			}
		}
	} else {
		rows, err := u.examRepo.ListPublic(ctx, limit, offset)
		if err != nil {
			return nil, err
		}
		exams = make([]dto.ExamResponse, len(rows))
		for i, e := range rows {
			exams[i] = dto.ExamResponse{
				ID:              e.ID,
				Title:           e.Title,
				StartTime:       pgToTime(e.StartTime),
				EndTime:         pgToTime(e.EndTime),
				DurationMinutes: int(e.DurationMinutes),
				Status:          ptrToStr(e.Status),
				CreatedAt:       pgToTime(e.CreatedAt),
			}
		}
	}

	return &dto.ExamListResponse{
		Exams:    exams,
		Total:    int64(len(exams)),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (u *examUseCase) Update(ctx context.Context, userID int64, examID int64, req *dto.UpdateExamRequest) (*dto.ExamResponse, error) {
	exam, err := u.examRepo.GetByID(ctx, examID)
	if err != nil {
		return nil, ErrExamNotFound
	}
	if exam.CreatedBy != userID {
		return nil, ErrUnauthorized
	}

	params := models.UpdateExamParams{ID: examID}
	if req.Title != nil {
		params.Title = req.Title
	}
	if req.Description != nil {
		params.Description = req.Description
	}
	if req.DurationMinutes != nil {
		d := int32(*req.DurationMinutes)
		params.DurationMinutes = &d
	}
	if req.AllowAiAssistance != nil {
		params.AllowAiAssistance = req.AllowAiAssistance
	}
	if req.ShuffleProblems != nil {
		params.ShuffleProblems = req.ShuffleProblems
	}
	if req.ShowResultImmediately != nil {
		params.ShowResultImmediately = req.ShowResultImmediately
	}
	if req.MaxAttempts != nil {
		m := int32(*req.MaxAttempts)
		params.MaxAttempts = &m
	}
	if req.IsPublic != nil {
		params.IsPublic = req.IsPublic
	}

	updated, err := u.examRepo.Update(ctx, params)
	if err != nil {
		return nil, err
	}

	return toExamResponseFromModel(updated), nil
}

func (u *examUseCase) Delete(ctx context.Context, userID int64, examID int64) error {
	exam, err := u.examRepo.GetByID(ctx, examID)
	if err != nil {
		return ErrExamNotFound
	}
	if exam.CreatedBy != userID {
		return ErrUnauthorized
	}
	return u.examRepo.Delete(ctx, examID)
}

// Problem management
func (u *examUseCase) AddProblem(ctx context.Context, userID int64, examID int64, req *dto.AddProblemRequest) error {
	exam, err := u.examRepo.GetByID(ctx, examID)
	if err != nil {
		return ErrExamNotFound
	}
	if exam.CreatedBy != userID {
		return ErrUnauthorized
	}

	points := int32(req.Points)
	sortOrder := int32(req.SortOrder)
	_, err = u.examRepo.AddProblem(ctx, models.AddProblemToExamParams{
		ExamID:    examID,
		ProblemID: req.ProblemID,
		Points:    &points,
		SortOrder: &sortOrder,
	})
	return err
}

func (u *examUseCase) RemoveProblem(ctx context.Context, userID int64, examID, problemID int64) error {
	exam, err := u.examRepo.GetByID(ctx, examID)
	if err != nil {
		return ErrExamNotFound
	}
	if exam.CreatedBy != userID {
		return ErrUnauthorized
	}
	return u.examRepo.RemoveProblem(ctx, examID, problemID)
}

func (u *examUseCase) ListProblems(ctx context.Context, examID int64) ([]dto.ExamProblemResponse, error) {
	problems, err := u.examRepo.ListProblems(ctx, examID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.ExamProblemResponse, len(problems))
	for i, p := range problems {
		result[i] = dto.ExamProblemResponse{
			ID:         p.ID,
			ProblemID:  p.ProblemID,
			Title:      p.Title,
			Slug:       p.Slug,
			Difficulty: p.Difficulty,
			Points:     int(ptrToInt32(p.Points)),
			SortOrder:  int(ptrToInt32(p.SortOrder)),
		}
	}
	return result, nil
}

// Participant management
func (u *examUseCase) AddParticipants(ctx context.Context, userID int64, examID int64, req *dto.AddParticipantsRequest) error {
	exam, err := u.examRepo.GetByID(ctx, examID)
	if err != nil {
		return ErrExamNotFound
	}
	if exam.CreatedBy != userID {
		return ErrUnauthorized
	}

	for _, uid := range req.UserIDs {
		_, _ = u.examRepo.AddParticipant(ctx, examID, uid)
	}
	return nil
}

func (u *examUseCase) RemoveParticipant(ctx context.Context, userID int64, examID, participantID int64) error {
	exam, err := u.examRepo.GetByID(ctx, examID)
	if err != nil {
		return ErrExamNotFound
	}
	if exam.CreatedBy != userID {
		return ErrUnauthorized
	}
	return u.examRepo.RemoveParticipant(ctx, examID, participantID)
}

func (u *examUseCase) ListParticipants(ctx context.Context, examID int64) ([]dto.ParticipantResponse, error) {
	participants, err := u.examRepo.ListParticipants(ctx, examID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.ParticipantResponse, len(participants))
	for i, p := range participants {
		result[i] = dto.ParticipantResponse{
			ID:         p.ID,
			UserID:     p.UserID,
			FullName:   p.FullName,
			Email:      p.Email,
			StudentID:  ptrToStr(p.StudentID),
			Status:     ptrToStr(p.Status),
			TotalScore: numericToFloat(p.TotalScore),
		}
		if p.StartedAt.Valid {
			s := p.StartedAt.Time.Format(time.RFC3339)
			result[i].StartedAt = &s
		}
		if p.SubmittedAt.Valid {
			s := p.SubmittedAt.Time.Format(time.RFC3339)
			result[i].SubmittedAt = &s
		}
	}
	return result, nil
}

// Student actions
func (u *examUseCase) StartExam(ctx context.Context, userID int64, examID int64) (*dto.StartExamResponse, error) {
	exam, err := u.examRepo.GetByID(ctx, examID)
	if err != nil {
		return nil, ErrExamNotFound
	}

	// Check if user is participant
	participant, err := u.examRepo.GetParticipant(ctx, examID, userID)
	if err != nil {
		return nil, ErrNotParticipant
	}

	// Check exam time window
	now := time.Now()
	if now.Before(exam.StartTime.Time) {
		return nil, ErrExamNotStarted
	}
	if now.After(exam.EndTime.Time) {
		return nil, ErrExamEnded
	}

	// Check if already submitted
	if ptrToStr(participant.Status) == "submitted" {
		return nil, ErrAlreadySubmitted
	}

	// Start exam
	if ptrToStr(participant.Status) != "in_progress" {
		_, err = u.examRepo.StartExam(ctx, examID, userID)
		if err != nil {
			return nil, err
		}
	}

	// Get problems
	problems, _ := u.examRepo.ListProblems(ctx, examID)
	problemResponses := make([]dto.ExamProblemResponse, len(problems))
	for i, p := range problems {
		problemResponses[i] = dto.ExamProblemResponse{
			ID:          p.ID,
			ProblemID:   p.ProblemID,
			Title:       p.Title,
			Slug:        p.Slug,
			Difficulty:  p.Difficulty,
			Description: p.Description,
			Points:      int(ptrToInt32(p.Points)),
		}
	}

	startedAt := time.Now()
	endsAt := startedAt.Add(time.Duration(exam.DurationMinutes) * time.Minute)

	return &dto.StartExamResponse{
		ExamID:          examID,
		Title:           exam.Title,
		DurationMinutes: int(exam.DurationMinutes),
		StartedAt:       startedAt.Format(time.RFC3339),
		EndsAt:          endsAt.Format(time.RFC3339),
		Problems:        problemResponses,
	}, nil
}

func (u *examUseCase) SubmitAnswer(ctx context.Context, userID, examID int64, req *dto.ExamSubmitRequest) (*dto.ExamSubmitResponse, error) {
	exam, err := u.examRepo.GetByID(ctx, examID)
	if err != nil {
		return nil, ErrExamNotFound
	}

	// Check participant
	participant, err := u.examRepo.GetParticipant(ctx, examID, userID)
	if err != nil {
		return nil, ErrNotParticipant
	}

	if ptrToStr(participant.Status) != "in_progress" {
		return nil, ErrAlreadySubmitted
	}

	// Find exam problem
	problems, _ := u.examRepo.ListProblems(ctx, examID)
	var examProblem *models.ListExamProblemsRow
	for _, p := range problems {
		if p.ProblemID == req.ProblemID {
			examProblem = &p
			break
		}
	}
	if examProblem == nil {
		return nil, ErrProblemNotInExam
	}

	// Check attempts
	attemptCount, _ := u.examRepo.CountExamSubmissions(ctx, examID, examProblem.ID, userID)
	if int(attemptCount) >= int(ptrToInt32(exam.MaxAttempts)) {
		return nil, ErrMaxAttemptsReached
	}

	// Get problem details
	problem, err := u.problemRepo.GetByID(ctx, req.ProblemID)
	if err != nil {
		return nil, err
	}

	// Execute solution
	dbType := runner.DBType(req.DatabaseType)
	expectedResult, err := u.runner.Execute(ctx, dbType, problem.SolutionQuery)
	if err != nil {
		return nil, err
	}

	// Execute user query
	actualResult, _ := u.runner.Execute(ctx, dbType, req.Code)

	// Compare
	orderMatters := ptrToBool(problem.OrderMatters)
	compareResult := u.runner.Compare(expectedResult, actualResult, orderMatters)

	// Calculate score
	var score float64
	maxScore := int(ptrToInt32(examProblem.Points))
	if compareResult.IsCorrect {
		score = float64(maxScore)
	}

	// Save submission
	expectedJSON, _ := json.Marshal(expectedResult.Rows)
	actualJSON, _ := json.Marshal(actualResult.Rows)
	execTimeMs := int32(actualResult.ExecutionMs)
	attemptNum := int32(attemptCount + 1)
	status := "wrong_answer"
	if compareResult.IsCorrect {
		status = "accepted"
	} else if actualResult.Error != "" {
		status = "error"
	}

	_, _ = u.examRepo.CreateExamSubmission(ctx, models.CreateExamSubmissionParams{
		ExamID:          examID,
		ExamProblemID:   examProblem.ID,
		UserID:          userID,
		Code:            req.Code,
		DatabaseType:    req.DatabaseType,
		Status:          status,
		ExecutionTimeMs: &execTimeMs,
		ExpectedOutput:  expectedJSON,
		ActualOutput:    actualJSON,
		ErrorMessage:    strPtr(actualResult.Error),
		IsCorrect:       &compareResult.IsCorrect,
		AttemptNumber:   &attemptNum,
	})

	return &dto.ExamSubmitResponse{
		IsCorrect:     compareResult.IsCorrect,
		Score:         score,
		MaxScore:      maxScore,
		ExecutionMs:   actualResult.ExecutionMs,
		Message:       compareResult.Message,
		Error:         actualResult.Error,
		AttemptNumber: int(attemptNum),
		MaxAttempts:   int(ptrToInt32(exam.MaxAttempts)),
	}, nil
}

func (u *examUseCase) FinishExam(ctx context.Context, userID int64, examID int64) (*dto.ExamResultResponse, error) {
	exam, err := u.examRepo.GetByID(ctx, examID)
	if err != nil {
		return nil, ErrExamNotFound
	}

	participant, err := u.examRepo.GetParticipant(ctx, examID, userID)
	if err != nil {
		return nil, ErrNotParticipant
	}

	// Submit exam
	_, err = u.examRepo.SubmitExam(ctx, examID, userID)
	if err != nil {
		return nil, err
	}

	return &dto.ExamResultResponse{
		ExamID:     examID,
		Title:      exam.Title,
		TotalScore: numericToFloat(participant.TotalScore),
		Status:     "submitted",
	}, nil
}

func (u *examUseCase) GetMyExams(ctx context.Context, userID int64) ([]dto.ExamResponse, error) {
	exams, err := u.examRepo.ListUserExams(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.ExamResponse, len(exams))
	for i, e := range exams {
		result[i] = dto.ExamResponse{
			ID:              e.ID,
			Title:           e.Title,
			StartTime:       pgToTime(e.StartTime),
			EndTime:         pgToTime(e.EndTime),
			DurationMinutes: int(e.DurationMinutes),
			Status:          ptrToStr(e.ParticipationStatus), // Correct field name
		}
	}
	return result, nil
}

// Helper functions
func toExamResponseFromModel(e *models.Exam) *dto.ExamResponse {
	return &dto.ExamResponse{
		ID:                    e.ID,
		Title:                 e.Title,
		Description:           ptrToStr(e.Description),
		CreatedBy:             e.CreatedBy,
		StartTime:             pgToTime(e.StartTime),
		EndTime:               pgToTime(e.EndTime),
		DurationMinutes:       int(e.DurationMinutes),
		AllowedDatabases:      e.AllowedDatabases,
		AllowAiAssistance:     ptrToBool(e.AllowAiAssistance),
		ShuffleProblems:       ptrToBool(e.ShuffleProblems),
		ShowResultImmediately: ptrToBool(e.ShowResultImmediately),
		MaxAttempts:           int(ptrToInt32(e.MaxAttempts)),
		IsPublic:              ptrToBool(e.IsPublic),
		Status:                ptrToStr(e.Status),
		CreatedAt:             pgToTime(e.CreatedAt),
	}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ptrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptrToBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func ptrToInt32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

func timeToPg(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func pgToTime(t pgtype.Timestamptz) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format(time.RFC3339)
}

func numericToFloat(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, _ := n.Float64Value()
	return f.Float64
}
