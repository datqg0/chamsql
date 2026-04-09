package domain

import (
	"encoding/json"
	"time"

	"backend/pkgs/messaging"
	"github.com/google/uuid"
)

const (
	AggregateTypeExam  = "exam"
	EventVersionV1     = 1
	EventSourceBackend = "backend"

	// Event types for exam domain
	EventTypeExamCreated   = "exam.created"
	EventTypeExamStarted   = "exam.started"
	EventTypeExamSubmitted = "exam.submitted"
	EventTypeExamFinished  = "exam.finished"
	EventTypeExamCancelled = "exam.cancelled"
)

type ExamEventPayload struct {
	ExamID          int64     `json:"examId"`
	UserID          int64     `json:"userId,omitempty"`
	Title           string    `json:"title"`
	CreatedBy       int64     `json:"createdBy,omitempty"`
	Status          string    `json:"status,omitempty"`
	StartTime       time.Time `json:"startTime,omitempty"`
	EndTime         time.Time `json:"endTime,omitempty"`
	DurationMinutes int32     `json:"durationMinutes,omitempty"`
	Score           float64   `json:"score,omitempty"`
	MaxScore        float64   `json:"maxScore,omitempty"`
}

func NewExamEventEnvelope(eventType string, examID int64, payload ExamEventPayload, correlationID string) []byte {
	payloadBytes, _ := json.Marshal(payload)

	envelope := messaging.EventEnvelope{
		EventID:       uuid.NewString(),
		CorrelationID: correlationID,
		EventType:     eventType,
		Version:       EventVersionV1,
		AggregateType: AggregateTypeExam,
		AggregateID:   examID,
		OccurredAt:    time.Now().UTC(),
		Source:        EventSourceBackend,
		Payload:       payloadBytes,
	}

	envelopeBytes, _ := json.Marshal(envelope)
	return envelopeBytes
}
