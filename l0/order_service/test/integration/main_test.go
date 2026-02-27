package integration

import (
	"context"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

var testCtx = context.Background()

func TestMain(m *testing.M) {
	log.Println("starting test containers...")
	cmd := exec.Command("docker-compose", "-f", "docker-compose.test.yml", "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal("failed to start test containers:", err)
	}

	waitForServices()
	runMigrations()

	code := m.Run()

	log.Println("stopping test containers...")
	cmd = exec.Command("docker-compose", "-f", "docker-compose.test.yml", "down", "-v")
	cmd.Run()

	os.Exit(code)
}

func waitForServices() {
	log.Println("waiting for services to be ready...")

	dsn := "postgres://postgres:postgres@localhost:5433/orders_test?sslmode=disable"
	for i := 0; i < 30; i++ {
		conn, err := pgxpool.New(testCtx, dsn)
		if err == nil {
			conn.Close()
			log.Println("postgresql is ready")
			break
		}
		if i == 29 {
			log.Fatal("postgresql not ready after 30 seconds")
		}
		time.Sleep(1 * time.Second)
	}

	client := redis.NewClient(&redis.Options{Addr: "localhost:6380"})
	defer client.Close()
	for i := 0; i < 30; i++ {
		if _, err := client.Ping(testCtx).Result(); err == nil {
			log.Println("redis is ready")
			break
		}
		if i == 29 {
			log.Fatal("redis not ready after 30 seconds")
		}
		time.Sleep(1 * time.Second)
	}

	for i := 0; i < 30; i++ {
		conn, err := kafka.Dial("tcp", "localhost:9093")
		if err == nil {
			conn.Close()
			log.Println("kafka is ready")
			break
		}
		if i == 29 {
			log.Fatal("kafka not ready after 30 seconds")
		}
		time.Sleep(1 * time.Second)
	}
}

func runMigrations() {
	dsn := "postgres://postgres:postgres@localhost:5433/orders_test?sslmode=disable"

	pool, err := pgxpool.New(testCtx, dsn)
	if err != nil {
		log.Fatal("failed to connect to postgres for migrations:", err)
	}
	defer pool.Close()

	migrationSQL, err := os.ReadFile("../../migrations/001_init_schema.up.sql")
	if err != nil {
		log.Fatal("failed to read migration:", err)
	}

	for i := 0; i < 10; i++ {
		_, err = pool.Exec(testCtx, string(migrationSQL))
		if err == nil {
			break
		}
		log.Printf("failed to execute migration (attempt %d/10): %v", i+1, err)
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		log.Fatal("failed to execute migration after 10 attempts:", err)
	}

	log.Println("migrations applied")
}

func cleanupDB() {
	dsn := "postgres://postgres:postgres@localhost:5433/orders_test?sslmode=disable"

	pool, err := pgxpool.New(testCtx, dsn)
	if err != nil {
		log.Printf("failed to connect to postgres for cleanup: %v", err)
		return
	}
	defer pool.Close()

	_, err = pool.Exec(testCtx, "TRUNCATE orders, delivery, payment, items CASCADE")
	if err != nil {
		log.Printf("cleanup warning: %v", err)
	}
}

func cleanupRedis() {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6380"})
	defer client.Close()
	client.FlushAll(testCtx)
}

func cleanupKafkaTopic() {
	//delete and recreate topic
	conn, err := kafka.Dial("tcp", "localhost:9093")
	if err != nil {
		log.Printf("Failed to connect to Kafka: %v", err)
		return
	}

	err = conn.DeleteTopics("test-orders")
	if err != nil {
		log.Printf("Delete topic error: %v", err)
	}
	conn.Close()

	time.Sleep(2 * time.Second)

	conn2, err := kafka.Dial("tcp", "localhost:9093")
	if err != nil {
		log.Printf("Failed to reconnect to Kafka: %v", err)
		return
	}
	defer conn2.Close()

	err = conn2.CreateTopics(kafka.TopicConfig{
		Topic:             "test-orders",
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
	if err != nil {
		log.Printf("Failed to create topic: %v", err)
	}

	time.Sleep(2 * time.Second)
}
