-- +goose Up
-- +goose StatementBegin
-- Create processed_events table for Kafka idempotent consumers
CREATE TABLE IF NOT EXISTS processed_events (
    id             BIGSERIAL PRIMARY KEY,
    event_id       TEXT NOT NULL,
    consumer_group TEXT NOT NULL,
    processed_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(event_id, consumer_group)
);

CREATE INDEX idx_processed_events_event_id ON processed_events(event_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS processed_events;
-- +goose StatementEnd
