package kafka

import (
	"github.com/IBM/sarama"
)

type Consumer struct {
	ready    chan bool
	messages chan *MessageWithMarkFunc
}

type MessageWithMarkFunc struct {
	Message *sarama.ConsumerMessage
	Mark    func()
}

func NewConsumer(bufferSize int) *Consumer {
	return &Consumer{
		ready:    make(chan bool),
		messages: make(chan *MessageWithMarkFunc, bufferSize),
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(c.ready)

	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited.
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing.
// loop and exit.
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/IBM/sarama/blob/main/consumer_group.go#L27-L29
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			// If the channel is full, we wait until the space is free.
			msg := message // copy value

			select {
			case c.messages <- &MessageWithMarkFunc{
				Message: msg,
				Mark: func() {
					session.MarkMessage(msg, "")
				},
			}:
			case <-session.Context().Done():
				return nil
			}
		case <-session.Context().Done():
			return nil
		}
	}
}

// GetMessages returns the channel where all received messages from kafka are recorded.
func (c *Consumer) GetMessages() <-chan *MessageWithMarkFunc {
	return c.messages
}
