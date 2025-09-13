package kafka

import (
	"context"
	"time"

	"github.com/IBM/sarama"
)

type Balancer int

const (
	// RoundRobin selects partitions in round-robin order.
	RoundRobin Balancer = iota

	// Hash uses the record key to hash and assign partitions.
	Hash

	// Random distributes messages randomly across partitions.
	Random

	// ReferenceHash mirrors the Java client's DefaultPartitioner behavior.
	ReferenceHash

	// ConsistentHash uses a consistent hashing algorithm.
	ConsistentHash
)

// RequiredAcks defines acknowledgement levels for message writes.
type RequiredAcks int16

const (
	// RequireAll waits for acknowledgement from all in-sync replicas.
	RequireAll RequiredAcks = -1

	// RequireNone does not wait for any acknowledgement.
	RequireNone RequiredAcks = 0

	// RequireOne waits for acknowledgement from the leader only.
	RequireOne RequiredAcks = 1
)

// WithBalancer configures the producer to use the specified Balancer strategy.
func WithBalancer(b Balancer) Option {
	return func(cfg *sarama.Config) {
		switch b {
		case Hash:
			cfg.Producer.Partitioner = sarama.NewHashPartitioner
		case Random:
			cfg.Producer.Partitioner = sarama.NewRandomPartitioner
		case ReferenceHash:
			cfg.Producer.Partitioner = sarama.NewReferenceHashPartitioner
		case ConsistentHash:
			cfg.Producer.Partitioner = sarama.NewConsistentCRCHashPartitioner
		default:
			cfg.Producer.Partitioner = sarama.NewRoundRobinPartitioner
		}
	}
}

// WithRequiredAcks configures the producer to use the specified RequiredAcks level.
func WithRequiredAcks(a RequiredAcks) Option {
	return func(cfg *sarama.Config) {
		cfg.Producer.RequiredAcks = sarama.RequiredAcks(a)
	}
}

type producer struct {
	syncProducer sarama.SyncProducer
	topic        string
}

func NewProducer(brokers []string, topic string, opts ...Option) (Producer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true

	for _, opt := range opts {
		opt(cfg)
	}

	syncProducer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return nil, err
	}

	return &producer{
		syncProducer: syncProducer,
		topic:        topic,
	}, nil
}

func (p *producer) PushMessage(ctx context.Context, key, value []byte) (partition int32, offset int64, err error) {
	msg := &sarama.ProducerMessage{
		Topic:     p.topic,
		Key:       sarama.ByteEncoder(key),
		Value:     sarama.ByteEncoder(value),
		Timestamp: time.Now(),
	}

	result := make(chan struct {
		partition int32
		offset    int64
		err       error
	}, 1)

	go func() {
		partition, offset, err := p.syncProducer.SendMessage(msg)
		result <- struct {
			partition int32
			offset    int64
			err       error
		}{partition, offset, err}
	}()

	select {
	case <-ctx.Done():
		return 0, 0, ctx.Err()
	case res := <-result:
		return res.partition, res.offset, res.err
	}
}

func (p *producer) Close() error {
	return p.syncProducer.Close()
}
