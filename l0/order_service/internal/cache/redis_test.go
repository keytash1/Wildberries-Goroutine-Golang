package cache

import (
	"strconv"
	"testing"
	"time"

	"order-service/internal/config"
	"order-service/internal/domain"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisCache_NewRedisCache_Success(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	port, err := strconv.Atoi(mr.Port())
	require.NoError(t, err)

	cfg := config.RedisConfig{
		Host: "127.0.0.1",
		Port: port,
		DB:   0,
		TTL:  3600,
	}

	cache, err := NewRedisCache(cfg)
	require.NoError(t, err)
	assert.NotNil(t, cache)
}

func TestRedisCache_NewRedisCache_InvalidPort(t *testing.T) {
	cfg := config.RedisConfig{
		Host: "127.0.0.1",
		Port: 9999,
		DB:   0,
		TTL:  3600,
	}

	cache, err := NewRedisCache(cfg)
	assert.Error(t, err)
	assert.Nil(t, cache)
}

func TestRedisCache_SetAndGet_Success(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	port, err := strconv.Atoi(mr.Port())
	require.NoError(t, err)

	cfg := config.RedisConfig{
		Host: "127.0.0.1",
		Port: port,
		DB:   0,
		TTL:  3600,
	}

	cache, err := NewRedisCache(cfg)
	require.NoError(t, err)

	now := time.Now().UTC().Truncate(time.Second)
	order := &domain.Order{
		OrderUID:    "test-123",
		TrackNumber: "TRACK123",
		Entry:       "WBIL",
		CustomerID:  "cust-1",
		DateCreated: now,
		Delivery: domain.Delivery{
			Name:  "Test User",
			Phone: "+123456789",
			City:  "Moscow",
		},
		Payment: domain.Payment{
			Transaction: "trans-123",
			Amount:      1000,
			Currency:    "RUB",
		},
		Items: []domain.Item{
			{
				ChrtID: 1,
				Name:   "Item 1",
				Price:  500,
			},
		},
	}

	err = cache.Set(order.OrderUID, order)
	assert.NoError(t, err)

	result, err := cache.Get(order.OrderUID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, order.OrderUID, result.OrderUID)
	assert.Equal(t, order.TrackNumber, result.TrackNumber)
	assert.Equal(t, order.CustomerID, result.CustomerID)
	assert.Equal(t, order.DateCreated.Unix(), result.DateCreated.Unix())
	assert.Equal(t, order.Delivery.Name, result.Delivery.Name)
	assert.Equal(t, order.Payment.Amount, result.Payment.Amount)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, order.Items[0].Name, result.Items[0].Name)
}

func TestRedisCache_Get_NotFound(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	port, err := strconv.Atoi(mr.Port())
	require.NoError(t, err)

	cfg := config.RedisConfig{
		Host: "127.0.0.1",
		Port: port,
		DB:   0,
		TTL:  3600,
	}

	cache, err := NewRedisCache(cfg)
	require.NoError(t, err)

	result, err := cache.Get("non-existent")
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestRedisCache_Set_TTL(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	port, err := strconv.Atoi(mr.Port())
	require.NoError(t, err)

	cfg := config.RedisConfig{
		Host: "127.0.0.1",
		Port: port,
		DB:   0,
		TTL:  1,
	}

	cache, err := NewRedisCache(cfg)
	require.NoError(t, err)

	order := &domain.Order{
		OrderUID:   "test-123",
		CustomerID: "cust-1",
	}

	err = cache.Set(order.OrderUID, order)
	assert.NoError(t, err)

	mr.FastForward(2 * time.Second)

	result, err := cache.Get(order.OrderUID)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestRedisCache_Restore_Success(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	port, err := strconv.Atoi(mr.Port())
	require.NoError(t, err)

	cfg := config.RedisConfig{
		Host: "127.0.0.1",
		Port: port,
		DB:   0,
		TTL:  3600,
	}

	cache, err := NewRedisCache(cfg)
	require.NoError(t, err)

	now := time.Now().UTC().Truncate(time.Second)
	orders := []*domain.Order{
		{
			OrderUID:    "order-1",
			TrackNumber: "TRACK1",
			CustomerID:  "cust-1",
			DateCreated: now,
		},
		{
			OrderUID:    "order-2",
			TrackNumber: "TRACK2",
			CustomerID:  "cust-2",
			DateCreated: now,
		},
	}

	err = cache.Restore(orders)
	assert.NoError(t, err)

	for _, expected := range orders {
		result, err := cache.Get(expected.OrderUID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expected.OrderUID, result.OrderUID)
		assert.Equal(t, expected.TrackNumber, result.TrackNumber)
		assert.Equal(t, expected.CustomerID, result.CustomerID)
	}
}

func TestRedisCache_Restore_WithInvalidOrder(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	port, err := strconv.Atoi(mr.Port())
	require.NoError(t, err)

	cfg := config.RedisConfig{
		Host: "127.0.0.1",
		Port: port,
		DB:   0,
		TTL:  3600,
	}

	cache, err := NewRedisCache(cfg)
	require.NoError(t, err)

	validOrder := &domain.Order{
		OrderUID:   "order-1",
		CustomerID: "cust-1",
	}

	orders := []*domain.Order{
		validOrder,
		nil,
	}

	err = cache.Restore(orders)
	assert.NoError(t, err)

	result, err := cache.Get("order-1")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "order-1", result.OrderUID)
}

func TestRedisCache_Close(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	port, err := strconv.Atoi(mr.Port())
	require.NoError(t, err)

	cfg := config.RedisConfig{
		Host: "127.0.0.1",
		Port: port,
		DB:   0,
		TTL:  3600,
	}

	cache, err := NewRedisCache(cfg)
	require.NoError(t, err)

	err = cache.Close()
	assert.NoError(t, err)
}
