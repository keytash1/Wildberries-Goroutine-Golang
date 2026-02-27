package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"order-service/internal/cache"
	"order-service/internal/config"
	"order-service/internal/domain"
	"order-service/internal/handler"
	"order-service/internal/repository"
	"order-service/internal/service"
)

func TestAPI_RealDB_GetOrder(t *testing.T) {
	cleanupDB()

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
	orderHandler := handler.NewOrderHandler(orderService)

	now := time.Now().UTC().Truncate(time.Second)
	order := &domain.Order{
		OrderUID:    "test-api-1",
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

	err = repo.Save(testCtx, order)
	require.NoError(t, err)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/order/:id", orderHandler.GetOrder)

	req, _ := http.NewRequestWithContext(testCtx, "GET", "/order/test-api-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.Order
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, order.OrderUID, response.OrderUID)
	assert.Equal(t, order.Delivery.Name, response.Delivery.Name)
	assert.Len(t, response.Items, 1)
}

func TestAPI_RealDB_OrderNotFound(t *testing.T) {
	cleanupDB()

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
	orderHandler := handler.NewOrderHandler(orderService)

	r := gin.Default()
	r.GET("/order/:id", orderHandler.GetOrder)

	req, _ := http.NewRequestWithContext(testCtx, "GET", "/order/non-existent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "order not found", response["error"])
}
