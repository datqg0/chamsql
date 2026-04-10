package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"backend/db"
	submission_domain "backend/internals/submission/domain"
	"backend/internals/submission/usecase"
	"backend/pkgs/kafka"
	"backend/pkgs/logger"
	kafka_config "backend/pkgs/messaging/kafka"
)

// GradingConsumer consumes student submissions and grades them
type GradingConsumer struct {
	kafkaClient    kafka.IKafka
	database       *db.Database
	gradingService usecase.IGradingService
}

// NewGradingConsumer creates a new grading consumer
func NewGradingConsumer(
	kafkaClient kafka.IKafka,
	database *db.Database,
	gradingService usecase.IGradingService,
) *GradingConsumer {
	return &GradingConsumer{
		kafkaClient:    kafkaClient,
		database:       database,
		gradingService: gradingService,
	}
}

// Start begins consuming student submission messages
func (c *GradingConsumer) Start(ctx context.Context) {
	if c.kafkaClient == nil {
		logger.Info("Grading Kafka consumer skipped: Kafka not available")
		return
	}

	logger.Info("Starting Grading Consumer...")

	consumer := c.kafkaClient.NewConsumer(
		kafka_config.TopicStudentSubmission,
		kafka_config.GroupGradingWorkers,
		c.handleMessage,
	)

	if err := consumer.Start(ctx); err != nil && err != context.Canceled {
		logger.Error("GradingConsumer stopped with error: %v", err)
	} else {
		logger.Info("GradingConsumer stopped gracefully")
	}
}

// handleMessage processes a submission message
func (c *GradingConsumer) handleMessage(ctx context.Context, msg kafka.Message) error {
	// Unmarshal student submission
	var submission submission_domain.StudentSubmissionRequest
	if err := json.Unmarshal(msg.Value, &submission); err != nil {
		logger.Error("Failed to unmarshal submission: %v", err)
		return nil // Don't retry on unmarshal errors
	}

	logger.Debug(
		"Processing submission: submissionID=%d, studentID=%d, problemID=%d",
		submission.SubmissionID,
		submission.StudentID,
		submission.ProblemID,
	)

	// Check idempotency
	if isAlreadyProcessed(ctx, c.database, fmt.Sprintf("submission-%d", submission.SubmissionID)) {
		logger.Debug("Submission already graded: %d", submission.SubmissionID)
		return nil
	}

	// Grade the submission
	result, err := c.gradingService.Grade(ctx, &submission)
	if err != nil {
		logger.Error("Grading failed for submission %d: %v", submission.SubmissionID, err)
		// Still mark as processed to avoid reprocessing
		_ = markAsProcessed(ctx, c.database, fmt.Sprintf("submission-%d", submission.SubmissionID), kafka_config.GroupGradingWorkers)
		return nil
	}

	// Publish grading result
	if err := c.publishGradingResult(ctx, result); err != nil {
		logger.Error("Failed to publish grading result for submission %d: %v", submission.SubmissionID, err)
		// Don't mark as processed yet - retry publishing
		return err
	}

	// Mark as processed
	if err := markAsProcessed(ctx, c.database, fmt.Sprintf("submission-%d", submission.SubmissionID), kafka_config.GroupGradingWorkers); err != nil {
		logger.Error("Failed to mark submission as processed: %v", err)
	}

	logger.Info(
		"Submission graded: submissionID=%d, score=%d/%d",
		result.SubmissionID,
		result.PassedTests,
		result.TotalTests,
	)

	return nil
}

// publishGradingResult publishes the grading result to Kafka
func (c *GradingConsumer) publishGradingResult(ctx context.Context, result *submission_domain.GradingResult) error {
	if c.kafkaClient == nil {
		return fmt.Errorf("kafka client not available")
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal grading result: %w", err)
	}

	producer := c.kafkaClient.NewProducer(kafka_config.TopicSubmissionGraded)
	msg := kafka.NewRawMessage(fmt.Sprintf("submission-%d", result.SubmissionID), resultJSON, nil)
	err = producer.Publish(ctx, msg)
	if err != nil {
		_ = producer.Close()
		return fmt.Errorf("failed to publish grading result: %w", err)
	}

	_ = producer.Close()
	return nil
}
