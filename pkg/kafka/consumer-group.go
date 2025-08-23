package kafka

import (
	"context"
	"fmt"
	"sync"

	"github.com/IBM/sarama"
)

type BalanceStrategy string

const (
	StickyBalanceStrategy     BalanceStrategy = "sticky"
	RoundrobinBalanceStrategy BalanceStrategy = "roundrobin"
	RangeBalanceStrategy      BalanceStrategy = "range"
)

type consumerGroupRunner struct {
	client              sarama.ConsumerGroup
	consumer            *Consumer
	ctx                 context.Context
	cancel              context.CancelFunc
	errChan             chan error
	infoChan            chan string
	wg                  sync.WaitGroup
	topics              []string
	consumptionIsPaused bool
}

func WithBalancerConsumer(b BalanceStrategy) Option {
	return func(config *sarama.Config) {
		switch b {
		case StickyBalanceStrategy:
			config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategySticky()}
		case RoundrobinBalanceStrategy:
			config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
		case RangeBalanceStrategy:
			config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
		default:
			config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
		}
	}
}

func NewConsumerGroupRunner(brokers []string, groupID string, topics []string, bufferSize int, opts ...Option) (ConsumerGroupRunner, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	for _, opt := range opts {
		opt(config)
	}

	client, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	consumer := NewConsumer(bufferSize)

	return &consumerGroupRunner{
		client:              client,
		consumer:            consumer,
		ctx:                 ctx,
		cancel:              cancel,
		infoChan:            make(chan string, 1),
		errChan:             make(chan error, 1),
		topics:              topics,
		consumptionIsPaused: false,
	}, nil
}

func (r *consumerGroupRunner) Run() {
	r.wg.Add(1)

	go func() {
		defer r.wg.Done()

		for {
			if err := r.client.Consume(r.ctx, r.topics, r.consumer); err != nil {
				r.errChan <- fmt.Errorf("kafka consume error: %w", err)

				break
			}

			if r.ctx.Err() != nil {
				break
			}

			r.consumer.ready = make(chan bool)
		}
	}()
	// Wait until the user is configured.
	<-r.consumer.ready
	r.infoChan <- UpAndRunning
}

// Messages returns the channel where all received messages from kafka are recorded.
func (r *consumerGroupRunner) Messages() <-chan *MessageWithMarkFunc {
	return r.consumer.GetMessages()
}

// Shutdown Performs a graceful shutdown.
func (r *consumerGroupRunner) Shutdown() error {
	r.cancel()
	r.wg.Wait()

	err := r.client.Close()
	close(r.consumer.messages)
	close(r.errChan)
	close(r.infoChan)

	return err
}

func (r *consumerGroupRunner) Error() <-chan error {
	return r.errChan
}

func (r *consumerGroupRunner) Info() <-chan string {
	return r.infoChan
}
