-- +goose Up
-- Create outbox_events table for transactional event publishing
CREATE TABLE outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',  -- pending, published, failed
    retry_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP,
    error_message TEXT,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_outbox_status ON outbox_events(status);
CREATE INDEX idx_outbox_topic ON outbox_events(topic);
CREATE INDEX idx_outbox_created_at ON outbox_events(created_at);
CREATE INDEX idx_outbox_status_created ON outbox_events(status, created_at);

-- Create processed_events table for idempotent consumer tracking
CREATE TABLE processed_events (
    event_id VARCHAR(255) PRIMARY KEY,
    consumer_group VARCHAR(255) NOT NULL,
    processed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_processed_consumer_group ON processed_events(consumer_group);

-- +goose Down
DROP TABLE IF EXISTS processed_events;
DROP TABLE IF EXISTS outbox_events;
