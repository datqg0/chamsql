package kafka_config

import (
	"backend/pkgs/kafka"
)

func RegisterSystemTopics(registry *kafka.Registry) {
	registry.Register(kafka.TopicDefinition{
		Name:              TopicBookingEvents,
		NumPartitions:     12,
		ReplicationFactor: 1,
		ConsumerGroup:     GroupBookingWorkers,
	})

}
