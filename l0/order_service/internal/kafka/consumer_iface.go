package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type KafkaReader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}
