package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"order-service/internal/config"
	"order-service/internal/domain"
	"order-service/internal/telemetry"
)

var cache_tracer = otel.Tracer("cache")

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(cfg config.RedisConfig) (*RedisCache, error) {
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

func (c *RedisCache) Set(ctx context.Context, id string, order *domain.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	if err := c.client.Set(ctx, "order:"+id, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set in redis: %w", err)
	}
	return nil
}

func (c *RedisCache) Get(ctx context.Context, id string) (*domain.Order, error) {
	ctx, span := cache_tracer.Start(ctx, "RedisCache.Get")
	defer span.End()

	span.SetAttributes(attribute.String("cache.key", id))

	data, err := c.client.Get(ctx, "order:"+id).Bytes()
	if errors.Is(err, redis.Nil) {
		span.SetAttributes(attribute.Bool("cache.hit", false))
		telemetry.RecordCacheMetrics(ctx, false)
		//cache miss
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get from redis: %w", err)
	}

	span.SetAttributes(attribute.Bool("cache.hit", true))
	telemetry.RecordCacheMetrics(ctx, true)

	var order domain.Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order: %w", err)
	}

	return &order, nil
}

func (c *RedisCache) Restore(ctx context.Context, orders []*domain.Order) error {
	for _, order := range orders {
		if order == nil {
			continue
		}
		data, err := json.Marshal(order)
		if err != nil {
			continue
		}
		c.client.Set(ctx, "order:"+order.OrderUID, data, c.ttl)
	}
	return nil
}
