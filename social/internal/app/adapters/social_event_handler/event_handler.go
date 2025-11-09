package social_event_handler

import (
	"context"
	"fmt"

	"github.com/BurMachine/Bigtech_microservices/social/internal/app/models"
	kafkalib "github.com/Burmachine/MSA/lib/kafka"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type KafkaEventsHandler struct {
	producer *kafkalib.Producer
}

func NewKafkaEventsHandler(producer *kafkalib.Producer) *KafkaEventsHandler {
	return &KafkaEventsHandler{
		producer: producer,
	}
}

func (h *KafkaEventsHandler) HandleBatch(ctx context.Context, events []*models.Event) (succeeded []uuid.UUID, failed []uuid.UUID, err error) {
	succeeded = make([]uuid.UUID, 0, len(events))
	failed = make([]uuid.UUID, 0, len(events))

	if len(events) == 0 {
		return succeeded, failed, nil
	}

	// Формируем kafka.Message для батча
	messages := make([]kafka.Message, 0, len(events))
	for _, ev := range events {
		messages = append(messages, kafka.Message{
			Topic: ev.Topic,
			Key:   []byte(ev.ID.String()),
			Value: ev.Payload,
			Time:  ev.CreatedAt,
		})
	}

	// Отправляем батчем через платформенный продюсер
	writeErr := h.producer.PublishBatch(ctx, messages)
	if writeErr != nil {
		// При ошибке все сообщения считаем неудачными
		for _, ev := range events {
			failed = append(failed, ev.ID)
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
	// Продюсер закрывается через cleanup функции платформы
	// Здесь ничего не делаем
	return nil
}
