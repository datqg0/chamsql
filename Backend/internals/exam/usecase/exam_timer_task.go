package usecase

import (
	"context"
	"fmt"

	"backend/pkgs/cronjob"
	"backend/pkgs/logger"
)

// ExamTimerTask implements the cronjob.Task interface for checking exam expiration
type ExamTimerTask struct {
	useCase IExamTimerUseCase
}

// NewExamTimerTask creates a new exam timer task
func NewExamTimerTask(useCase IExamTimerUseCase) cronjob.Task {
	return &ExamTimerTask{
		useCase: useCase,
	}
}

// Name returns the name of the task
func (t *ExamTimerTask) Name() string {
	return "ExamTimerTask"
}

// Execute runs the exam timer check
func (t *ExamTimerTask) Execute(ctx context.Context) error {
	logger.Debug("ExamTimerTask: Starting exam expiration check")

	if err := t.useCase.CheckAndExpireExams(ctx); err != nil {
		return fmt.Errorf("exam timer task failed: %w", err)
	}

	logger.Debug("ExamTimerTask: Completed exam expiration check")
	return nil
}
