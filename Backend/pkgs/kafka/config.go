package kafka

import "strings"

type Config struct {
	Enabled  bool
	Brokers  []string
	ClientID string
}

type TopicDefinition struct {
	Name              string
	NumPartitions     int
	ReplicationFactor int
	ConsumerGroup     string // default consumer group for consumers of this topic
}

func (td TopicDefinition) Validate() bool {
	return strings.TrimSpace(td.Name) != ""
}

func ParseBrokers(raw string) []string {
	parts := strings.Split(raw, ",")
	brokers := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			brokers = append(brokers, trimmed)
		}
	}
	return brokers
}
