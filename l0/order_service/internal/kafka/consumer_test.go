package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"order-service/internal/domain"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func (m *MockOrderService) SaveOrder(order *domain.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func TestConsumer_Start_Stop(t *testing.T) {
	mockReader := new(MockReader)
	mockService := new(MockOrderService)

	mockReader.On("Close").Return(nil)

	consumer := &Consumer{
		reader:       mockReader,
		orderService: mockService,
		stopChan:     make(chan struct{}),
	}

	go consumer.Start()
	time.Sleep(10 * time.Millisecond)

	err := consumer.Stop()
	assert.NoError(t, err)

	mockReader.AssertExpectations(t)
}

func TestConsumer_processMessage_Success(t *testing.T) {
	mockService := new(MockOrderService)

	consumer := &Consumer{
		orderService: mockService,
	}

	order := &domain.Order{OrderUID: "test-123"}
	data, _ := json.Marshal(order)
	msg := kafka.Message{Value: data}

	mockService.On("SaveOrder", order).Return(nil)

	err := consumer.processMessage(msg)

	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

func TestConsumer_processMessage_UnmarshalError(t *testing.T) {
	mockService := new(MockOrderService)

	consumer := &Consumer{
		orderService: mockService,
	}

	msg := kafka.Message{Value: []byte(`{"invalid": json`)}
	err := consumer.processMessage(msg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal")
	mockService.AssertNotCalled(t, "SaveOrder")
}

func TestConsumer_processMessage_SaveError(t *testing.T) {
	mockService := new(MockOrderService)

	consumer := &Consumer{
		orderService: mockService,
	}

	order := &domain.Order{OrderUID: "test-123"}
	data, _ := json.Marshal(order)
	msg := kafka.Message{Value: data}

	saveErr := errors.New("database error")
	mockService.On("SaveOrder", order).Return(saveErr)

	err := consumer.processMessage(msg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save order")
	mockService.AssertExpectations(t)
}
