package kafka

import (
	"encoding/json"
	"fmt"
	"time"
)

type Message struct {
	Key []byte
	Value []byte
	Headers map[string]string
	Topic string
	Timestamp time.Time
	Partition int
	Offset    int64
}

func NewJSONMessage(key string, value interface{}, headers map[string]string) (Message, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return Message{}, fmt.Errorf("kafka: marshal message value: %w", err)
	}

	msg := Message{
		Value:   data,
		Headers: headers,
	}
	if key != "" {
		msg.Key = []byte(key)
	}
	return msg, nil
}

func NewRawMessage(key string, value []byte, headers map[string]string) Message {
	msg := Message{
		Value:   value,
		Headers: headers,
	}
	if key != "" {
		msg.Key = []byte(key)
	}
	return msg
}
