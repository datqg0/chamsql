package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"backend/db"
	"backend/internals/student/controller/dto"
	"backend/pkgs/runner"
	"backend/sql/models"
)

// IPracticeUseCase - Practice problem functionality for students
type IPracticeUseCase interface {
	ListPublicProblems(ctx context.Context, page, pageSize int, difficulty, topic string) (*dto.ListPublicProblemsResponse, error)
	GetPublicProblem(ctx context.Context, problemID, userID int64) (*dto.GetPublicProblemResponse, error)
	GetPublicProblemBySlug(ctx context.Context, slug string, userID int64) (*dto.GetPublicProblemResponse, error)
	PracticeSubmitCode(ctx context.Context, problemID, userID int64, req *dto.PracticeSubmitCodeRequest) (*dto.PracticeSubmitCodeResponse, error)
	ListPracticeSubmissions(ctx context.Context, problemID, userID int64, page, pageSize int) (*dto.ListPracticeSubmissionsResponse, error)
}

type practiceUseCase struct {
	db       *db.Database
	queries  *models.Queries
	executor CodeExecutor
}

// NewPracticeUseCase - Create new practice usecase
func NewPracticeUseCase(database *db.Database, queryRunner runner.Runner) IPracticeUseCase {
	return &practiceUseCase{
		db:       database,
		queries:  models.New(database.GetPool()),
		executor: NewCodeExecutor(queryRunner),
	}
}

