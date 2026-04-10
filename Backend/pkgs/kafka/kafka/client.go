package kafka

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"backend/pkgs/logger"

	kg "github.com/segmentio/kafka-go"
)

type IKafka interface {
	Close() error
	Brokers() []string
	EnsureTopics(ctx context.Context, topics []TopicDefinition) error
	NewProducer(topic string, opts ...ProducerOption) IProducer
	NewConsumer(topic, group string, handler MessageHandler, opts ...ConsumerOption) IConsumer
}

type client struct {
	brokers  []string
	clientID string
}

func NewKafka(cfg Config) (IKafka, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka: brokers are required")
	}

	normalized := make([]string, 0, len(cfg.Brokers))
	for _, broker := range cfg.Brokers {
		trimmed := strings.TrimSpace(broker)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}
	if len(normalized) == 0 {
		return nil, fmt.Errorf("kafka: brokers are empty after normalization")
	}

	clientID := cfg.ClientID
	if clientID == "" {
		clientID = "kafka-client"
	}

	logger.Info("Kafka client initialized: brokers=%s, clientID=%s", strings.Join(normalized, ","), clientID)
	return &client{brokers: normalized, clientID: clientID}, nil
}

func (c *client) Brokers() []string {
	return append([]string(nil), c.brokers...)
}

func (c *client) Close() error {
	logger.Info("Kafka client closed")
	return nil
}

func (c *client) EnsureTopics(ctx context.Context, topics []TopicDefinition) error {
	if len(topics) == 0 {
		return nil
	}

	kafkaTopics := make([]kg.TopicConfig, 0, len(topics))
	for _, topic := range topics {
		if !topic.Validate() {
			continue
		}
		partitions := topic.NumPartitions
		if partitions <= 0 {
			partitions = 6
		}
		replication := topic.ReplicationFactor
		if replication <= 0 {
			replication = 1
		}
		kafkaTopics = append(kafkaTopics, kg.TopicConfig{
			Topic:             topic.Name,
			NumPartitions:     partitions,
			ReplicationFactor: replication,
		})
	}

	if len(kafkaTopics) == 0 {
		return nil
	}

	const maxRetries = 5
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := c.createTopics(ctx, kafkaTopics); err != nil {
			lastErr = err
			if attempt < maxRetries {
				backoff := time.Duration(attempt) * 2 * time.Second
				logger.Warn("Kafka topics ensure attempt %d/%d failed: %v — retrying in %v", attempt, maxRetries, err, backoff)
				select {
				case <-time.After(backoff):
				case <-ctx.Done():
					return ctx.Err()
				}
				continue
			}
		} else {
			logger.Info("Kafka topics ensured: count=%d", len(kafkaTopics))
			return nil
		}
	}

	return fmt.Errorf("kafka: ensure topics after %d retries: %w", maxRetries, lastErr)
}

func (c *client) createTopics(ctx context.Context, topics []kg.TopicConfig) error {
	conn, err := kg.DialContext(ctx, "tcp", c.brokers[0])
	if err != nil {
		return fmt.Errorf("kafka: dial broker: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("kafka: get controller: %w", err)
	}

	controllerAddr := fmt.Sprintf("%s:%d", controller.Host, controller.Port)
	if !c.isReachable(controllerAddr) {
		controllerAddr = c.brokers[0]
	}

	controllerConn, err := kg.DialContext(ctx, "tcp", controllerAddr)
	if err != nil {
		return fmt.Errorf("kafka: dial controller: %w", err)
	}
	defer controllerConn.Close()

	if err := controllerConn.CreateTopics(topics...); err != nil {
		if strings.Contains(err.Error(), "Topic with this name already exists") {
			logger.Info("Kafka topics already exist, skip create")
			return nil
		}
		return fmt.Errorf("kafka: create topics: %w", err)
	}

	return nil
}

func (c *client) isReachable(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (c *client) NewProducer(topic string, opts ...ProducerOption) IProducer {
	return newProducer(c.brokers, topic, opts...)
}

func (c *client) NewConsumer(topic, group string, handler MessageHandler, opts ...ConsumerOption) IConsumer {
	return newConsumer(c.brokers, topic, group, handler, opts...)
}
