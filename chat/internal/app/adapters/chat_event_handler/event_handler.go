package chat_event_handler

import (
	"context"
	"fmt"
	"time"

	"github.com/BurMachine/Bigtech_microservices/chat/internal/app/models"
	"github.com/segmentio/kafka-go"
)

type KafkaEventsHandler struct {
	writer *kafka.Writer
}

func NewKafkaEventsHandler(brokers []string, topic string) *KafkaEventsHandler {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...), // ← БЕЗ Topic
		Topic:        topic,
		Balancer:     &kafka.Hash{}, // Партиционирование по ключу
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		MaxAttempts:  3,
		RequiredAcks: kafka.RequireAll,
		Compression:  kafka.Snappy,
	}

	return &KafkaEventsHandler{writer: writer}
}

func (h *KafkaEventsHandler) HandleEvent(ctx context.Context, event *models.Event) (err error) {
	message := kafka.Message{
		Key:   []byte(event.ID.String()),
		Value: event.Payload,
		Time:  event.CreatedAt,
	}

	err = h.writer.WriteMessages(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to write single message: %w", err)
	}

	return nil
}

func (h *KafkaEventsHandler) Close() error {
	return h.writer.Close()
}
