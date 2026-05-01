-- =============================================
-- OUTBOX EVENTS
-- =============================================

-- name: SaveOutboxEvent :one
INSERT INTO outbox_events (id, topic, payload, status, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, 'pending', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, topic, payload, status;

-- name: FetchPendingEvents :many
SELECT id, topic, payload
FROM outbox_events
WHERE status = 'pending' AND retry_count < 3
ORDER BY created_at ASC
LIMIT $1;

-- name: MarkEventPublished :exec
UPDATE outbox_events
SET status = 'published', published_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: MarkEventFailed :exec
UPDATE outbox_events
SET status = 'failed', retry_count = retry_count + 1, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: CountStuckEvents :one
SELECT COUNT(*) FROM outbox_events WHERE status = 'pending' AND retry_count >= 3;

-- =============================================
-- PROCESSED EVENTS (for idempotent consumers)
-- =============================================

-- name: MarkEventProcessed :exec
INSERT INTO processed_events (event_id, consumer_group, processed_at)
VALUES ($1, $2, CURRENT_TIMESTAMP)
ON CONFLICT (event_id) DO NOTHING;

-- name: IsEventProcessed :one
SELECT EXISTS(
    SELECT 1 FROM processed_events
    WHERE event_id = $1 AND consumer_group = $2
) as processed;
