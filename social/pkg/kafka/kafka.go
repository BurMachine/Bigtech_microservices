package kafka

import (
	"context"
	"errors"
	"fmt"

	"github.com/segmentio/kafka-go"
)

// WriterConfig определяет настройки для продюсера kafka-go.
type WriterConfig struct {
	Brokers              []string
	Topic                string
	RequiredAcks         kafka.RequiredAcks
	MaxAttempts          int
	BatchSize            int
	BatchBytes           int64
	Compression          kafka.Compression
	AllowAutoTopicCreate bool
}

// DefaultWriterConfig возвращает конфигурацию по умолчанию.
func DefaultWriterConfig(brokers []string) *WriterConfig {
	return &WriterConfig{
		Brokers:              brokers,
		Topic:                "test",           // Топик задаётся при отправке или в обработчике
		RequiredAcks:         kafka.RequireAll, // Аналог sarama.WaitForAll
		MaxAttempts:          30,
		BatchSize:            100,
		BatchBytes:           1 << 20, // 1MB
		Compression:          kafka.Snappy,
		AllowAutoTopicCreate: true, // DEV=true, PROD=false
	}
}

// NewSyncProducer создаёт синхронный продюсер с использованием kafka-go.
func NewSyncProducer(ctx context.Context, brokers []string) (*kafka.Writer, error) {
	cfg := DefaultWriterConfig(brokers)

	if len(cfg.Brokers) == 0 {
		return nil, errors.New("kafka: no brokers provided")
	}

	// Создаём Writer
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Topic:                  cfg.Topic,     // Может быть пустым, если топик задаётся в сообщениях
		Balancer:               &kafka.Hash{}, // Аналог sarama.NewHashPartitioner
		MaxAttempts:            cfg.MaxAttempts,
		BatchSize:              cfg.BatchSize,
		BatchBytes:             cfg.BatchBytes,
		RequiredAcks:           cfg.RequiredAcks,
		Compression:            cfg.Compression,
		AllowAutoTopicCreation: cfg.AllowAutoTopicCreate,
	}

	// Проверяем подключение к брокерам
	conn, err := kafka.DialContext(ctx, "tcp", cfg.Brokers[0])
	if err != nil {
		return nil, fmt.Errorf("kafka: failed to connect to broker %s: %w", cfg.Brokers[0], err)
	}
	defer conn.Close()

	return writer, nil
}
