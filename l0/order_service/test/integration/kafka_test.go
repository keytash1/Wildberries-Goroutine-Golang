package integration

import (
	"context"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"order-service/internal/cache"
	"order-service/internal/config"
	"order-service/internal/domain"
	myKafka "order-service/internal/kafka"
	"order-service/internal/repository"
	"order-service/internal/service"
)

func TestKafkaConsumer_RealKafka_ProcessMessage(t *testing.T) {
	cleanupDB()
	cleanupRedis()
	cleanupKafkaTopic()

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

	producer := myKafka.NewProducer("localhost:9093", "test-orders")
	defer producer.Close()

	consumerCtx, consumerCancel := context.WithCancel(testCtx)
	defer consumerCancel()

	consumer := myKafka.NewConsumer(config.KafkaConfig{
		Brokers:       "localhost:9093",
		Topic:         "test-orders",
		ConsumerGroup: "test-group",
	}, orderService)

	go consumer.Start(consumerCtx)
	time.Sleep(10 * time.Second)

	order := &domain.Order{
		OrderUID:    "test-kafka-1",
		TrackNumber: "TRACK123",
		CustomerID:  "cust-1",
		DateCreated: time.Now(),
		Delivery: domain.Delivery{
			Name:    "Test User",
			Phone:   "+123456789",
			City:    "Moscow",
			Address: "Street 1",
			Email:   "test@test.com",
		},
		Payment: domain.Payment{
			Transaction: "trans-123",
			Currency:    "USD",
			Provider:    "test",
			Amount:      1000,
		},
		Items: []domain.Item{
			{
				ChrtID:     1,
				Price:      500,
				Rid:        "rid-1",
				Name:       "Item 1",
				TotalPrice: 500,
			},
		},
	}

	err = producer.SendOrder(order)
	require.NoError(t, err)

	time.Sleep(10 * time.Second)

	dbOrder, err := repo.GetByID(testCtx, "test-kafka-1")
	require.NoError(t, err)
	require.NotNil(t, dbOrder)
	assert.Equal(t, order.CustomerID, dbOrder.CustomerID)

	cached, err := redisCache.Get(testCtx, "test-kafka-1")
	require.NoError(t, err)
	assert.NotNil(t, cached)
}

func TestKafkaConsumer_RealKafka_InvalidMessage(t *testing.T) {
	cleanupDB()
	cleanupRedis()
	cleanupKafkaTopic()

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

	consumerCtx, consumerCancel := context.WithCancel(testCtx)
	defer consumerCancel()

	consumer := myKafka.NewConsumer(config.KafkaConfig{
		Brokers:       "localhost:9093",
		Topic:         "test-orders",
		ConsumerGroup: "test-group",
	}, orderService)

	go consumer.Start(consumerCtx)
	time.Sleep(10 * time.Second)

	writer := &kafka.Writer{
		Addr:  kafka.TCP("localhost:9093"),
		Topic: "test-orders",
	}
	defer writer.Close()

	err = writer.WriteMessages(testCtx, kafka.Message{
		Value: []byte(`{"invalid": json`),
	})
	require.NoError(t, err)

	time.Sleep(10 * time.Second)

	orders, err := repo.GetAll(testCtx)
	require.NoError(t, err)
	assert.Empty(t, orders)
}
