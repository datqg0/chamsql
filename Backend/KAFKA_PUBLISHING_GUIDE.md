# Kafka Event Publishing Guide - Quick Reference

## Publishing Events in Use Cases

### 1. Create Outbox Repository Interface

First, update your domain repositories to support outbox event publishing:

```go
// In internals/exam/repository/outbox_repository.go
package repository

import (
	"context"
)

type IOutboxRepository interface {
	Save(ctx context.Context, topic string, payload []byte) error
}

type OutboxRepository struct {
	db Database
}

func NewOutboxRepository(db Database) IOutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) Save(ctx context.Context, topic string, payload []byte) error {
	_, err := r.db.GetPool().Exec(
		ctx,
		`INSERT INTO outbox_events (topic, payload, status) 
		 VALUES ($1, $2, 'pending')`,
		topic,
		payload,
	)
	return err
}
```

### 2. Inject Outbox Repository into Use Case

```go
type examUseCase struct {
	examRepo      examRepo.IExamRepository
	problemRepo   problemRepo.IProblemRepository
	outboxRepo    OutboxRepository  // Add this
	runner        runner.Runner
	cfg           *configs.Config
}

func NewExamUseCase(
	examRepo examRepo.IExamRepository,
	problemRepo problemRepo.IProblemRepository,
	outboxRepo OutboxRepository,  // Add this
	queryRunner runner.Runner,
	cfg *configs.Config,
) IExamUseCase {
	return &examUseCase{
		examRepo:    examRepo,
		problemRepo: problemRepo,
		outboxRepo:  outboxRepo,  // Add this
		runner:      queryRunner,
		cfg:         cfg,
	}
}
```

### 3. Register Outbox Repository in DI Container

```go
// In di/container.go
providers := []interface{}{
	// ... existing providers ...
	
	// Exam domain
	examRepo.NewExamRepository,
	examRepo.NewOutboxRepository,  // Add this
	examUc.NewExamUseCase,
	
	// ... other providers ...
}
```

### 4. Publish Events in Use Cases

```go
// Example: In exam/usecase/usecase.go

func (u *examUseCase) Create(ctx context.Context, userID int64, req *dto.CreateExamRequest) (*dto.ExamResponse, error) {
	// Create exam in database
	exam, err := u.examRepo.Create(ctx, models.CreateExamParams{
		Title:           req.Title,
		Description:     strPtr(req.Description),
		CreatedBy:       userID,
		StartTime:       timeToPg(req.StartTime),
		EndTime:         timeToPg(req.EndTime),
		DurationMinutes: int32(req.DurationMinutes),
		// ... other fields
	})
	if err != nil {
		return nil, err
	}

	// ✅ Publish ExamCreated event to outbox
	payload := exam_domain.ExamEventPayload{
		ExamID:          exam.ID,
		Title:           exam.Title,
		CreatedBy:       userID,
		StartTime:       exam.StartTime.Time,
		EndTime:         exam.EndTime.Time,
		DurationMinutes: exam.DurationMinutes,
	}
	eventBytes := exam_domain.NewExamEventEnvelope(
		exam_domain.EventTypeExamCreated,
		exam.ID,
		payload,
		fmt.Sprintf("exam-create-%d", exam.ID),
	)
	
	if err := u.outboxRepo.Save(ctx, kafka_config.TopicExamEvents, eventBytes); err != nil {
		logger.Warn("Failed to save exam creation event to outbox: %v", err)
		// Continue anyway - event will be lost but operation succeeds
	}

	return mapper.ToDTO(exam), nil
}

func (u *examUseCase) StartExam(ctx context.Context, userID int64, examID int64) (*dto.StartExamResponse, error) {
	// Start exam logic...
	
	// ✅ Publish ExamStarted event
	payload := exam_domain.ExamEventPayload{
		ExamID: examID,
		UserID: userID,
	}
	eventBytes := exam_domain.NewExamEventEnvelope(
		exam_domain.EventTypeExamStarted,
		examID,
		payload,
		fmt.Sprintf("exam-start-%d-%d", examID, userID),
	)
	u.outboxRepo.Save(ctx, kafka_config.TopicExamEvents, eventBytes)
	
	return response, nil
}

func (u *examUseCase) FinishExam(ctx context.Context, userID int64, examID int64) (*dto.ExamResultResponse, error) {
	// Grade exam and calculate score...
	result := &ExamResult{
		Score:    totalScore,
		MaxScore: maxScore,
	}
	
	// ✅ Publish ExamFinished event
	payload := exam_domain.ExamEventPayload{
		ExamID:   examID,
		UserID:   userID,
		Score:    result.Score,
		MaxScore: result.MaxScore,
	}
	eventBytes := exam_domain.NewExamEventEnvelope(
		exam_domain.EventTypeExamFinished,
		examID,
		payload,
		fmt.Sprintf("exam-finish-%d-%d", examID, userID),
	)
	u.outboxRepo.Save(ctx, kafka_config.TopicExamEvents, eventBytes)
	
	return mapper.ToResultDTO(result), nil
}
```

