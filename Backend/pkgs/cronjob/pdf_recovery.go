package cronjob

import (
	"context"
	"time"

	pdfRepo "backend/internals/pdf/repository"
	"backend/pkgs/logger"
)

// PDFRecoveryTask cleans up PDF uploads that got stuck in intermediate states
type PDFRecoveryTask struct {
	repo    pdfRepo.IPDFRepository
	timeout time.Duration
}

// NewPDFRecoveryTask creates a new recovery task
func NewPDFRecoveryTask(repo pdfRepo.IPDFRepository, timeout time.Duration) *PDFRecoveryTask {
	return &PDFRecoveryTask{
		repo:    repo,
		timeout: timeout,
	}
}

func (t *PDFRecoveryTask) Name() string {
	return "pdf_recovery"
}

func (t *PDFRecoveryTask) Execute(ctx context.Context) error {
	logger.Info("Running PDF recovery task...")
	err := t.repo.ResetStuckPDFUploads(ctx, t.timeout)
	if err != nil {
		return err
	}
	return nil
}
