package producer

import (
	"context"

	"backend/pkgs/kafka"
	"backend/pkgs/logger"
)

type KafkaProducerService interface {
	PublishEvent(ctx context.Context, topic string, key []byte, payload []byte, headers map[string]string) error
}

type kafkaProducerService struct {
	client kafka.IKafka
}

func NewKafkaProducerService(client kafka.IKafka) KafkaProducerService {
	return &kafkaProducerService{client: client}
}

func (s *kafkaProducerService) PublishEvent(ctx context.Context, topic string, key []byte, payload []byte, headers map[string]string) error {
	if s.client == nil {
		logger.Warn("KafkaProducerService: Kafka not available, skipping event for topic %s", topic)
		return nil
	}

	p := s.client.NewProducer(topic)
	defer p.Close()

	msg := kafka.Message{
		Key:     key,
		Value:   payload,
		Headers: headers,
	}

	if err := p.Publish(ctx, msg); err != nil {
		logger.Error("KafkaProducerService: failed to publish event to %s: %v", topic, err)
		return err
	}

	logger.Info("KafkaProducerService: published event to %s", topic)
	return nil
}
