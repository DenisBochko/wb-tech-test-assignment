// Package kafka provides Kafka-based high-level implementations for the ebus package interfaces.
// It includes configurable producers and consumers built on the https://github.com/IBM/sarama.
package kafka

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
)

const (
	UpAndRunning = "kafka consumer up and running"
)

var ErrSendMessageTimeout = fmt.Errorf("kafka send message timeout")

// Option defines a configuration function for Kafka producers.
type Option func(*sarama.Config)

// Producer defines the behavior for publishing messages to an event bus.
// Implementations should ensure messages are delivered according to configured options.
type Producer interface {
	PushMessage(ctx context.Context, key, value []byte) (partition int32, offset int64, err error)
	Close() error
}

// ConsumerGroupRunner defines the behavior for consuming messages from an event bus.
type ConsumerGroupRunner interface {
	Run()
	Messages() <-chan *MessageWithMarkFunc
	Shutdown() error
	Error() <-chan error
	Info() <-chan string
}
