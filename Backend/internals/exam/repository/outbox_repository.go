package repository

import (
	"context"
	"fmt"

	"backend/db"
	"backend/sql/models"
)

type IExamOutboxRepository interface {
	// Publish an exam event to the outbox (eventEnvelope is already serialized as JSON bytes from NewExamEventEnvelope)
	PublishEvent(ctx context.Context, topic string, eventEnvelope []byte) error
}

type examOutboxRepository struct {
	db      *db.Database
	queries *models.Queries
}

func NewExamOutboxRepository(database *db.Database) IExamOutboxRepository {
	return &examOutboxRepository{
		db:      database,
		queries: models.New(database.GetPool()),
	}
}

func (r *examOutboxRepository) PublishEvent(ctx context.Context, topic string, eventEnvelope []byte) error {
	_, err := r.queries.SaveOutboxEvent(ctx, models.SaveOutboxEventParams{
		Topic:   topic,
		Payload: eventEnvelope,
	})
	if err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}

	return nil
}
