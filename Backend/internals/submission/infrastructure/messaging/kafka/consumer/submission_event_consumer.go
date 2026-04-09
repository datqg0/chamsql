package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"backend/db"
	submission_domain "backend/internals/submission/domain"
	"backend/pkgs/kafka"
	"backend/pkgs/logger"
	"backend/pkgs/messaging"
	kafka_config "backend/pkgs/messaging/kafka"
)

type SubmissionEventConsumer struct {
	kafkaClient kafka.IKafka
	database    *db.Database
}

func NewSubmissionEventConsumer(kafkaClient kafka.IKafka, database *db.Database) *SubmissionEventConsumer {
	return &SubmissionEventConsumer{
		kafkaClient: kafkaClient,
		database:    database,
	}
}

func (c *SubmissionEventConsumer) Start(ctx context.Context) {
	if c.kafkaClient == nil {
		logger.Info("Submission Kafka consumer skipped: Kafka not available")
		return
	}

	logger.Info("Starting Submission Event Consumer...")

	consumer := c.kafkaClient.NewConsumer(
		kafka_config.TopicSubmissionEvents,
		kafka_config.GroupSubmissionWorkers,
		c.handleMessage,
	)

	if err := consumer.Start(ctx); err != nil && err != context.Canceled {
		logger.Error("SubmissionEventConsumer stopped with error: %v", err)
	} else {
		logger.Info("SubmissionEventConsumer stopped gracefully")
	}
}

func (c *SubmissionEventConsumer) handleMessage(ctx context.Context, msg kafka.Message) error {
	// Unmarshal the event envelope
	var envelope messaging.EventEnvelope
	if err := json.Unmarshal(msg.Value, &envelope); err != nil {
		logger.Error("Failed to unmarshal Submission event: %v", err)
		return nil // Don't retry on unmarshal errors
	}

	// Check idempotency - ensure we haven't processed this event before
	if isAlreadyProcessed(ctx, c.database, envelope.EventID) {
		logger.Debug("Submission event already processed: %s", envelope.EventID)
		return nil
	}

	// Handle the event based on its type
	switch envelope.EventType {
	case submission_domain.EventTypeSubmissionCreated:
		c.handleSubmissionCreated(ctx, &envelope)
	case submission_domain.EventTypeSubmissionGraded:
		c.handleSubmissionGraded(ctx, &envelope)
	case submission_domain.EventTypeSubmissionRejected:
		c.handleSubmissionRejected(ctx, &envelope)
	case submission_domain.EventTypeSubmissionAccepted:
		c.handleSubmissionAccepted(ctx, &envelope)
	default:
		logger.Warn("Unknown Submission event type: %s", envelope.EventType)
	}

	// Mark event as processed
	if err := markAsProcessed(ctx, c.database, envelope.EventID, kafka_config.GroupSubmissionWorkers); err != nil {
		logger.Error("Failed to mark Submission event as processed: %v", err)
	}

	return nil
}

func (c *SubmissionEventConsumer) handleSubmissionCreated(ctx context.Context, envelope *messaging.EventEnvelope) {
	var payload submission_domain.SubmissionEventPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		logger.Error("Failed to unmarshal SubmissionCreated payload: %v", err)
		return
	}

	logger.Info("Submission created event: submissionID=%d, examID=%d, userID=%d", payload.SubmissionID, payload.ExamID, payload.UserID)
	// TODO: Implement submission creation side effects
	// Examples:
	// - Notify graders
	// - Update statistics
	// - Trigger workflows
}

func (c *SubmissionEventConsumer) handleSubmissionGraded(ctx context.Context, envelope *messaging.EventEnvelope) {
	var payload submission_domain.SubmissionEventPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		logger.Error("Failed to unmarshal SubmissionGraded payload: %v", err)
		return
	}

	logger.Info("Submission graded event: submissionID=%d, score=%.2f, gradedBy=%d", payload.SubmissionID, payload.Score, payload.GradedBy)
	// TODO: Implement submission graded side effects
	// Examples:
	// - Send notifications to student
	// - Update leaderboards
	// - Trigger downstream workflows
}

func (c *SubmissionEventConsumer) handleSubmissionRejected(ctx context.Context, envelope *messaging.EventEnvelope) {
	var payload submission_domain.SubmissionEventPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		logger.Error("Failed to unmarshal SubmissionRejected payload: %v", err)
		return
	}

	logger.Info("Submission rejected event: submissionID=%d, reason=%s", payload.SubmissionID, payload.Feedback)
	// TODO: Implement submission rejection side effects
}

func (c *SubmissionEventConsumer) handleSubmissionAccepted(ctx context.Context, envelope *messaging.EventEnvelope) {
	var payload submission_domain.SubmissionEventPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		logger.Error("Failed to unmarshal SubmissionAccepted payload: %v", err)
		return
	}

	logger.Info("Submission accepted event: submissionID=%d", payload.SubmissionID)
	// TODO: Implement submission acceptance side effects
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
