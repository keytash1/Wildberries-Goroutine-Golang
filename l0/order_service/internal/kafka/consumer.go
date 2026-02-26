package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"order-service/internal/config"
	"order-service/internal/domain"
	"order-service/internal/service"
	"order-service/internal/telemetry"
)

var consumer_tracer = otel.Tracer("kafka-consumer")

// cache like repo, kafkaCons like handler
type Consumer struct {
	reader       KafkaReader //router
	orderService service.OrderServiceInterface
	stopChan     chan struct{}
}

func NewConsumer(cfg config.KafkaConfig, orderService *service.OrderService) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{cfg.Brokers},
		Topic:       cfg.Topic,
		GroupID:     cfg.ConsumerGroup,
		StartOffset: kafka.FirstOffset,
		MinBytes:    10e3,
		MaxBytes:    10e6,
		MaxWait:     1 * time.Second,
	})

	return &Consumer{
		reader:       reader,
		orderService: orderService,
		stopChan:     make(chan struct{}),
	}
}

func (c *Consumer) Start(ctx context.Context) {
	log.Printf("Kafka consumer started [pid=%d]", os.Getpid())
	for {
		select {
		case <-c.stopChan:
			log.Println("Kafka consumer stopped")
			return
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				//after maxwait
				log.Printf("Kafka fetch error: %v", err)
				continue
			}
			if err := c.processMessage(ctx, msg); err != nil {
				log.Printf("Failed to process message: %v", err)
				continue
			}
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("Failed to commit message: %v", err)
			}
			log.Printf("Message recieved")
		}
	}
}

// обработка полученного сообщения
func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) error {
	ctx, span := consumer_tracer.Start(ctx, "Consumer.processMessage")
	defer span.End()

	var order domain.Order
	//из []bytes json в struct
	if err := json.Unmarshal(msg.Value, &order); err != nil {
		span.RecordError(err)
		telemetry.RecordKafkaMetrics(ctx, false)
		return fmt.Errorf("failed to unmarshal: %w", err)
	}

	span.SetAttributes(attribute.String("order.id", order.OrderUID))
	log.Printf("Received order: %s", order.OrderUID)

	//сохраняем в бд и кэш
	log.Printf("Try to recieve message")
	if err := c.orderService.SaveOrder(ctx, &order); err != nil {
		span.RecordError(err)
		telemetry.RecordKafkaMetrics(ctx, false)
		return fmt.Errorf("failed to save order: %w", err)
	}

	span.SetStatus(codes.Ok, "message processed successfully")
	log.Printf("Order saved: %s", order.OrderUID)

	telemetry.RecordKafkaMetrics(ctx, true)
	return nil
}

func (c *Consumer) Stop() error {
	//выходим из цикла в start
	close(c.stopChan)
	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("failed to close reader: %w", err)
	}
	return nil
}
