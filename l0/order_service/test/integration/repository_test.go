package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"order-service/internal/config"
	"order-service/internal/domain"
	"order-service/internal/repository"
)

func TestPostgresRepo_ReadlDB_SaveAndGet(t *testing.T) {
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

	now := time.Now().Truncate(time.Second)
	order := &domain.Order{
		OrderUID:    "test-repo-1",
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

	saved, err := repo.GetByID(testCtx, "test-repo-1")
	require.NoError(t, err)
	require.NotNil(t, saved)

	assert.Equal(t, order.OrderUID, saved.OrderUID)
	assert.Equal(t, order.Delivery.Name, saved.Delivery.Name)
	assert.Len(t, saved.Items, 1)
}

func TestPostgresRepo_RealDB_NotFound(t *testing.T) {
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

	order, err := repo.GetByID(testCtx, "non-existent")
	require.NoError(t, err)
	assert.Nil(t, order)
}
