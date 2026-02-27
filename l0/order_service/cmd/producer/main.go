package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"order-service/internal/config"
	"order-service/internal/domain"
	"order-service/internal/kafka"
)

func main() {
	cfg := config.Load()

	data, err := os.ReadFile("model.json")
	if err != nil {
		log.Fatal("Failed to read model.json:", err)
	}

	var order domain.Order
	if err := json.Unmarshal(data, &order); err != nil {
		log.Fatal("Failed to parse JSON:", err)
	}

	producer := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	defer producer.Close()
	log.Println(cfg.Kafka.Brokers, cfg.Kafka.Topic)

	// valid
	log.Printf("Sending order: %s", order.OrderUID)
	if err := producer.SendOrder(&order); err != nil {
		log.Fatal("Failed to send:", err)
	}
	log.Println("Valid order sent")

	// invalid empty uid
	invalidOrder := order
	invalidOrder.OrderUID = ""
	log.Printf("Sending invalid order (empty UID)")
	if err := producer.SendOrder(&invalidOrder); err != nil {
		log.Printf("Expected error for invalid order: %v", err)
	} else {
		log.Printf("Unexpected success for invalid order")
	}

	//more valid tests
	testOrders := []string{"test-order-1", "test-order-2", "test-order-3"}

	for _, id := range testOrders {
		testOrder := order
		testOrder.OrderUID = fmt.Sprintf("%s-%d", id, time.Now().UnixNano())
		testOrder.CustomerID = "test-customer-" + id
		testOrder.DateCreated = time.Now()
		log.Printf("Sending test order: %s", id)
		if err := producer.SendOrder(&testOrder); err != nil {
			log.Printf("Failed to send %s: %v", id, err)
		} else {
			log.Printf("Test order %s sent", id)
		}
		time.Sleep(500 * time.Millisecond)
	}

	log.Println("Done")
}
