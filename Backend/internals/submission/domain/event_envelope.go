package domain

import (
	"encoding/json"
	"time"

	"backend/pkgs/messaging"
	"github.com/google/uuid"
)

const (
	AggregateTypeSubmission = "submission"
	EventVersionV1          = 1
	EventSourceBackend      = "backend"

	// Event types for submission domain
	EventTypeSubmissionCreated  = "submission.created"
	EventTypeSubmissionGraded   = "submission.graded"
	EventTypeSubmissionRejected = "submission.rejected"
	EventTypeSubmissionAccepted = "submission.accepted"
)

type SubmissionEventPayload struct {
	SubmissionID int64     `json:"submissionId"`
	ExamID       int64     `json:"examId"`
	UserID       int64     `json:"userId"`
	Status       string    `json:"status,omitempty"`
	Score        float64   `json:"score,omitempty"`
	MaxScore     float64   `json:"maxScore,omitempty"`
	Feedback     string    `json:"feedback,omitempty"`
	GradedBy     int64     `json:"gradedBy,omitempty"`
	SubmittedAt  time.Time `json:"submittedAt,omitempty"`
	GradedAt     time.Time `json:"gradedAt,omitempty"`
}

func NewSubmissionEventEnvelope(eventType string, submissionID int64, payload SubmissionEventPayload, correlationID string) []byte {
	payloadBytes, _ := json.Marshal(payload)

	envelope := messaging.EventEnvelope{
		EventID:       uuid.NewString(),
		CorrelationID: correlationID,
		EventType:     eventType,
		Version:       EventVersionV1,
		AggregateType: AggregateTypeSubmission,
		AggregateID:   submissionID,
		OccurredAt:    time.Now().UTC(),
		Source:        EventSourceBackend,
		Payload:       payloadBytes,
	}

	envelopeBytes, _ := json.Marshal(envelope)
	return envelopeBytes
}