### 5. Similar Pattern for Submission Events

```go
// In submission/usecase/usecase.go

func (u *submissionUseCase) Create(ctx context.Context, examID, userID int64) error {
	// Create submission...
	
	// ✅ Publish SubmissionCreated event
	payload := submission_domain.SubmissionEventPayload{
		SubmissionID: submission.ID,
		ExamID:       examID,
		UserID:       userID,
		Status:       "submitted",
		SubmittedAt:  time.Now(),
	}
	eventBytes := submission_domain.NewSubmissionEventEnvelope(
		submission_domain.EventTypeSubmissionCreated,
		submission.ID,
		payload,
		fmt.Sprintf("submission-create-%d", submission.ID),
	)
	u.outboxRepo.Save(ctx, kafka_config.TopicSubmissionEvents, eventBytes)
	
	return nil
}

func (u *submissionUseCase) Grade(ctx context.Context, submissionID int64, score float64, gradedBy int64) error {
	// Grade submission...
	
	// ✅ Publish SubmissionGraded event
	payload := submission_domain.SubmissionEventPayload{
		SubmissionID: submissionID,
		Score:        score,
		GradedBy:     gradedBy,
		GradedAt:     time.Now(),
	}
	eventBytes := submission_domain.NewSubmissionEventEnvelope(
		submission_domain.EventTypeSubmissionGraded,
		submissionID,
		payload,
		fmt.Sprintf("submission-grade-%d", submissionID),
	)
	u.outboxRepo.Save(ctx, kafka_config.TopicSubmissionEvents, eventBytes)
	
	return nil
}
```

## Event Consumption

Events are automatically consumed and processed by:
- `exam_event_consumer.go` - Handles exam domain events
- `submission_event_consumer.go` - Handles submission domain events

Both consumers implement:
- **Idempotent processing**: Checks if event was already processed
- **Error recovery**: Logs errors but doesn't crash
- **Graceful shutdown**: Responds to context cancellation

## Extending Event Handlers

To add side effects when events occur, update the consumer handlers:

```go
// In exam_event_consumer.go

func (c *ExamEventConsumer) handleExamFinished(ctx context.Context, envelope *messaging.EventEnvelope) {
	var payload exam_domain.ExamEventPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		logger.Error("Failed to unmarshal ExamFinished payload: %v", err)
		return
	}

	logger.Info("Exam finished: examID=%d, userID=%d, score=%.2f", 
		payload.ExamID, payload.UserID, payload.Score)

	// TODO: Add side effects here:
	// 1. Send notification email to student with results
	// 2. Update user leaderboard/statistics
	// 3. Trigger next exam if part of a series
	// 4. Archive exam results
	// 5. Update dashboard with new completion data
	
	// Example: Send notification
	// if err := c.notificationService.SendExamResultNotification(ctx, payload); err != nil {
	//     logger.Error("Failed to send result notification: %v", err)
	// }
}
```

## Key Principles

1. **Event publishing is optional for persistence**: Events are saved to outbox, not required to succeed
2. **Events are async**: Published to outbox immediately, consumed asynchronously
3. **Idempotency is enforced**: Same event ID won't be processed twice
4. **No distributed transactions**: Events are eventually consistent
5. **Graceful degradation**: Missing Kafka doesn't break the application

## Testing

To test event publishing locally:

```bash
# 1. Start Kafka broker
docker run -d --name kafka -p 9092:9092 -e KAFKA_AUTO_CREATE_TOPICS_ENABLE=true confluentinc/cp-kafka

# 2. Run application with Kafka enabled
KAFKA_ENABLED=true KAFKA_BROKERS=localhost:9092 ./app

# 3. Trigger an event (create exam, start exam, etc.)

# 4. Check consumer logs for event processing
# Look for: "Exam created event: examID=..."
```

## Monitoring

Key metrics to track:
- Consumer lag: `processed_events` count vs current offset
- Event processing latency: Time from publish to processed
- Error rate: Failed message counts
- Outbox pending count: Events awaiting processing
