package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"wb-tech-test-assignment/internal/model"
)

const (
	defaultTTL       = 24 * time.Hour
	defaultBatchSize = 100
)

type DefaultOrderRepository interface {
	PutOrder(ctx context.Context, order model.Order) error
	GetOrder(ctx context.Context, orderUID string) (model.Order, error)
	GetOrdersBatch(ctx context.Context, limit, offset int) ([]model.Order, error)
}

type OrderWithCacheRepository struct {
	rdb  *redis.Client
	repo DefaultOrderRepository
}

func NewOrderWithCacheRepository(rdb *redis.Client, defaultRepo DefaultOrderRepository) *OrderWithCacheRepository {
	return &OrderWithCacheRepository{
		rdb:  rdb,
		repo: defaultRepo,
	}
}

func (o *OrderWithCacheRepository) PutOrder(ctx context.Context, order model.Order) error {
	if err := o.repo.PutOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to put order in DB: %w", err)
	}

	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	if err := o.rdb.Set(ctx, order.OrderUID, data, defaultTTL).Err(); err != nil {
		return fmt.Errorf("failed to set order in redis: %w", err)
	}

	return nil
}

func (o *OrderWithCacheRepository) GetOrder(ctx context.Context, orderUID string) (model.Order, error) {
	var order model.Order

	val, err := o.rdb.Get(ctx, orderUID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			order, err = o.repo.GetOrder(ctx, orderUID)
			if err != nil {
				return model.Order{}, fmt.Errorf("failed to get order from DB: %w", err)
			}

			data, err := json.Marshal(order)
			if err != nil {
				return model.Order{}, fmt.Errorf("failed to marshal order: %w", err)
			}

			if err := o.rdb.Set(ctx, order.OrderUID, data, defaultTTL).Err(); err != nil {
				return model.Order{}, fmt.Errorf("failed to set order in redis: %w", err)
			}

			return order, nil
		}
	}

	if err := json.Unmarshal([]byte(val), &order); err != nil {
		return model.Order{}, fmt.Errorf("failed to unmarshal order: %w", err)
	}

	return order, nil
}

func (o *OrderWithCacheRepository) WarmupCache(ctx context.Context) error {
	offset := 0

	for {
		orders, err := o.repo.GetOrdersBatch(ctx, defaultBatchSize, offset)
		if err != nil {
			return fmt.Errorf("failed to get orders batch: %w", err)
		}

		if len(orders) == 0 {
			break
		}

		for _, order := range orders {
			data, err := json.Marshal(order)
			if err != nil {
				return fmt.Errorf("failed to marshal order: %w", err)
			}

			if err := o.rdb.Set(ctx, order.OrderUID, data, defaultTTL).Err(); err != nil {
				return fmt.Errorf("failed to set order in redis: %w", err)
			}
		}

		offset += defaultBatchSize
	}

	return nil
}
