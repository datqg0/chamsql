package kafka_config

import (
	"backend/pkgs/kafka"
)

func RegisterSystemTopics(registry *kafka.Registry) {
	// Exam domain events
	registry.Register(kafka.TopicDefinition{
		Name:              TopicExamEvents,
		NumPartitions:     3,
		ReplicationFactor: 1,
		ConsumerGroup:     GroupExamWorkers,
	})

	// Submission domain events
	registry.Register(kafka.TopicDefinition{
		Name:              TopicSubmissionEvents,
		NumPartitions:     6,
		ReplicationFactor: 1,
		ConsumerGroup:     GroupSubmissionWorkers,
	})
}
