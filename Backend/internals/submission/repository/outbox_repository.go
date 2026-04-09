package repository

import (
	"context"
	"fmt"

	"backend/db"
	"backend/sql/models"
)

type ISubmissionOutboxRepository interface {
	// Publish a submission event to the outbox (eventEnvelope is already serialized as JSON bytes from NewSubmissionEventEnvelope)
	PublishEvent(ctx context.Context, topic string, eventEnvelope []byte) error
}

type submissionOutboxRepository struct {
	db      *db.Database
	queries *models.Queries
}

func NewSubmissionOutboxRepository(database *db.Database) ISubmissionOutboxRepository {
	return &submissionOutboxRepository{
		db:      database,
		queries: models.New(database.GetPool()),
	}
}

func (r *submissionOutboxRepository) PublishEvent(ctx context.Context, topic string, eventEnvelope []byte) error {
	_, err := r.queries.SaveOutboxEvent(ctx, models.SaveOutboxEventParams{
		Topic:   topic,
		Payload: eventEnvelope,
	})
	if err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}

	return nil
}
