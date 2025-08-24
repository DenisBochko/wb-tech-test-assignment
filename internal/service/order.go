package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"wb-tech-test-assignment/internal/config"
	"wb-tech-test-assignment/internal/model"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"wb-tech-test-assignment/pkg/kafka"
)

type OrderService struct {
	log      *zap.Logger
	cfg      *config.Subscriber
	Consumer kafka.ConsumerGroupRunner
}

func NewOrderService(log *zap.Logger, cfg *config.Subscriber, Consumer kafka.ConsumerGroupRunner) *OrderService {
	return &OrderService{
		log:      log,
		cfg:      cfg,
		Consumer: Consumer,
	}
}

func (s *OrderService) Run(ctx context.Context) error {
	s.Consumer.Run()

	wg := &sync.WaitGroup{}

	messages := make(chan *kafka.MessageWithMarkFunc, s.cfg.OrdersSubscriber.BufferSize)
	defer close(messages)

	wg.Add(s.cfg.WorkerCount)
	for i := 0; i < s.cfg.WorkerCount; i++ {
		go func() {
			defer wg.Done()
			s.worker(i, messages)

			s.log.Info("worker finished", zap.Int("worker_id", i))
		}()
	}

	for msg := range s.Consumer.Messages() {
		select {
		case messages <- msg:
		case <-ctx.Done():
			return ctx.Err()
		case err := <-s.Consumer.Error():
			return err
		}
	}

	return nil
}

func (s *OrderService) Shutdown() error {
	err := s.Consumer.Shutdown()
	if err != nil {
		return fmt.Errorf("failed to shutdown consumer: %w", err)
	}

	return nil
}

func (s *OrderService) worker(id int, message <-chan *kafka.MessageWithMarkFunc) {
	s.log.Info("worker start", zap.Int("worker_id", id))

	for msg := range message {
		var order model.Order

		err := json.Unmarshal(msg.Message.Value, &order)
		if err != nil {
			s.log.Error("Error unmarshalling message", zap.Error(err), zap.Int("worker_id", id))
		}

		validate := validator.New()
		if err := validate.Struct(order); err != nil {
			s.log.Error("Validation error", zap.Error(err), zap.Int("worker_id", id))
		} else {
			s.log.Info("Order is valid ✅", zap.Int("worker_id", id))
		}
	}
}
