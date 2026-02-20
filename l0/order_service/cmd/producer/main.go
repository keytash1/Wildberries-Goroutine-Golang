package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"order-service/internal/config"
	"order-service/internal/domain"
	"order-service/internal/kafka"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	data, err := ioutil.ReadFile("model.json")
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

	//сначала анмаршал потому опять маршал?
	log.Printf("Sending order: %s", order.OrderUID)
	if err := producer.SendOrder(&order); err != nil {
		log.Fatal("Failed to send:", err)
	}
	log.Println("Order sent")

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
