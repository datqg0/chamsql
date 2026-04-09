# Event Handlers & Side Effects Implementation Guide

## Overview

The Kafka event infrastructure is now in place for publishing and consuming events. This guide shows how to implement event handlers for side effects like notifications, leaderboard updates, and analytics.

## Current Architecture

```
┌─────────────────┐
│  Use Case       │  (Create exam, Submit answer)
└────────┬────────┘
         │ PublishEvent()
         ▼
┌─────────────────────┐
│  Outbox Repository  │  (Save to outbox_events)
└────────┬────────────┘
         │ (async)
         ▼
┌─────────────────────┐
│  Outbox Processor   │  (Poll & batch publish)
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│  Kafka Broker       │  (Store in partition)
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│  Event Consumer     │  (Ready to consume)
└──────────┬──────────┘
           │
           ▼
      ┌─────────────────────┐
      │  Event Handler      │  ◄─ IMPLEMENT HERE
      │  (Side effects)     │
      └─────────────────────┘
```

## What Needs to Be Done

### 1. Start Outbox Processor in main.go

The outbox processor needs to be started as a goroutine to poll and publish pending events.

**File**: `cmd/app/main.go`

**Implementation Pattern**:
```go
import "backend/pkgs/messaging/outbox"

// In setupRoutes or after all dependencies are initialized
func setupEventProcessing(ctx context.Context, container *di.Container) {
    container.Invoke(func(
        db *db.Database,
        kafkaClient kafka.IKafka,
    ) {
        // Create outbox processor
        outboxRepo := &YourOutboxRepositoryImpl{db: db, queries: models.New(db.GetPool())}
        processor := outbox.NewProcessor(outboxRepo, nil, kafkaClient)
        
        // Start in goroutine
        go processor.Start(ctx)
        logger.Info("Outbox processor started")
    })
}
```

**Why It's Important**:
- Polls `outbox_events` table every 5 seconds
- Publishes pending events to Kafka
- Retries failed events with backoff
- Without this, events stay in the database and never reach Kafka consumers

### 2. Implement Event Handlers for Exam Events

**Location**: `internals/exam/infrastructure/messaging/kafka/consumer/handlers/`

**Example - Exam Creation Handler**:

```go
// File: internals/exam/infrastructure/messaging/kafka/consumer/handlers/exam_created_handler.go
package handlers

import (
    "context"
    "encoding/json"
    
    "backend/internals/exam/domain"
    "backend/pkgs/logger"
)

type ExamCreatedEventHandler struct {
    // Inject dependencies needed for side effects
    // notificationService NotificationService
    // analyticsService    AnalyticsService
}

func NewExamCreatedEventHandler() *ExamCreatedEventHandler {
    return &ExamCreatedEventHandler{}
}

func (h *ExamCreatedEventHandler) Handle(ctx context.Context, payload []byte) error {
    var eventPayload domain.ExamEventPayload
    if err := json.Unmarshal(payload, &eventPayload); err != nil {
        logger.Error("Failed to unmarshal exam created event: %v", err)
        return err
    }
    
    logger.Info("Handling exam.created event for exam %d", eventPayload.ExamID)
    
    // Side Effects:
    
    // 1. Send notification to students
    // h.notificationService.NotifyExamCreated(ctx, eventPayload.ExamID, eventPayload.Title)
    
    // 2. Log analytics event
    // h.analyticsService.TrackEvent(ctx, "exam_created", map[string]interface{}{
    //     "examId": eventPayload.ExamID,
    //     "title": eventPayload.Title,
    // })
    
    // 3. Update search index / cache
    // h.cacheService.InvalidateExamCache()
    
    // 4. Update read model for dashboards
    // h.readModelRepository.CreateExamReadModel(ctx, eventPayload)
    
    return nil
}
```

**Example - Exam Started Handler**:

```go
// File: internals/exam/infrastructure/messaging/kafka/consumer/handlers/exam_started_handler.go
package handlers

import (
    "context"
    "encoding/json"
    
    "backend/internals/exam/domain"
    "backend/pkgs/logger"
)

type ExamStartedEventHandler struct {
    // notificationService NotificationService
}

func NewExamStartedEventHandler() *ExamStartedEventHandler {
    return &ExamStartedEventHandler{}
}

func (h *ExamStartedEventHandler) Handle(ctx context.Context, payload []byte) error {
    var eventPayload domain.ExamEventPayload
    if err := json.Unmarshal(payload, &eventPayload); err != nil {
        logger.Error("Failed to unmarshal exam started event: %v", err)
        return err
    }
    
    logger.Info("User %d started exam %d", eventPayload.UserID, eventPayload.ExamID)
    
    // Side Effects:
    
    // 1. Update participant status in cache
    // h.cacheService.UpdateParticipantStatus(eventPayload.ExamID, eventPayload.UserID, "in_progress")
    
    // 2. Start exam timeout tracking
    // h.timerService.StartExamTimer(ctx, eventPayload.ExamID, eventPayload.UserID, eventPayload.DurationMinutes)
    
    // 3. Log activity
    // h.auditService.LogActivity("exam_started", eventPayload.UserID, eventPayload.ExamID)
    
    return nil
}
```

