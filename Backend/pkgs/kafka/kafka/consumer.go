package kafka

import (
	"context"
	"math/rand"
	"time"

	"backend/pkgs/logger"

	kg "github.com/segmentio/kafka-go"
)

type IConsumer interface {
	Start(ctx context.Context) error
	Close() error
}

type MessageHandler func(ctx context.Context, msg Message) error

type ConsumerOption func(*consumerConfig)

type consumerConfig struct {
	startOffset    int64
	maxBytes       int
	minBytes       int
	commitInterval time.Duration
	maxWait        time.Duration
	retryBackoff   RetryBackoff
}

type RetryBackoff struct {
	Min    time.Duration
	Max    time.Duration
	Jitter float64 // fraction of backoff to randomize (0.0 – 1.0)
}

func defaultConsumerConfig() consumerConfig {
	return consumerConfig{
		startOffset:    kg.LastOffset,
		maxBytes:       10 * 1024 * 1024, // 10MB
		minBytes:       1,
		commitInterval: time.Second,
		maxWait:        3 * time.Second,
		retryBackoff: RetryBackoff{
			Min:    1 * time.Second,
			Max:    30 * time.Second,
			Jitter: 0.2,
		},
	}
}

func WithStartOffset(offset int64) ConsumerOption {
	return func(c *consumerConfig) { c.startOffset = offset }
}

func WithMaxBytes(n int) ConsumerOption {
	return func(c *consumerConfig) {
		if n > 0 {
			c.maxBytes = n
		}
	}
}

func WithCommitInterval(d time.Duration) ConsumerOption {
	return func(c *consumerConfig) {
		if d > 0 {
			c.commitInterval = d
		}
	}
}

func WithMaxWait(d time.Duration) ConsumerOption {
	return func(c *consumerConfig) {
		if d > 0 {
			c.maxWait = d
		}
	}
}

func WithRetryBackoff(min, max time.Duration, jitter float64) ConsumerOption {
	return func(c *consumerConfig) {
		if min > 0 {
			c.retryBackoff.Min = min
		}
		if max > 0 {
			c.retryBackoff.Max = max
		}
		if jitter >= 0 && jitter <= 1 {
			c.retryBackoff.Jitter = jitter
		}
	}
}

type consumer struct {
	reader  *kg.Reader
	handler MessageHandler
	topic   string
	group   string
	cfg     consumerConfig
	brokers []string
}

func newConsumer(brokers []string, topic, group string, handler MessageHandler, opts ...ConsumerOption) IConsumer {
	cfg := defaultConsumerConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	reader := kg.NewReader(kg.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        group,
		StartOffset:    cfg.startOffset,
		MinBytes:       cfg.minBytes,
		MaxBytes:       cfg.maxBytes,
		MaxWait:        cfg.maxWait,
		CommitInterval: cfg.commitInterval,
	})

	logger.Info("Kafka consumer created: topic=%s, group=%s", topic, group)
	return &consumer{
		reader:  reader,
		handler: handler,
		topic:   topic,
		group:   group,
		cfg:     cfg,
		brokers: brokers,
	}
}

func (c *consumer) Start(ctx context.Context) error {
	logger.Info("Kafka consumer started: topic=%s, group=%s", c.topic, c.group)
	defer logger.Info("Kafka consumer stopped: topic=%s, group=%s", c.topic, c.group)

	attempt := 0
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			attempt++
			backoff := calcBackoff(attempt, c.cfg.retryBackoff)
			logger.Error("Kafka consumer fetch error (topic=%s, group=%s, attempt=%d): %v", c.topic, c.group, attempt, err)
			logger.Warn("Kafka consumer retrying in %s", backoff)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
			continue
		}

		attempt = 0

		consumed := fromKafkaGoMessage(msg)
		if err := c.handler(ctx, consumed); err != nil {
			logger.Error("Kafka consumer handler error (topic=%s, partition=%d, offset=%d): %v",
				c.topic, msg.Partition, msg.Offset, err)
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			logger.Error("Kafka consumer commit error (topic=%s, partition=%d, offset=%d): %v",
				c.topic, msg.Partition, msg.Offset, err)
		}
	}
}

func (c *consumer) Close() error {
	if err := c.reader.Close(); err != nil {
		logger.Error("Kafka consumer close error (topic=%s, group=%s): %v", c.topic, c.group, err)
		return err
	}
	logger.Info("Kafka consumer closed: topic=%s, group=%s", c.topic, c.group)
	return nil
}

func fromKafkaGoMessage(msg kg.Message) Message {
	headers := make(map[string]string, len(msg.Headers))
	for _, h := range msg.Headers {
		headers[h.Key] = string(h.Value)
	}

	return Message{
		Key:       msg.Key,
		Value:     msg.Value,
		Headers:   headers,
		Topic:     msg.Topic,
		Timestamp: msg.Time,
		Partition: msg.Partition,
		Offset:    msg.Offset,
	}
}

func calcBackoff(attempt int, rb RetryBackoff) time.Duration {
	if attempt < 1 {
		return rb.Min
	}

	backoff := rb.Min
	for i := 1; i < attempt; i++ {
		backoff *= 2
		if backoff >= rb.Max {
			backoff = rb.Max
			break
		}
	}

	if rb.Jitter > 0 {
		jitterMax := int64(float64(backoff) * rb.Jitter)
		if jitterMax > 0 {
			backoff += time.Duration(rand.Int63n(jitterMax))
		}
	}

	if backoff > rb.Max {
		return rb.Max
	}
	return backoff
}
