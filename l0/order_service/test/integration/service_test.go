package integration

import (
	"order-service/internal/cache"
	"order-service/internal/config"
	"order-service/internal/domain"
	"order-service/internal/repository"
	"order-service/internal/service"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_RealDBAndRedis_SaveAndGet(t *testing.T) {
	cleanupDB()
	cleanupRedis()

	repo, err := repository.NewPostgresOrderRepo(testCtx, config.DBConfig{
		Host:     "localhost",
		Port:     5433,
		User:     "postgres",
		Password: "postgres",
		Name:     "orders_test",
	})
	require.NoError(t, err)
	defer repo.Close()

	redisCache, err := cache.NewRedisCache(config.RedisConfig{
		Host: "localhost",
		Port: 6380,
		DB:   0,
		TTL:  3600,
	})
	require.NoError(t, err)
	defer redisCache.Close()

	orderService := service.NewOrderService(repo, redisCache)

	now := time.Now().UTC().Truncate(time.Second)
	order := &domain.Order{
		OrderUID:    "test-service-1",
		TrackNumber: "TRACK123",
		CustomerID:  "cust-1",
		DateCreated: now,
		Delivery: domain.Delivery{
			Name:    "Test User",
			Phone:   "+123456789",
			Zip:     "123456",
			City:    "Moscow",
			Address: "Street 1",
			Region:  "MO",
			Email:   "test@test.com",
		},
		Payment: domain.Payment{
			Transaction:  "trans-123",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "test",
			Amount:       1000,
			PaymentDt:    123456789,
			Bank:         "test-bank",
			DeliveryCost: 100,
			GoodsTotal:   900,
			CustomFee:    0,
		},
		Items: []domain.Item{
			{
				ChrtID:      1,
				TrackNumber: "TRACK123",
				Price:       500,
				Rid:         "rid-1",
				Name:        "Item 1",
				Sale:        0,
				Size:        "M",
				TotalPrice:  500,
				NmID:        100,
				Brand:       "Test Brand",
				Status:      1,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		DeliveryService:   "test-delivery",
		Shardkey:          "0",
		SmID:              1,
		OofShard:          "0",
	}

	err = orderService.SaveOrder(testCtx, order)
	require.NoError(t, err)

	retrieved, err := orderService.GetOrder(testCtx, "test-service-1")
	require.NoError(t, err)
	assert.Equal(t, order.OrderUID, retrieved.OrderUID)

	cached, err := redisCache.Get(testCtx, "test-service-1")
	require.NoError(t, err)
	assert.NotNil(t, cached)
}
