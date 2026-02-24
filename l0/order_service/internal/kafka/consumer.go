package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"

	"order-service/internal/config"
	"order-service/internal/domain"
	"order-service/internal/service"
)

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

func (c *Consumer) Start() {
	log.Printf("Kafka consumer started [pid=%d]", os.Getpid())
	for {
		select {
		case <-c.stopChan:
			log.Println("Kafka consumer stopped")
			return
		default:
			msg, err := c.reader.FetchMessage(context.Background())
			if err != nil {
				//after maxwait
				log.Printf("Kafka fetch error: %v", err)
				continue
			}
			if err := c.processMessage(msg); err != nil {
				log.Printf("Failed to process message: %v", err)
				continue
			}
			if err := c.reader.CommitMessages(context.Background(), msg); err != nil {
				log.Printf("Failed to commit message: %v", err)
			}
			log.Printf("Message recieved")
		}
	}
}

// обработка полученного сообщения
func (c *Consumer) processMessage(msg kafka.Message) error {
	var order domain.Order
	//из []bytes json в struct
	if err := json.Unmarshal(msg.Value, &order); err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}

	log.Printf("Received order: %s", order.OrderUID)

	//сохраняем в бд и кэш
	log.Printf("Try to recieve message")
	if err := c.orderService.SaveOrder(&order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	log.Printf("Order saved: %s", order.OrderUID)
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
