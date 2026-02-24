package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"order-service/internal/domain"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Save(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockRepo) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockRepo) GetAll(ctx context.Context) ([]*domain.Order, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.Order), args.Error(1)
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Set(ctx context.Context, id string, order *domain.Order) error {
	args := m.Called(ctx, id, order)
	return args.Error(0)
}

func (m *MockCache) Get(ctx context.Context, id string) (*domain.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockCache) Restore(ctx context.Context, orders []*domain.Order) error {
	args := m.Called(ctx, orders)
	return args.Error(0)
}

func createValidOrder() *domain.Order {
	now := time.Now()
	return &domain.Order{
		OrderUID:    "test-123",
		TrackNumber: "TRACK123",
		CustomerID:  "cust-1",
		DateCreated: now,
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
}

func TestOrderService_GetOrder_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	service := NewOrderService(mockRepo, mockCache)

	expectedOrder := &domain.Order{
		OrderUID:   "test-123",
		CustomerID: "customer-1",
	}

	mockCache.On("Get", mock.Anything, "test-123").Return(expectedOrder, nil)

	result, err := service.GetOrder(ctx, "test-123")

	assert.NoError(t, err)
	assert.Equal(t, expectedOrder, result)
	mockCache.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything)
}

func TestOrderService_GetOrder_CacheMiss_DBHit(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	service := NewOrderService(mockRepo, mockCache)

	expectedOrder := &domain.Order{
		OrderUID:   "test-123",
		CustomerID: "customer-1",
	}

	mockCache.On("Get", mock.Anything, "test-123").Return(nil, nil)
	mockRepo.On("GetByID", mock.Anything, "test-123").Return(expectedOrder, nil)
	mockCache.On("Set", mock.Anything, "test-123", expectedOrder).Return(nil)

	result, err := service.GetOrder(ctx, "test-123")

	assert.NoError(t, err)
	assert.Equal(t, expectedOrder, result)
	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestOrderService_GetOrder_CacheError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	service := NewOrderService(mockRepo, mockCache)

	expectedOrder := &domain.Order{
		OrderUID:   "test-123",
		CustomerID: "customer-1",
	}

	mockCache.On("Get", mock.Anything, "test-123").Return(nil, errors.New("redis connection error"))
	mockRepo.On("GetByID", mock.Anything, "test-123").Return(expectedOrder, nil)
	mockCache.On("Set", mock.Anything, "test-123", expectedOrder).Return(nil)

	result, err := service.GetOrder(ctx, "test-123")

	assert.NoError(t, err)
	assert.Equal(t, expectedOrder, result)
	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestOrderService_GetOrder_NotFound(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	service := NewOrderService(mockRepo, mockCache)

	mockCache.On("Get", mock.Anything, "test-123").Return(nil, nil)
	mockRepo.On("GetByID", mock.Anything, "test-123").Return(nil, nil)

	result, err := service.GetOrder(ctx, "test-123")

	assert.ErrorIs(t, err, ErrOrderNotFound)
	assert.Nil(t, result)
	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestOrderService_GetOrder_DBError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	service := NewOrderService(mockRepo, mockCache)

	dbErr := errors.New("connection refused")

	mockCache.On("Get", mock.Anything, "test-123").Return(nil, nil)
	mockRepo.On("GetByID", mock.Anything, "test-123").Return(nil, dbErr)

	result, err := service.GetOrder(ctx, "test-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Nil(t, result)
	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestOrderService_GetOrder_CacheSetError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	service := NewOrderService(mockRepo, mockCache)

	expectedOrder := &domain.Order{
		OrderUID:   "test-123",
		CustomerID: "customer-1",
	}

	mockCache.On("Get", mock.Anything, "test-123").Return(nil, nil)
	mockRepo.On("GetByID", mock.Anything, "test-123").Return(expectedOrder, nil)
	mockCache.On("Set", mock.Anything, "test-123", expectedOrder).Return(errors.New("redis unavailable"))

	result, err := service.GetOrder(ctx, "test-123")

	assert.NoError(t, err)
	assert.Equal(t, expectedOrder, result)
	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestOrderService_SaveOrder_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	service := NewOrderService(mockRepo, mockCache)

	order := createValidOrder()

	mockRepo.On("Save", mock.Anything, order).Return(nil)
	mockCache.On("Set", mock.Anything, "test-123", order).Return(nil)

	err := service.SaveOrder(ctx, order)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestOrderService_SaveOrder_NilOrder(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	service := NewOrderService(mockRepo, mockCache)

	err := service.SaveOrder(ctx, nil)

	assert.ErrorIs(t, err, ErrInvalidOrder)
	mockRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	mockCache.AssertNotCalled(t, "Set", mock.Anything, mock.Anything, mock.Anything)
}

func TestOrderService_SaveOrder_EmptyUID(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	service := NewOrderService(mockRepo, mockCache)

	order := createValidOrder()
	order.OrderUID = ""

	err := service.SaveOrder(ctx, order)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing order_uid")
	mockRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	mockCache.AssertNotCalled(t, "Set", mock.Anything, mock.Anything, mock.Anything)
}

func TestOrderService_SaveOrder_DBError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	service := NewOrderService(mockRepo, mockCache)

	order := createValidOrder()
	dbErr := errors.New("duplicate key")

	mockRepo.On("Save", mock.Anything, order).Return(dbErr)

	err := service.SaveOrder(ctx, order)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save to db")
	mockRepo.AssertExpectations(t)
	mockCache.AssertNotCalled(t, "Set", mock.Anything, mock.Anything, mock.Anything)
}

func TestOrderService_SaveOrder_CacheError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	service := NewOrderService(mockRepo, mockCache)

	order := createValidOrder()

	mockRepo.On("Save", mock.Anything, order).Return(nil)
	mockCache.On("Set", mock.Anything, "test-123", order).Return(errors.New("redis unavailable"))

	err := service.SaveOrder(ctx, order)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestOrderService_RestoreCache_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	service := NewOrderService(mockRepo, mockCache)

	orders := []*domain.Order{
		{OrderUID: "order-1", CustomerID: "cust-1"},
		{OrderUID: "order-2", CustomerID: "cust-2"},
	}

	mockRepo.On("GetAll", mock.Anything).Return(orders, nil)
	mockCache.On("Restore", mock.Anything, orders).Return(nil)

	err := service.RestoreCache(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestOrderService_RestoreCache_RepoError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	service := NewOrderService(mockRepo, mockCache)

	repoErr := errors.New("database connection failed")

	mockRepo.On("GetAll", mock.Anything).Return([]*domain.Order(nil), repoErr)

	err := service.RestoreCache(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get all orders")
	mockRepo.AssertExpectations(t)
	mockCache.AssertNotCalled(t, "Restore", mock.Anything, mock.Anything)
}

func TestOrderService_RestoreCache_RestoreError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockRepo)
	mockCache := new(MockCache)
	service := NewOrderService(mockRepo, mockCache)

	orders := []*domain.Order{
		{OrderUID: "order-1", CustomerID: "cust-1"},
	}

	cacheErr := errors.New("redis restore failed")

	mockRepo.On("GetAll", mock.Anything).Return(orders, nil)
	mockCache.On("Restore", mock.Anything, orders).Return(cacheErr)

	err := service.RestoreCache(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to restore cache")
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestOrderService_SaveOrder_Validation(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*domain.Order)
		wantErr bool
	}{
		{
			name:    "valid order",
			modify:  func(o *domain.Order) {},
			wantErr: false,
		},
		{
			name:    "nil order",
			modify:  nil,
			wantErr: true,
		},
		{
			name:    "missing order_uid",
			modify:  func(o *domain.Order) { o.OrderUID = "" },
			wantErr: true,
		},
		{
			name:    "missing track_number",
			modify:  func(o *domain.Order) { o.TrackNumber = "" },
			wantErr: true,
		},
		{
			name:    "missing customer_id",
			modify:  func(o *domain.Order) { o.CustomerID = "" },
			wantErr: true,
		},
		{
			name:    "missing date_created",
			modify:  func(o *domain.Order) { o.DateCreated = time.Time{} },
			wantErr: true,
		},
		{
			name:    "missing delivery name",
			modify:  func(o *domain.Order) { o.Delivery.Name = "" },
			wantErr: true,
		},
		{
			name:    "missing delivery phone",
			modify:  func(o *domain.Order) { o.Delivery.Phone = "" },
			wantErr: true,
		},
		{
			name:    "missing delivery city",
			modify:  func(o *domain.Order) { o.Delivery.City = "" },
			wantErr: true,
		},
		{
			name:    "missing delivery address",
			modify:  func(o *domain.Order) { o.Delivery.Address = "" },
			wantErr: true,
		},
		{
			name:    "missing delivery email",
			modify:  func(o *domain.Order) { o.Delivery.Email = "" },
			wantErr: true,
		},
		{
			name:    "missing payment transaction",
			modify:  func(o *domain.Order) { o.Payment.Transaction = "" },
			wantErr: true,
		},
		{
			name:    "missing payment currency",
			modify:  func(o *domain.Order) { o.Payment.Currency = "" },
			wantErr: true,
		},
		{
			name:    "missing payment provider",
			modify:  func(o *domain.Order) { o.Payment.Provider = "" },
			wantErr: true,
		},
		{
			name:    "invalid payment amount",
			modify:  func(o *domain.Order) { o.Payment.Amount = 0 },
			wantErr: true,
		},
		{
			name:    "no items",
			modify:  func(o *domain.Order) { o.Items = []domain.Item{} },
			wantErr: true,
		},
		{
			name: "invalid item chrt_id",
			modify: func(o *domain.Order) {
				o.Items[0].ChrtID = 0
			},
			wantErr: true,
		},
		{
			name: "invalid item price",
			modify: func(o *domain.Order) {
				o.Items[0].Price = 0
			},
			wantErr: true,
		},
		{
			name: "missing item rid",
			modify: func(o *domain.Order) {
				o.Items[0].Rid = ""
			},
			wantErr: true,
		},
		{
			name: "missing item name",
			modify: func(o *domain.Order) {
				o.Items[0].Name = ""
			},
			wantErr: true,
		},
		{
			name: "invalid item total_price",
			modify: func(o *domain.Order) {
				o.Items[0].TotalPrice = 0
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockRepo := new(MockRepo)
			mockCache := new(MockCache)
			service := NewOrderService(mockRepo, mockCache)

			var order *domain.Order
			if tt.name == "nil order" {
				order = nil
			} else {
				baseOrder := createValidOrder()
				order = &domain.Order{
					OrderUID:    baseOrder.OrderUID,
					TrackNumber: baseOrder.TrackNumber,
					CustomerID:  baseOrder.CustomerID,
					DateCreated: baseOrder.DateCreated,
					Delivery: domain.Delivery{
						Name:    baseOrder.Delivery.Name,
						Phone:   baseOrder.Delivery.Phone,
						Zip:     baseOrder.Delivery.Zip,
						City:    baseOrder.Delivery.City,
						Address: baseOrder.Delivery.Address,
						Region:  baseOrder.Delivery.Region,
						Email:   baseOrder.Delivery.Email,
					},
					Payment: domain.Payment{
						Transaction:  baseOrder.Payment.Transaction,
						RequestID:    baseOrder.Payment.RequestID,
						Currency:     baseOrder.Payment.Currency,
						Provider:     baseOrder.Payment.Provider,
						Amount:       baseOrder.Payment.Amount,
						PaymentDt:    baseOrder.Payment.PaymentDt,
						Bank:         baseOrder.Payment.Bank,
						DeliveryCost: baseOrder.Payment.DeliveryCost,
						GoodsTotal:   baseOrder.Payment.GoodsTotal,
						CustomFee:    baseOrder.Payment.CustomFee,
					},
					Items: make([]domain.Item, len(baseOrder.Items)),
				}
				copy(order.Items, baseOrder.Items)

				if tt.modify != nil {
					tt.modify(order)
				}
			}

			if !tt.wantErr && order != nil {
				mockRepo.On("Save", mock.Anything, order).Return(nil).Once()
				mockCache.On("Set", mock.Anything, order.OrderUID, order).Return(nil).Once()
			}

			err := service.SaveOrder(ctx, order)

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidOrder)
				mockRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
				mockCache.AssertNotCalled(t, "Set", mock.Anything, mock.Anything, mock.Anything)
			} else {
				assert.NoError(t, err)
				mockRepo.AssertExpectations(t)
				mockCache.AssertExpectations(t)
			}
		})
	}
}
