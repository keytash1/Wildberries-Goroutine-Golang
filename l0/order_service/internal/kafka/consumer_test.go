package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"order-service/internal/domain"
)

type MockReader struct {
	mock.Mock
}

func (m *MockReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	args := m.Called(ctx)
	return args.Get(0).(kafka.Message), args.Error(1)
}

func (m *MockReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func (m *MockReader) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) SaveOrder(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderService) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderService) RestoreCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestConsumer_Start_Stop(t *testing.T) {
	ctx := context.Background()
	mockReader := new(MockReader)
	mockService := new(MockOrderService)

	mockReader.On("Close").Return(nil)

	consumer := &Consumer{
		reader:       mockReader,
		orderService: mockService,
		stopChan:     make(chan struct{}),
	}

	go consumer.Start(ctx)
	time.Sleep(10 * time.Millisecond)

	err := consumer.Stop()
	assert.NoError(t, err)

	mockReader.AssertExpectations(t)
}

func TestConsumer_processMessage_Success(t *testing.T) {
	ctx := context.Background()
	mockService := new(MockOrderService)

	consumer := &Consumer{
		orderService: mockService,
	}

	order := &domain.Order{OrderUID: "test-123"}
	data, _ := json.Marshal(order)
	msg := kafka.Message{Value: data}

	mockService.On("SaveOrder", mock.Anything, order).Return(nil)

	err := consumer.processMessage(ctx, msg)

	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

func TestConsumer_processMessage_UnmarshalError(t *testing.T) {
	ctx := context.Background()
	mockService := new(MockOrderService)

	consumer := &Consumer{
		orderService: mockService,
	}

	msg := kafka.Message{Value: []byte(`{"invalid": json`)}
	err := consumer.processMessage(ctx, msg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal")
	mockService.AssertNotCalled(t, "SaveOrder", mock.Anything, mock.Anything)
}

func TestConsumer_processMessage_SaveError(t *testing.T) {
	ctx := context.Background()
	mockService := new(MockOrderService)

	consumer := &Consumer{
		orderService: mockService,
	}

	order := &domain.Order{OrderUID: "test-123"}
	data, _ := json.Marshal(order)
	msg := kafka.Message{Value: data}

	saveErr := errors.New("database error")
	mockService.On("SaveOrder", mock.Anything, order).Return(saveErr)

	err := consumer.processMessage(ctx, msg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save order")
	mockService.AssertExpectations(t)
}