// ListPublicProblems - List all public problems available for practice
func (p *practiceUseCase) ListPublicProblems(ctx context.Context, page, pageSize int, difficulty, topic string) (*dto.ListPublicProblemsResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := int32((page - 1) * pageSize)
	limit := int32(pageSize)

	// List public problems
	problems, err := p.queries.ListProblems(ctx, models.ListProblemsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list public problems: %w", err)
	}

	// Convert to DTOs
	problemResponses := make([]dto.PublicProblemBrief, 0)
	for _, problem := range problems {
		topicName := ""
		topicSlug := ""
		if problem.TopicName != nil {
			topicName = *problem.TopicName
		}
		if problem.TopicSlug != nil {
			topicSlug = *problem.TopicSlug
		}

		hints := (*string)(nil)
		if len(problem.Hints) > 0 {
			hintsStr := string(problem.Hints)
			hints = &hintsStr
		}

		createdBy := int64(0)
		if problem.CreatedBy != nil {
			createdBy = *problem.CreatedBy
		}

		problemResponses = append(problemResponses, dto.PublicProblemBrief{
			ProblemID:   problem.ID,
			Title:       problem.Title,
			Slug:        problem.Slug,
			Description: problem.Description,
			Difficulty:  problem.Difficulty,
			TopicID:     problem.TopicID,
			TopicName:   topicName,
			TopicSlug:   topicSlug,
			Hints:       hints,
			CreatedBy:   createdBy,
			CreatedAt:   problem.CreatedAt.Time.Format(time.RFC3339),
		})
	}

	// Get total count
	total, err := p.queries.CountProblems(ctx)
	if err != nil {
		total = int64(len(problemResponses))
	}

	return &dto.ListPublicProblemsResponse{
		Problems: problemResponses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetPublicProblem - Get full details of a public problem by ID
func (p *practiceUseCase) GetPublicProblem(ctx context.Context, problemID, userID int64) (*dto.GetPublicProblemResponse, error) {
	// Get problem details
	problem, err := p.queries.GetProblemByID(ctx, problemID)
	if err != nil {
		return nil, fmt.Errorf("problem not found: %w", err)
	}

	// Check if problem is public
	if problem.IsPublic == nil || !*problem.IsPublic {
		return nil, fmt.Errorf("problem is not public")
	}

	// Get user's practice stats for this problem
	userProgress, err := p.queries.GetProblemWithUserProgress(ctx, models.GetProblemWithUserProgressParams{
		Slug:   problem.Slug,
		UserID: userID,
	})

	practiceStats := (*dto.PracticeStats)(nil)
	if err == nil {
		attempts := int64(0)
		if userProgress.Attempts != nil {
			attempts = int64(*userProgress.Attempts)
		}

		correctAttempts := int64(0)
		if userProgress.IsSolved != nil && *userProgress.IsSolved {
			correctAttempts = 1
		}

		practiceStats = &dto.PracticeStats{
			TotalAttempts:   attempts,
			CorrectAttempts: correctAttempts,
			IsSolved:        userProgress.IsSolved != nil && *userProgress.IsSolved,
			BestTimeMs:      userProgress.BestTimeMs,
		}
	}

	// Get latest submissions for user (up to 5)
	latestSubmissions, _ := p.queries.ListUserSubmissionsForProblem(ctx, models.ListUserSubmissionsForProblemParams{
		UserID:    userID,
		ProblemID: problemID,
		Limit:     5,
	})

	practiceSubmissions := make([]dto.PracticeSubmission, 0)
	for i, submission := range latestSubmissions {
		errorMsg := submission.ErrorMessage
		executionTime := submission.ExecutionTimeMs

		practiceSubmissions = append(practiceSubmissions, dto.PracticeSubmission{
			SubmissionID:    submission.ID,
			Code:            submission.Code,
			Status:          submission.Status,
			IsCorrect:       submission.IsCorrect != nil && *submission.IsCorrect,
			AttemptNumber:   int32(i + 1),
			ExecutionTimeMs: executionTime,
			ErrorMessage:    errorMsg,
			SubmittedAt:     submission.SubmittedAt.Time.Format(time.RFC3339),
		})
	}

	sampleOutput := ""
	if len(problem.SampleOutput) > 0 {
		sampleOutput = string(problem.SampleOutput)
	}

	hints := (*string)(nil)
	if len(problem.Hints) > 0 {
		hintsStr := string(problem.Hints)
		hints = &hintsStr
	}

	orderMatters := (*bool)(nil)
	if problem.OrderMatters != nil {
		orderMatters = problem.OrderMatters
	}

	supportedDatabases := []string{}
	if problem.SupportedDatabases != nil {
		supportedDatabases = problem.SupportedDatabases
	}

	return &dto.GetPublicProblemResponse{
		ProblemID:          problem.ID,
		Title:              problem.Title,
		Slug:               problem.Slug,
		Description:        problem.Description,
		Difficulty:         problem.Difficulty,
		TopicID:            problem.TopicID,
		InitScript:         &problem.InitScript,
		SampleOutput:       &sampleOutput,
		Hints:              hints,
		OrderMatters:       orderMatters,
		SupportedDatabases: supportedDatabases,
		PracticeStats:      practiceStats,
		LatestSubmissions:  practiceSubmissions,
	}, nil
}

// GetPublicProblemBySlug - Get full details of a public problem by slug
func (p *practiceUseCase) GetPublicProblemBySlug(ctx context.Context, slug string, userID int64) (*dto.GetPublicProblemResponse, error) {
	// Get problem by slug
	problem, err := p.queries.GetProblemBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("problem not found: %w", err)
	}

	// Check if problem is public
	if problem.IsPublic == nil || !*problem.IsPublic {
		return nil, fmt.Errorf("problem is not public")
	}

	// Get user's practice stats
	userProgress, err := p.queries.GetProblemWithUserProgress(ctx, models.GetProblemWithUserProgressParams{
		Slug:   slug,
		UserID: userID,
	})

	practiceStats := (*dto.PracticeStats)(nil)
	if err == nil {
		attempts := int64(0)
		if userProgress.Attempts != nil {
			attempts = int64(*userProgress.Attempts)
		}

		correctAttempts := int64(0)
		if userProgress.IsSolved != nil && *userProgress.IsSolved {
			correctAttempts = 1
		}

		practiceStats = &dto.PracticeStats{
			TotalAttempts:   attempts,
			CorrectAttempts: correctAttempts,
			IsSolved:        userProgress.IsSolved != nil && *userProgress.IsSolved,
			BestTimeMs:      userProgress.BestTimeMs,
		}
	}

	// Get latest submissions
	latestSubmissions, _ := p.queries.ListUserSubmissionsForProblem(ctx, models.ListUserSubmissionsForProblemParams{
		UserID:    userID,
		ProblemID: problem.ID,
		Limit:     5,
	})

	practiceSubmissions := make([]dto.PracticeSubmission, 0)
	for i, submission := range latestSubmissions {
		errorMsg := submission.ErrorMessage
		executionTime := submission.ExecutionTimeMs

		practiceSubmissions = append(practiceSubmissions, dto.PracticeSubmission{
			SubmissionID:    submission.ID,
			Code:            submission.Code,
			Status:          submission.Status,
			IsCorrect:       submission.IsCorrect != nil && *submission.IsCorrect,
			AttemptNumber:   int32(i + 1),
			ExecutionTimeMs: executionTime,
			ErrorMessage:    errorMsg,
			SubmittedAt:     submission.SubmittedAt.Time.Format(time.RFC3339),
		})
	}

	sampleOutput := ""
	if len(problem.SampleOutput) > 0 {
		sampleOutput = string(problem.SampleOutput)
	}

	hints := (*string)(nil)
	if len(problem.Hints) > 0 {
		hintsStr := string(problem.Hints)
		hints = &hintsStr
	}

	orderMatters := (*bool)(nil)
	if problem.OrderMatters != nil {
		orderMatters = problem.OrderMatters
	}

	supportedDatabases := []string{}
	if problem.SupportedDatabases != nil {
		supportedDatabases = problem.SupportedDatabases
	}

	return &dto.GetPublicProblemResponse{
		ProblemID:          problem.ID,
		Title:              problem.Title,
		Slug:               problem.Slug,
		Description:        problem.Description,
		Difficulty:         problem.Difficulty,
		TopicID:            problem.TopicID,
		InitScript:         &problem.InitScript,
		SampleOutput:       &sampleOutput,
		Hints:              hints,
		OrderMatters:       orderMatters,
		SupportedDatabases: supportedDatabases,
		PracticeStats:      practiceStats,
		LatestSubmissions:  practiceSubmissions,
	}, nil
}

// PracticeSubmitCode - Submit code for practice problem (not in exam)
func (p *practiceUseCase) PracticeSubmitCode(ctx context.Context, problemID, userID int64, req *dto.PracticeSubmitCodeRequest) (*dto.PracticeSubmitCodeResponse, error) {
	// 1. Get problem details
	problem, err := p.queries.GetProblemByID(ctx, problemID)
	if err != nil {
		return nil, fmt.Errorf("problem not found: %w", err)
	}

	// Check if problem is public
	if problem.IsPublic == nil || !*problem.IsPublic {
		return nil, fmt.Errorf("problem is not public")
	}

	// 2. Create submission record
	dbType := req.DatabaseType
	if dbType == "" {
		dbType = "postgresql"
	}

	submission, err := p.queries.CreateSubmission(ctx, models.CreateSubmissionParams{
		UserID:       userID,
		ProblemID:    problemID,
		Code:         req.Code,
		DatabaseType: dbType,
		Status:       "pending",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create submission: %w", err)
	}

	// 3. Execute code
	timeout := 30 * time.Second
	execResult, err := p.executor.ExecuteCode(ctx, req.Code, problem.InitScript, problem.SolutionQuery, dbType, timeout)
	if err != nil {
		return nil, fmt.Errorf("code execution failed: %w", err)
	}

	// 4. Convert output to JSON strings
	actualOutput, _ := json.Marshal(execResult.Output)
	expectedOutput, _ := json.Marshal(execResult.ExpectedOutput)

	// 5. Determine status and score
	statusStr := "accepted"
	if !execResult.Success {
		statusStr = "error"
	} else if !execResult.IsCorrect {
		statusStr = "wrong_answer"
	}

	isCorrect := execResult.IsCorrect
	executionTimeMs := int32(execResult.ExecutionTime)

	// 6. Update submission with results directly via database
	updateSQL := `UPDATE submissions SET
		status = $2,
		actual_output = $3,
		expected_output = $4,
		error_message = $5,
		execution_time_ms = $6,
		is_correct = $7
	WHERE id = $1`

	_, err = p.db.GetPool().Exec(ctx, updateSQL,
		submission.ID,
		statusStr,
		actualOutput,
		expectedOutput,
		execResult.ErrorMessage,
		executionTimeMs,
		isCorrect,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update submission: %w", err)
	}

	// 7. Mark problem as solved if correct
	if isCorrect {
		_, _ = p.queries.MarkProblemSolved(ctx, models.MarkProblemSolvedParams{
			UserID:    userID,
			ProblemID: problemID,
		})
	}

	// 8. Get updated stats for response
	totalAttempts, _ := p.queries.CountUserSubmissions(ctx, userID)
	correctAttempts, _ := p.queries.CountCorrectSubmissions(ctx, userID)

	return &dto.PracticeSubmitCodeResponse{
		SubmissionID:    submission.ID,
		ProblemID:       problemID,
		Status:          statusStr,
		IsCorrect:       isCorrect,
		ExecutionTimeMs: &executionTimeMs,
		ErrorMessage:    &execResult.ErrorMessage,
		ActualOutput:    pointer(string(actualOutput)),
		ExpectedOutput:  pointer(string(expectedOutput)),
		SubmittedAt:     submission.SubmittedAt.Time.Format(time.RFC3339),
		AttemptNumber:   1,
		TotalAttempts:   totalAttempts,
		CorrectAttempts: correctAttempts,
	}, nil
}

// Helper function to return pointer to string
func pointer(s string) *string {
	return &s
}

// ListPracticeSubmissions - List practice submissions for a problem
func (p *practiceUseCase) ListPracticeSubmissions(ctx context.Context, problemID, userID int64, page, pageSize int) (*dto.ListPracticeSubmissionsResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Get submissions
	submissions, err := p.queries.ListUserSubmissionsForProblem(ctx, models.ListUserSubmissionsForProblemParams{
		UserID:    userID,
		ProblemID: problemID,
		Limit:     int32(pageSize),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list submissions: %w", err)
	}

	// Convert to DTOs
	submissionResponses := make([]dto.PracticeSubmission, 0)
	for i, submission := range submissions {
		errorMsg := submission.ErrorMessage
		executionTime := submission.ExecutionTimeMs

		submissionResponses = append(submissionResponses, dto.PracticeSubmission{
			SubmissionID:    submission.ID,
			Code:            submission.Code,
			Status:          submission.Status,
			IsCorrect:       submission.IsCorrect != nil && *submission.IsCorrect,
			AttemptNumber:   int32(i + 1),
			ExecutionTimeMs: executionTime,
			ErrorMessage:    errorMsg,
			SubmittedAt:     submission.SubmittedAt.Time.Format(time.RFC3339),
		})
	}

	// Get total count
	total, err := p.queries.CountUserSubmissions(ctx, userID)
	if err != nil {
		total = int64(len(submissionResponses))
	}

	return &dto.ListPracticeSubmissionsResponse{
		Submissions: submissionResponses,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}
