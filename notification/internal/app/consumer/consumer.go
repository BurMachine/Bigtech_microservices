package inbox_consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/inbox_repo"
	kafkalib "github.com/Burmachine/MSA/lib/kafka"
	loggerlib "github.com/Burmachine/MSA/lib/logger"
	"github.com/Burmachine/MSA/lib/metrics"
	"github.com/segmentio/kafka-go"
)

type InboxConsumer struct {
	consumer *kafkalib.Consumer
	repo     inbox_repo.InboxRepo
	logger   *loggerlib.Logger
}

func NewInboxConsumer(
	brokers []string,
	groupID string,
	consumerName string,
	topic string,
	repo inbox_repo.InboxRepo,
	logger *loggerlib.Logger,
	m *metrics.Metrics,
) (*InboxConsumer, error) {

	ic := &InboxConsumer{
		repo:   repo,
		logger: logger,
	}

	// Создаем handler для обработки сообщений
	handler := func(ctx context.Context, msg kafka.Message) error {
		return ic.handleMessage(ctx, msg)
	}

	// Создаем consumer с метриками
	consumer := kafkalib.NewConsumer(
		kafkalib.ConsumerConfig{
			Brokers:      brokers,
			GroupID:      groupID,
			Topic:        topic,
			ManualCommit: false,
		},
		handler,
		logger,
		m,
	)

	ic.consumer = consumer

	return ic, nil
}

// handleMessage обрабатывает одно сообщение из Kafka
func (ic *InboxConsumer) handleMessage(ctx context.Context, msg kafka.Message) error {
	ic.logger.Debug(ctx, "received message",
		"topic", msg.Topic,
		"partition", msg.Partition,
		"offset", msg.Offset,
		"key", string(msg.Key),
	)

	// Парсим сообщение (если нужно для валидации)
	var event map[string]interface{}
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		ic.logger.Error(ctx, "failed to unmarshal message", "error", err)
		return fmt.Errorf("failed to unmarshal: %w", err)
	}

	// Создаем InboxMessage с правильными полями
	inboxMsg := inbox_repo.InboxMessage{
		ID:          string(msg.Key), // Или используйте UUID: uuid.New().String()
		Topic:       msg.Topic,
		Partition:   msg.Partition,
		KafkaOffset: msg.Offset,
		Payload:     msg.Value,
	}

	// Используем BatchInsert для одного сообщения
	if err := ic.repo.BatchInsert(ctx, []inbox_repo.InboxMessage{inboxMsg}); err != nil {
		ic.logger.Error(ctx, "failed to insert into inbox", "error", err, "key", string(msg.Key))
		return fmt.Errorf("failed to insert: %w", err)
	}

	ic.logger.Debug(ctx, "message saved to inbox", "key", string(msg.Key))
	return nil
}

// Run запускает consumer
func (ic *InboxConsumer) Run(ctx context.Context) error {
	return ic.consumer.Start(ctx)
}

// Close закрывает consumer
func (ic *InboxConsumer) Close() error {
	return ic.consumer.Stop(context.Background())
}
