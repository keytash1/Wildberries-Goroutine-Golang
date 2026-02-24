package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"

	"order-service/internal/domain"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			Async:        false,
			BatchTimeout: 10 * time.Millisecond,
		},
	}
}

func (p *Producer) SendOrder(order *domain.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	msg := kafka.Message{
		Value: data,
		Headers: []kafka.Header{
			{Key: "source", Value: []byte("producer")},
			{Key: "timestamp", Value: []byte(time.Now().String())},
		},
	}

	if err := p.writer.WriteMessages(context.Background(), msg); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
