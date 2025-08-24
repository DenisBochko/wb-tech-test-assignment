package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"wb-tech-test-assignment/internal/config"
	"wb-tech-test-assignment/internal/model"
	"wb-tech-test-assignment/internal/repository"
	"wb-tech-test-assignment/pkg/postgres"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"wb-tech-test-assignment/pkg/kafka"
)

type OrderRepository interface {
	InsertOrder(ctx context.Context, ext repository.RepoExtension, order model.Order) error
	InsertDelivery(ctx context.Context, ext repository.RepoExtension, orderUID string, delivery model.Delivery) error
	InsertPayment(ctx context.Context, ext repository.RepoExtension, orderUID string, payment model.Payment) error
	InsertItems(ctx context.Context, ext repository.RepoExtension, orderUID string, items []model.Item) error
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

	tx, err := s.db.Pool().Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("error starting transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	err = s.orderRepo.InsertOrder(ctx, tx, order)
	if err != nil {
		return "", fmt.Errorf("failed to insert order: %w", err)
	}

	err = s.orderRepo.InsertDelivery(ctx, tx, order.OrderUID, order.Delivery)
	if err != nil {
		return "", fmt.Errorf("failed to insert delivery: %w", err)
	}

	err = s.orderRepo.InsertPayment(ctx, tx, order.OrderUID, order.Payment)
	if err != nil {
		return "", fmt.Errorf("failed to insert payment: %w", err)
	}

	err = s.orderRepo.InsertItems(ctx, tx, order.OrderUID, order.Items)
	if err != nil {
		return "", fmt.Errorf("failed to insert items: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("error committing transaction: %w", err)
	}

	return order.OrderUID, nil
}
