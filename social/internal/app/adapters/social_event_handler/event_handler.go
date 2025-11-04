package social_event_handler

import (
	"context"
	"fmt"
	"time"

	"github.com/BurMachine/Bigtech_microservices/social/internal/app/models"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type KafkaEventsHandler struct {
	writer *kafka.Writer
}

func NewKafkaEventsHandler(brokers []string) *KafkaEventsHandler {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...), // ← БЕЗ Topic
		Balancer:     &kafka.Hash{},         // Партиционирование по ключу
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		MaxAttempts:  3,
		RequiredAcks: kafka.RequireAll,
		Compression:  kafka.Snappy,
	}

	return &KafkaEventsHandler{writer: writer}
}

func (h *KafkaEventsHandler) HandleBatch(ctx context.Context, events []*models.Event) (succeeded []uuid.UUID, failed []uuid.UUID, err error) {
	succeeded = make([]uuid.UUID, 0, len(events))
	failed = make([]uuid.UUID, 0, len(events))

	messages := make([]kafka.Message, 0, len(events))
	for _, ev := range events {
		messages = append(messages, kafka.Message{
			Topic: ev.Topic, // ← Теперь работает
			Key:   []byte(ev.ID.String()),
			Value: ev.Payload,
			Time:  ev.CreatedAt,
		})
	}

	// Отправляем батчем (эффективнее)
	writeErr := h.writer.WriteMessages(ctx, messages...)
	if writeErr != nil {
		// Определяем какие сообщения провалились
		for i, ev := range events {
			if i < len(messages) {
				failed = append(failed, ev.ID)
			}
		}
		return succeeded, failed, fmt.Errorf("failed to write batch: %w", writeErr)
	}

	// Все успешны
	for _, ev := range events {
		succeeded = append(succeeded, ev.ID)
	}

	return succeeded, failed, nil
}

func (h *KafkaEventsHandler) Close() error {
	return h.writer.Close()
}
