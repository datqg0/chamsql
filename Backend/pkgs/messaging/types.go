package messaging

import (
	"encoding/json"
	"time"
)

type EventEnvelope struct {
	EventID       string          `json:"eventId"`
	CorrelationID string          `json:"correlationId,omitempty"`
	EventType     string          `json:"eventType"`
	Version       int             `json:"version"`
	AggregateType string          `json:"aggregateType"`
	AggregateID   int64           `json:"aggregateId"`
	OccurredAt    time.Time       `json:"occurredAt"`
	Source        string          `json:"source"`
	Payload       json.RawMessage `json:"payload"`
}
