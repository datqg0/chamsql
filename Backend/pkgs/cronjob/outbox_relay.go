package cronjob

import (
    "context"
    "backend/db"
    "backend/pkgs/kafka"
    "backend/pkgs/logger"
    "backend/sql/models"
)

type OutboxRelayTask struct {
    queries     *models.Queries
    kafkaClient kafka.IKafka
}

func NewOutboxRelayTask(database *db.Database, kafkaClient kafka.IKafka) *OutboxRelayTask {
    return &OutboxRelayTask{
        queries:     models.New(database.GetPool()),
        kafkaClient: kafkaClient,
    }
}

func (t *OutboxRelayTask) Name() string { return "outbox-relay" }

func (t *OutboxRelayTask) Execute(ctx context.Context) error {
    if t.kafkaClient == nil {
        return nil // Kafka không có, bỏ qua
    }

    events, err := t.queries.FetchPendingEvents(ctx, 50) // xử lý 50 events mỗi lần
    if err != nil {
        return err
    }
    if len(events) == 0 {
        return nil
    }

    for _, event := range events {
        producer := t.kafkaClient.NewProducer(event.Topic)
        msg := kafka.NewRawMessage(event.ID.String(), event.Payload, nil)

        if err := producer.Publish(ctx, msg); err != nil {
            logger.Error("OutboxRelay: failed topic=%s id=%s: %v", event.Topic, event.ID, err)
            _ = t.queries.MarkEventFailed(ctx, event.ID)
        } else {
            _ = t.queries.MarkEventPublished(ctx, event.ID)
            logger.Debug("OutboxRelay: published topic=%s id=%s", event.Topic, event.ID)
        }
        _ = producer.Close()
    }

    // Kiểm tra và cảnh báo nếu có event bị stuck (retry >= 3)
    stuckCount, _ := t.queries.CountStuckEvents(ctx)
    if stuckCount > 0 {
        logger.Warn("OutboxRelay: %d events are stuck (retry >= 3). Check Kafka connection or payload format.", stuckCount)
    }

    return nil
}
