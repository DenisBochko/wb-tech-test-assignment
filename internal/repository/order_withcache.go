package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"wb-tech-test-assignment/internal/model"

	"github.com/redis/go-redis/v9"
)

const defaultTTL = 10 * time.Second

type DefaultOrderRepository interface {
	PutOrder(ctx context.Context, order model.Order) error
	GetOrder(ctx context.Context, orderUID string) (model.Order, error)
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
			fmt.Println("ЗАКАЗА НЕТ В КЕШЕ")

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
