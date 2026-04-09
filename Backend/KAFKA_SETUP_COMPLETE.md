# Kafka Event-Driven Architecture Implementation - Setup Summary

## Completion Status: ✅ COMPLETE

This document summarizes the Kafka event-driven architecture implementation for the chamsql Backend project, based on patterns from dadpt/backend.

## What Was Implemented

### 1. **Infrastructure Setup**
- ✅ Copied Kafka packages from dadpt/backend:
  - `pkgs/kafka/` - Kafka client, producer, consumer, registry
  - `pkgs/messaging/` - Event envelopes, outbox processor, RabbitMQ integration

### 2. **Configuration**
- ✅ Updated `configs/configs.go`:
  - Added `KafkaEnabled`, `KafkaBrokers`, `KafkaClientID` fields
  - These are loaded from environment variables

- ✅ Updated `go.mod`:
  - Added `github.com/segmentio/kafka-go v0.4.47` dependency

### 3. **Topic Definition**
- ✅ Updated `pkgs/messaging/kafka/topics.go`:
  - **Exam Events Topic**: `chamsql-exam-events-v1`
    - 3 partitions (distributed by exam ID)
    - Consumer group: `chamsql-exam-workers`
  
  - **Submission Events Topic**: `chamsql-submission-events-v1`
    - 6 partitions (distributed by submission ID)
    - Consumer group: `chamsql-submission-workers`

### 4. **Event Definitions**
- ✅ Created `internals/exam/domain/event_envelope.go`:
  - Event types: `exam.created`, `exam.started`, `exam.submitted`, `exam.finished`
  - Event payload structure with exam metadata
  - Helper function: `NewExamEventEnvelope()`

- ✅ Created `internals/submission/domain/event_envelope.go`:
  - Event types: `submission.created`, `submission.graded`, `submission.rejected`, `submission.accepted`
  - Event payload structure with submission metadata
  - Helper function: `NewSubmissionEventEnvelope()`

### 5. **Dependency Injection**
- ✅ Updated `di/container.go`:
  - Added `provideKafkaRegistry()`: Initializes topic registry
  - Added `provideKafka()`: Creates Kafka client with proper error handling
  - Integrated Kafka into provider chain

### 6. **Event Consumers**
- ✅ Created `internals/exam/infrastructure/messaging/kafka/consumer/exam_event_consumer.go`:
  - Listens to exam events
  - Implements idempotent message handling (checks `processed_events` table)
  - Handlers for each event type (created, started, submitted, finished)
  - Graceful error handling and logging

- ✅ Created `internals/submission/infrastructure/messaging/kafka/consumer/submission_event_consumer.go`:
  - Listens to submission events
  - Implements idempotent message handling
  - Handlers for each event type (created, graded, rejected, accepted)
  - Graceful error handling and logging

### 7. **Database Migrations**
- ✅ Created `sql/schema/004_create_outbox_tables.sql`:
  - `outbox_events` table: Stores events for publishing with outbox pattern
    - Fields: id, topic, payload (JSONB), status, retry_count, timestamps
    - Indexes for efficient polling
  
  - `processed_events` table: Tracks processed event IDs for idempotency
    - Fields: event_id (primary key), consumer_group, processed_at
    - Index on consumer_group

### 8. **Application Startup**
- ✅ Updated `cmd/app/main.go`:
  - Initializes Kafka client and registry
  - Ensures Kafka topics exist on startup
  - Starts exam and submission event consumers as goroutines
  - Implements graceful shutdown with context cancellation
  - Proper cleanup of Kafka and database connections

## Architecture Patterns Used

### Event Envelope Standard
```json
{
  "eventId": "uuid",
  "correlationId": "uuid",
  "eventType": "exam.created",
  "version": 1,
  "aggregateType": "exam",
  "aggregateId": 123,
  "occurredAt": "2026-04-09T23:45:00Z",
  "source": "backend",
  "payload": { /* domain-specific data */ }
}
```

