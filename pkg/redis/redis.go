package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

type Redis interface {
	RDB() *goredis.Client
	Close() error
}

type Config struct {
	Host     string
	Port     uint16
	Password string
	DB       int
}

type redis struct {
	rdb *goredis.Client
}

func New(cfg *Config) (Redis, error) {
	rdb := goredis.NewClient(&goredis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if resp := rdb.Ping(context.Background()); resp.Err() != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", resp.Err())
	}

	return &redis{rdb: rdb}, nil
}

func (r *redis) RDB() *goredis.Client {
	return r.rdb
}

func (r *redis) Close() error {
	return r.rdb.Close()
}
