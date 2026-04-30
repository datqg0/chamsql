package usecase

import (
	"context"
	"fmt"
	"time"

	"backend/internals/exam/domain"
	"backend/internals/exam/repository"
	"backend/pkgs/logger"
)

// IExamTimerUseCase defines the interface for exam timer operations
type IExamTimerUseCase interface {
	CheckAndExpireExams(ctx context.Context) error
}

// examTimerUseCase implements exam timer logic
type examTimerUseCase struct {
	repository repository.IExamRepository
	outboxRepo repository.IExamOutboxRepository
}

// NewExamTimerUseCase creates a new exam timer usecase
func NewExamTimerUseCase(
	repo repository.IExamRepository,
	outboxRepo repository.IExamOutboxRepository,
) IExamTimerUseCase {
	return &examTimerUseCase{
		repository: repo,
		outboxRepo: outboxRepo,
	}
}

// CheckAndExpireExams checks for exams that have passed their end time and publishes time_expired events
// This method runs periodically (every 10-30 seconds) to identify and notify about expired exams
func (u *examTimerUseCase) CheckAndExpireExams(ctx context.Context) error {
	now := time.Now().UTC()

	// Get all exams - we'll filter expired ones in-memory
	// Fetch in batches to avoid memory issues with large exam counts
	limit := int32(100)
	var offset int32 = 0
	expiredCount := 0
	hasMore := true

	for hasMore {
		exams, err := u.repository.List(ctx, limit, offset)
		if err != nil {
			return fmt.Errorf("failed to list exams: %w", err)
		}

		if len(exams) == 0 {
			hasMore = false
			break
		}

		for _, exam := range exams {
			// Check if exam has ended
			if !exam.EndTime.Valid {
				continue
			}

			endTime := exam.EndTime.Time
			if now.After(endTime) {
				// Check if exam status is still "ongoing" or "published"
				// (not yet marked as completed)
				if exam.Status != nil && (*exam.Status == "published" || *exam.Status == "ongoing") {
					// Publish exam.time_expired event
					payload := domain.ExamEventPayload{
						ExamID:          exam.ID,
						Title:           exam.Title,
						CreatedBy:       exam.CreatedBy,
						Status:          "expired",
						EndTime:         endTime,
						DurationMinutes: exam.DurationMinutes,
					}

					eventEnvelope := domain.NewExamEventEnvelope(
						domain.EventTypeExamTimeExpired,
						exam.ID,
						payload,
						fmt.Sprintf("exam-timer-%d", exam.ID),
					)

					if err := u.outboxRepo.PublishEvent(ctx, "chamsql-exam-events-v1", eventEnvelope); err != nil {
						logger.Error("Failed to publish exam.time_expired event for exam %d: %v", exam.ID, err)
						// Vẫn tiếp tục update DB trực tiếp dù Kafka lỗi
					} else {
						logger.Info("Published exam.time_expired event for exam %d (ended at %v)", exam.ID, endTime)
					}

					// Update DB trực tiếp — đảm bảo exam bị đánh dấu completed
					// dù Kafka down hoặc consumer chưa xử lý kịp
					if _, dbErr := u.repository.UpdateStatus(ctx, exam.ID, "completed"); dbErr != nil {
						logger.Error("Failed to directly update exam %d status to completed: %v", exam.ID, dbErr)
					}

					expiredCount++
				}
			}
		}

		// If we got fewer exams than the limit, there are no more pages
		if int32(len(exams)) < limit {
			hasMore = false
		} else {
			offset += limit
		}
	}

	if expiredCount > 0 {
		logger.Info("CheckAndExpireExams: Found %d expired exams", expiredCount)
	}

	return nil
}