### Consumer Pattern
- **At-least-once semantics**: Events are processed at least once
- **Idempotency**: Each event ID is tracked in `processed_events` table
- **Error handling**: Messages are logged and marked as processed (not retried indefinitely)
- **Graceful shutdown**: Context-based cancellation of consumer loops

### Topic Partitioning
- **Exam Topic**: 3 partitions (suitable for moderate concurrency)
- **Submission Topic**: 6 partitions (higher concurrency for submissions)
- **Keying strategy**: By aggregate ID for ordering within same entity

## How to Publish Events (Next Step)

In your use cases, events should be published to the outbox:

```go
// In exam usecase
func (u *examUseCase) Create(ctx context.Context, userID int64, req *dto.CreateExamRequest) (*dto.ExamResponse, error) {
	// Create exam in database
	exam, err := u.examRepo.Create(ctx, ...)
	
	// Publish event to outbox (same transaction)
	payload := exam_domain.ExamEventPayload{
		ExamID:    exam.ID,
		Title:     exam.Title,
		CreatedBy: userID,
		// ... other fields
	}
	eventBytes := exam_domain.NewExamEventEnvelope(
		exam_domain.EventTypeExamCreated,
		exam.ID,
		payload,
		correlationID,
	)
	
	// Save to outbox (transactional)
	if err := u.outboxRepo.Save(ctx, kafka_config.TopicExamEvents, eventBytes); err != nil {
		return nil, err
	}
	
	return mapper.ToDTO(exam), nil
}
```

## Configuration

Add these environment variables:

```bash
# Kafka
KAFKA_ENABLED=true
KAFKA_BROKERS=localhost:9092,localhost:9093,localhost:9094
KAFKA_CLIENT_ID=chamsql-backend
```

## Testing Kafka Events Locally

1. **Start Kafka with Docker**:
   ```bash
   docker run -d --name kafka -p 9092:9092 \
     -e KAFKA_BROKER_ID=1 \
     -e KAFKA_ZOOKEEPER_CONNECT=localhost:2181 \
     -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
     -e KAFKA_AUTO_CREATE_TOPICS_ENABLE=true \
     confluentinc/cp-kafka:latest
   ```

2. **Verify topics are created** (in application logs on first startup)

3. **Monitor consumers**:
   ```bash
   kafka-consumer-groups --bootstrap-server localhost:9092 \
     --group chamsql-exam-workers --describe
   ```

## Next Steps

1. **Implement event publishing** in exam and submission use cases
2. **Create outbox processor** if not using RabbitMQ for publishing
3. **Add more event handlers** as business requirements evolve
4. **Implement monitoring**:
   - Track consumer lag
   - Monitor event processing errors
   - Alert on failed events
5. **Setup production Kafka cluster** with proper replication and monitoring

## Key Files

### Core Files
- `pkgs/kafka/` - Kafka infrastructure (client, producer, consumer)
- `pkgs/messaging/` - Event messaging utilities

### Configuration
- `configs/configs.go` - Kafka configuration loading
- `pkgs/messaging/kafka/topics.go` - Topic definitions
- `pkgs/messaging/kafka/register.go` - Topic registration

### Domain Events
- `internals/exam/domain/event_envelope.go` - Exam domain events
- `internals/submission/domain/event_envelope.go` - Submission domain events

### Consumers
- `internals/exam/infrastructure/messaging/kafka/consumer/exam_event_consumer.go`
- `internals/submission/infrastructure/messaging/kafka/consumer/submission_event_consumer.go`

### Database
- `sql/schema/004_create_outbox_tables.sql` - Outbox and processed events tables

### Application
- `di/container.go` - Dependency injection configuration
- `cmd/app/main.go` - Application entry point with consumer startup

## References

- Event-driven architecture pattern from dadpt/backend
- Kafka segmentio/kafka-go library
- Outbox pattern for transactional event publishing
- Idempotent consumer pattern for exactly-once semantics
