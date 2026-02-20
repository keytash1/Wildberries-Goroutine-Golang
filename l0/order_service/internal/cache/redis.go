package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"order-service/internal/config"
	"order-service/internal/domain"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(cfg config.RedisConfig) (OrderCache, error) {
	//redis options = pgx dsn
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis %w", err)
	}

	return &RedisCache{
		client: client,
		ttl:    time.Duration(cfg.TTL) * time.Second,
	}, nil
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

func (c *RedisCache) Set(id string, order *domain.Order) error {
	ctx := context.Background()

	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	return c.client.Set(ctx, "order:"+id, data, c.ttl).Err()
}

func (c *RedisCache) Get(id string) (*domain.Order, error) {
	ctx := context.Background()

	data, err := c.client.Get(ctx, "order:"+id).Bytes()
	if err == redis.Nil {
		//cache miss
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get from redis: %w", err)
	}

	var order domain.Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order: %w", err)
	}

	return &order, nil
}

func (c *RedisCache) Restore(orders []*domain.Order) error {
	ctx := context.Background()
	for _, order := range orders {
		data, err := json.Marshal(order)
		if err != nil {
			continue
		}
		c.client.Set(ctx, "order:"+order.OrderUID, data, c.ttl)
	}
	return nil
}
