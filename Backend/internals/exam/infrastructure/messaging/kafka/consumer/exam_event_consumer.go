package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"backend/db"
	exam_domain "backend/internals/exam/domain"
	"backend/pkgs/kafka"
	"backend/pkgs/logger"
	"backend/pkgs/messaging"
	kafka_config "backend/pkgs/messaging/kafka"
)

type ExamEventConsumer struct {
	kafkaClient kafka.IKafka
	database    *db.Database
}

func NewExamEventConsumer(kafkaClient kafka.IKafka, database *db.Database) *ExamEventConsumer {
	return &ExamEventConsumer{
		kafkaClient: kafkaClient,
		database:    database,
	}
}

func (c *ExamEventConsumer) Start(ctx context.Context) {
	if c.kafkaClient == nil {
		logger.Info("Exam Kafka consumer skipped: Kafka not available")
		return
	}

	logger.Info("Starting Exam Event Consumer...")

	consumer := c.kafkaClient.NewConsumer(
		kafka_config.TopicExamEvents,
		kafka_config.GroupExamWorkers,
		c.handleMessage,
	)

	if err := consumer.Start(ctx); err != nil && err != context.Canceled {
		logger.Error("ExamEventConsumer stopped with error: %v", err)
	} else {
		logger.Info("ExamEventConsumer stopped gracefully")
	}
}

func (c *ExamEventConsumer) handleMessage(ctx context.Context, msg kafka.Message) error {
	// Unmarshal the event envelope
	var envelope messaging.EventEnvelope
	if err := json.Unmarshal(msg.Value, &envelope); err != nil {
		logger.Error("Failed to unmarshal Exam event: %v", err)
		return nil // Don't retry on unmarshal errors
	}

	// Check idempotency - ensure we haven't processed this event before
	if isAlreadyProcessed(ctx, c.database, envelope.EventID) {
		logger.Debug("Exam event already processed: %s", envelope.EventID)
		return nil
	}

	// Handle the event based on its type
	switch envelope.EventType {
	case exam_domain.EventTypeExamCreated:
		c.handleExamCreated(ctx, &envelope)
	case exam_domain.EventTypeExamStarted:
		c.handleExamStarted(ctx, &envelope)
	case exam_domain.EventTypeExamSubmitted:
		c.handleExamSubmitted(ctx, &envelope)
	case exam_domain.EventTypeExamFinished:
		c.handleExamFinished(ctx, &envelope)
	case exam_domain.EventTypeExamTimeExpired:
		c.handleExamTimeExpired(ctx, &envelope)
	case exam_domain.EventTypeExamTimeExtended:
		c.handleExamTimeExtended(ctx, &envelope)
	default:
		logger.Warn("Unknown Exam event type: %s", envelope.EventType)
	}

	// Mark event as processed
	if err := markAsProcessed(ctx, c.database, envelope.EventID, kafka_config.GroupExamWorkers); err != nil {
		logger.Error("Failed to mark Exam event as processed: %v", err)
	}

	return nil
}

func (c *ExamEventConsumer) handleExamCreated(ctx context.Context, envelope *messaging.EventEnvelope) {
	var payload exam_domain.ExamEventPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		logger.Error("Failed to unmarshal ExamCreated payload: %v", err)
		return
	}

	logger.Info("Exam created event: examID=%d, title=%s, createdBy=%d", payload.ExamID, payload.Title, payload.CreatedBy)
	// TODO: Implement exam creation side effects
	// Examples:
	// - Update search indexes
	// - Send notifications
	// - Trigger related workflows
}

func (c *ExamEventConsumer) handleExamStarted(ctx context.Context, envelope *messaging.EventEnvelope) {
	var payload exam_domain.ExamEventPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		logger.Error("Failed to unmarshal ExamStarted payload: %v", err)
		return
	}

	logger.Info("Exam started event: examID=%d, userID=%d", payload.ExamID, payload.UserID)
	// TODO: Implement exam start side effects
}

func (c *ExamEventConsumer) handleExamSubmitted(ctx context.Context, envelope *messaging.EventEnvelope) {
	var payload exam_domain.ExamEventPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		logger.Error("Failed to unmarshal ExamSubmitted payload: %v", err)
		return
	}

	logger.Info("Exam submitted event: examID=%d, userID=%d", payload.ExamID, payload.UserID)
	// TODO: Implement exam submission side effects
}

func (c *ExamEventConsumer) handleExamFinished(ctx context.Context, envelope *messaging.EventEnvelope) {
	var payload exam_domain.ExamEventPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		logger.Error("Failed to unmarshal ExamFinished payload: %v", err)
		return
	}

	logger.Info("Exam finished event: examID=%d, userID=%d, score=%.2f", payload.ExamID, payload.UserID, payload.Score)
	// TODO: Implement exam finish side effects
	// Examples:
	// - Send result notifications
	// - Update user statistics
	// - Trigger grading workflows
}

func (c *ExamEventConsumer) handleExamTimeExpired(ctx context.Context, envelope *messaging.EventEnvelope) {
	var payload exam_domain.ExamEventPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		logger.Error("Failed to unmarshal ExamTimeExpired payload: %v", err)
		return
	}

	logger.Info("Exam time expired event: examID=%d, endTime=%v", payload.ExamID, payload.EndTime)

	// Update exam status to "completed" to lock it
	pool := c.database.GetPool()
	result, err := pool.Exec(
		ctx,
		"UPDATE exams SET status = 'completed', updated_at = NOW() WHERE id = $1 AND status != 'completed'",
		payload.ExamID,
	)
	if err != nil {
		logger.Error("Failed to mark exam as completed: examID=%d, error=%v", payload.ExamID, err)
		return
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected > 0 {
		logger.Info("Exam marked as completed: examID=%d", payload.ExamID)
	}

	// TODO: Additional side effects
	// - Auto-submit incomplete participants
	// - Send notifications to students
	// - Update statistics
}

func (c *ExamEventConsumer) handleExamTimeExtended(ctx context.Context, envelope *messaging.EventEnvelope) {
	var payload exam_domain.ExamEventPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		logger.Error("Failed to unmarshal ExamTimeExtended payload: %v", err)
		return
	}

	logger.Info("Exam time extended event: examID=%d, newEndTime=%v", payload.ExamID, payload.EndTime)

	// Update exam end_time
	pool := c.database.GetPool()
	_, err := pool.Exec(
		ctx,
		"UPDATE exams SET end_time = $2, updated_at = NOW() WHERE id = $1",
		payload.ExamID,
		payload.EndTime,
	)
	if err != nil {
		logger.Error("Failed to extend exam time: examID=%d, error=%v", payload.ExamID, err)
		return
	}

	logger.Info("Exam time extended: examID=%d, newEndTime=%v", payload.ExamID, payload.EndTime)
}

// isAlreadyProcessed checks if an event has already been processed (idempotency)
func isAlreadyProcessed(ctx context.Context, database *db.Database, eventID string) bool {
	var count int
	pool := database.GetPool()
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM processed_events WHERE event_id = $1", eventID).Scan(&count)
	if err != nil {
		logger.Error("Failed to check if event processed: %v", err)
		return false
	}
	return count > 0
}

// markAsProcessed marks an event as processed in the database
func markAsProcessed(ctx context.Context, database *db.Database, eventID string, consumerGroup string) error {
	pool := database.GetPool()
	_, err := pool.Exec(
		ctx,
		"INSERT INTO processed_events (event_id, consumer_group) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		eventID,
		consumerGroup,
	)
	if err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}
	return nil
}