**Example - Exam Finished Handler**:

```go
// File: internals/exam/infrastructure/messaging/kafka/consumer/handlers/exam_finished_handler.go
package handlers

import (
    "context"
    "encoding/json"
    
    "backend/internals/exam/domain"
    "backend/pkgs/logger"
)

type ExamFinishedEventHandler struct {
    // leaderboardService  LeaderboardService
    // notificationService NotificationService
}

func NewExamFinishedEventHandler() *ExamFinishedEventHandler {
    return &ExamFinishedEventHandler{}
}

func (h *ExamFinishedEventHandler) Handle(ctx context.Context, payload []byte) error {
    var eventPayload domain.ExamEventPayload
    if err := json.Unmarshal(payload, &eventPayload); err != nil {
        logger.Error("Failed to unmarshal exam finished event: %v", err)
        return err
    }
    
    logger.Info("User %d finished exam %d with score %.2f", 
        eventPayload.UserID, eventPayload.ExamID, eventPayload.Score)
    
    // Side Effects:
    
    // 1. Update leaderboard
    // h.leaderboardService.UpdateScore(ctx, eventPayload.UserID, eventPayload.Score)
    
    // 2. Send completion notification
    // h.notificationService.NotifyExamCompleted(ctx, eventPayload.UserID, eventPayload.Score)
    
    // 3. Trigger badge/achievement system
    // h.achievementService.CheckAchievements(ctx, eventPayload.UserID, eventPayload.Score)
    
    // 4. Update analytics
    // h.analyticsService.RecordExamCompletion(ctx, eventPayload.ExamID, eventPayload.Score)
    
    // 5. Update cache
    // h.cacheService.InvalidateUserStats(eventPayload.UserID)
    
    return nil
}
```

### 3. Implement Event Handlers for Submission Events

**Location**: `internals/submission/infrastructure/messaging/kafka/consumer/handlers/`

**Example - Submission Graded Handler**:

```go
// File: internals/submission/infrastructure/messaging/kafka/consumer/handlers/submission_graded_handler.go
package handlers

import (
    "context"
    "encoding/json"
    
    "backend/internals/submission/domain"
    "backend/pkgs/logger"
)

type SubmissionGradedEventHandler struct {
    // notificationService NotificationService
    // achievementService  AchievementService
}

func NewSubmissionGradedEventHandler() *SubmissionGradedEventHandler {
    return &SubmissionGradedEventHandler{}
}

func (h *SubmissionGradedEventHandler) Handle(ctx context.Context, payload []byte) error {
    var eventPayload domain.SubmissionEventPayload
    if err := json.Unmarshal(payload, &eventPayload); err != nil {
        logger.Error("Failed to unmarshal submission graded event: %v", err)
        return err
    }
    
    logger.Info("Submission %d graded with score %.2f for user %d", 
        eventPayload.SubmissionID, eventPayload.Score, eventPayload.UserID)
    
    // Side Effects:
    
    // 1. Send notification
    // if eventPayload.Score >= 8.0 {
    //     h.notificationService.NotifyAccepted(ctx, eventPayload.UserID, eventPayload.Score)
    // } else {
    //     h.notificationService.NotifyRejected(ctx, eventPayload.UserID, eventPayload.Score)
    // }
    
    // 2. Update user stats
    // h.userStatsService.UpdateStats(ctx, eventPayload.UserID)
    
    // 3. Check for problem mastery
    // h.masteryService.CheckMastery(ctx, eventPayload.UserID)
    
    // 4. Update leaderboard
    // h.leaderboardService.UpdateScore(ctx, eventPayload.UserID, eventPayload.Score)
    
    return nil
}
```

### 4. Register Handlers in Consumer

**File**: `internals/exam/infrastructure/messaging/kafka/consumer/exam_event_consumer.go`

Update the `handleMessage` method to use handlers:

