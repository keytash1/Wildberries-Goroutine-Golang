package config

import (
	"os"
	"strconv"
)

type Config struct {
	DB    DBConfig
	Redis RedisConfig
	Kafka KafkaConfig
	HTTP  HTTPConfig
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	TTL      int
}

type KafkaConfig struct {
	Brokers       string
	Topic         string
	GroupID       string
	ConsumerGroup string
}

type HTTPConfig struct {
	Port string
}

func Load() (*Config, error) {
	return &Config{
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "orders"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			TTL:      getEnvAsInt("REDIS_TTL", 86400),
		},
		Kafka: KafkaConfig{
			Brokers:       getEnv("KAFKA_BROKERS", "kafka:9092"),
			Topic:         getEnv("KAFKA_TOPIC", "orders"),
			GroupID:       getEnv("KAFKA_GROUP_ID", "order-service"),
			ConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "order-group"),
		},
		HTTP: HTTPConfig{
			Port: getEnv("HTTP_PORT", "8081"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
