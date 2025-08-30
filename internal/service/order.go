package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"wb-tech-test-assignment/internal/config"
	"wb-tech-test-assignment/internal/model"
	"wb-tech-test-assignment/pkg/kafka"
	"wb-tech-test-assignment/pkg/postgres"
)

type OrderRepository interface {
	PutOrder(ctx context.Context, order model.Order) error
	GetOrder(ctx context.Context, orderUID string) (model.Order, error)
}

type OrderService struct {
	log       *zap.Logger
	cfg       *config.Subscriber
	consumer  kafka.ConsumerGroupRunner
	db        postgres.Postgres
	orderRepo OrderRepository
}

func NewOrderService(log *zap.Logger, cfg *config.Subscriber, consumer kafka.ConsumerGroupRunner, db postgres.Postgres, orderRepo OrderRepository) *OrderService {
	return &OrderService{
		log:       log,
		cfg:       cfg,
		consumer:  consumer,
		db:        db,
		orderRepo: orderRepo,
	}
}

func (s *OrderService) Run(ctx context.Context) error {
	s.consumer.Run()

	wg := &sync.WaitGroup{}

	messages := make(chan *kafka.MessageWithMarkFunc, s.cfg.OrdersSubscriber.BufferSize)
	defer close(messages)

	wg.Add(s.cfg.WorkerCount)
	for i := 0; i < s.cfg.WorkerCount; i++ {
		go func() {
			defer wg.Done()
			s.worker(ctx, i, messages)

			s.log.Info("worker finished", zap.Int("worker_id", i))
		}()
	}

	for msg := range s.consumer.Messages() {
		select {
		case messages <- msg:
		case <-ctx.Done():
			return ctx.Err()
		case err := <-s.consumer.Error():
			return err
		}
	}

	return nil
}

func (s *OrderService) Shutdown() error {
	err := s.consumer.Shutdown()
	if err != nil {
		return fmt.Errorf("failed to shutdown consumer: %w", err)
	}

	return nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderUID string) (model.Order, error) {
	order, err := s.orderRepo.GetOrder(ctx, orderUID)
	if err != nil {
		return model.Order{}, fmt.Errorf("failed to get order: %w", err)
	}

	return order, nil
}

func (s *OrderService) worker(ctx context.Context, id int, message <-chan *kafka.MessageWithMarkFunc) {
	s.log.Info("worker start", zap.Int("worker_id", id))

	for msg := range message {
		orderUID, err := s.processOrder(ctx, msg.Message.Value)
		if err != nil {
			s.log.Error("Failed to process order", zap.Error(err), zap.Int("worker_id", id), zap.String("order_uid", orderUID))
		}

		s.log.Info("Order processed", zap.Int("worker_id", id), zap.String("order_uid", orderUID))

		msg.Mark()
	}
}

func (s *OrderService) processOrder(ctx context.Context, message []byte) (string, error) {
	var order model.Order
	if err := json.Unmarshal(message, &order); err != nil {
		return "", fmt.Errorf("failed to unmarshal order: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(order); err != nil {
		return "", fmt.Errorf("failed to validate order: %w", err)
	}

	if err := s.orderRepo.PutOrder(ctx, order); err != nil {
		return "", fmt.Errorf("failed to put order: %w", err)
	}

	return order.OrderUID, nil
}