```go
func (c *ExamEventConsumer) handleMessage(ctx context.Context, envelope *messaging.EventEnvelope) error {
    // Check idempotency
    processed, err := c.queries.IsEventProcessed(ctx, models.IsEventProcessedParams{
        EventID:       envelope.EventID,
        ConsumerGroup: c.groupID,
    })
    
    if err == nil && processed {
        logger.Debug("Event %s already processed, skipping", envelope.EventID)
        return nil
    }
    
    // Dispatch to handlers based on event type
    var handler ExamEventHandler
    switch envelope.EventType {
    case domain.EventTypeExamCreated:
        handler = handlers.NewExamCreatedEventHandler()
    case domain.EventTypeExamStarted:
        handler = handlers.NewExamStartedEventHandler()
    case domain.EventTypeExamFinished:
        handler = handlers.NewExamFinishedEventHandler()
    default:
        logger.Warn("Unknown event type: %s", envelope.EventType)
        return nil
    }
    
    // Handle event
    if err := handler.Handle(ctx, envelope.Payload); err != nil {
        logger.Error("Failed to handle event %s: %v", envelope.EventID, err)
        return err
    }
    
    // Mark as processed
    if err := c.queries.MarkEventProcessed(ctx, models.MarkEventProcessedParams{
        EventID:       envelope.EventID,
        ConsumerGroup: c.groupID,
    }); err != nil {
        logger.Error("Failed to mark event %s as processed: %v", envelope.EventID, err)
    }
    
    return nil
}

// Handler interface
type ExamEventHandler interface {
    Handle(ctx context.Context, payload []byte) error
}
```

## Dependency Injection for Handlers

Create a factory function to inject dependencies:

```go
// File: internals/exam/infrastructure/messaging/kafka/consumer/handlers/factory.go
package handlers

import "backend/internals/exam/repository"

func NewExamEventHandlers(
    notificationService NotificationService,
    analyticsService AnalyticsService,
) map[string]ExamEventHandler {
    return map[string]ExamEventHandler{
        "exam.created":  NewExamCreatedEventHandler(notificationService, analyticsService),
        "exam.started":  NewExamStartedEventHandler(notificationService),
        "exam.finished": NewExamFinishedEventHandler(analyticsService),
    }
}
```

## Testing Event Handlers

### Unit Tests

```go
// File: internals/exam/infrastructure/messaging/kafka/consumer/handlers/exam_created_handler_test.go
package handlers

import (
    "context"
    "encoding/json"
    "testing"
    
    "backend/internals/exam/domain"
    "github.com/stretchr/testify/assert"
)

func TestExamCreatedEventHandler(t *testing.T) {
    handler := NewExamCreatedEventHandler()
    
    payload := domain.ExamEventPayload{
        ExamID:    1,
        Title:     "Test Exam",
        CreatedBy: 100,
    }
    
    payloadBytes, _ := json.Marshal(payload)
    err := handler.Handle(context.Background(), payloadBytes)
    
    assert.NoError(t, err)
}
```

### Integration Tests

```go
// File: internals/exam/infrastructure/messaging/kafka/consumer/exam_event_consumer_integration_test.go
package consumer

import (
    "context"
    "testing"
    
    "backend/internals/exam/domain"
    "github.com/stretchr/testify/assert"
)

func TestExamEventConsumerIntegration(t *testing.T) {
    // Setup consumer with mock Kafka
    // Publish test event
    // Verify handler was called
    // Verify side effects occurred
}
```

## Performance Considerations

1. **Batch Processing**: Process multiple events before committing offset
2. **Async Side Effects**: Use goroutines for non-critical side effects
3. **Circuit Breaker**: Implement circuit breaker for external service calls
4. **Dead Letter Queue**: Route handler errors to DLQ for manual inspection
5. **Monitoring**: Track handler execution time and error rates

## Error Handling Strategy

```go
// With retry and circuit breaker
type ResilientEventHandler struct {
    handler EventHandler
    retries int
    breaker CircuitBreaker
}

func (r *ResilientEventHandler) Handle(ctx context.Context, payload []byte) error {
    if r.breaker.IsOpen() {
        return fmt.Errorf("circuit breaker open for %s", r.handler.Name())
    }
    
    for attempt := 0; attempt < r.retries; attempt++ {
        err := r.handler.Handle(ctx, payload)
        if err == nil {
            r.breaker.RecordSuccess()
            return nil
        }
        
        r.breaker.RecordFailure()
        if attempt < r.retries-1 {
            time.Sleep(exponentialBackoff(attempt))
        }
    }
    
    return fmt.Errorf("handler failed after %d retries", r.retries)
}
```

## Summary

The event infrastructure is production-ready. The remaining work is implementing:

1. **Outbox Processor Startup** - Start in main.go (2-3 lines)
2. **Event Handlers** - Business logic for each event type
3. **Dependency Injection** - Wire handlers with services
4. **Side Effects** - Notifications, analytics, leaderboards, etc.
5. **Monitoring & Testing** - Metrics and integration tests

Each of these is independent and can be implemented gradually without impacting other parts of the system.
