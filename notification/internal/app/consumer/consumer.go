package inbox_consumer

import (
	"context"
	"time"

	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/handler"
	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/inbox_repo"
	kafkalib "github.com/Burmachine/MSA/lib/kafka"
	loggerlib "github.com/Burmachine/MSA/lib/logger"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type InboxConsumer struct {
	consumer     *kafkalib.Consumer
	repo         inbox_repo.InboxRepo
	batchSize    int
	batchTimeout time.Duration
	consumerName string
	logger       *loggerlib.Logger

	batch      []kafka.Message
	batchTimer *time.Timer
}

func NewInboxConsumer(
	brokers []string,
	groupID string,
	consumerName string,
	topic string,
	repo inbox_repo.InboxRepo,
	logger *loggerlib.Logger,
) (*InboxConsumer, error) {
	ic := &InboxConsumer{
		repo:         repo,
		batchSize:    128,
		batchTimeout: 300 * time.Millisecond,
		consumerName: consumerName,
		logger:       logger,
		batch:        make([]kafka.Message, 0, 128),
	}

	// Создаем consumer с MANUAL COMMIT
	consumer := kafkalib.NewConsumer(
		kafkalib.ConsumerConfig{
			Brokers:      brokers,
			GroupID:      groupID,
			Topic:        topic,
			ManualCommit: true, // Важно!
		},
		ic.messageHandler,
		logger,
	)

	ic.consumer = consumer
	return ic, nil
}

func (ic *InboxConsumer) messageHandler(ctx context.Context, msg kafka.Message) error {
	// Добавляем в batch
	ic.batch = append(ic.batch, msg)

	// Если batch заполнен - флашим
	if len(ic.batch) >= ic.batchSize {
		return ic.flushInbox(ctx)
	}

	return nil
}

func (ic *InboxConsumer) Run(ctx context.Context) error {
	ic.logger.Info(ctx, "starting inbox consumer",
		zap.String("consumer", ic.consumerName),
		zap.Int("batchSize", ic.batchSize),
		zap.Duration("batchTimeout", ic.batchTimeout),
	)

	// Таймер для периодического flush
	ic.batchTimer = time.NewTimer(ic.batchTimeout)
	defer ic.batchTimer.Stop()

	// Горутина для таймера
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ic.batchTimer.C:
				if len(ic.batch) > 0 {
					if err := ic.flushInbox(ctx); err != nil {
						ic.logger.Error(ctx, "error flushing on timer", zap.Error(err))
					}
				}
				ic.batchTimer.Reset(ic.batchTimeout)
			}
		}
	}()

	// Запускаем consumer
	return ic.consumer.Start(ctx)
}

func (ic *InboxConsumer) Close() error {
	ctx := context.Background()
	ic.logger.Info(ctx, "closing inbox consumer")

	// Флашим остатки
	if len(ic.batch) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := ic.flushInbox(ctx); err != nil {
			ic.logger.Error(ctx, "error flushing on close", zap.Error(err))
		}
	}

	// Закрываем consumer
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return ic.consumer.Stop(ctx)
}

func (ic *InboxConsumer) flushInbox(ctx context.Context) error {
	if len(ic.batch) == 0 {
		return nil
	}

	ic.logger.Debug(ctx, "flushing batch", zap.Int("size", len(ic.batch)))

	var inboxMsgs []inbox_repo.InboxMessage

	for _, msg := range ic.batch {
		id := handler.ExtractID(&msg)
		if id == "" {
			ic.logger.Warn(ctx, "skipping message without ID",
				zap.String("topic", msg.Topic),
				zap.Int("partition", msg.Partition),
				zap.Int64("offset", msg.Offset),
			)
			continue
		}

		inboxMsgs = append(inboxMsgs, inbox_repo.InboxMessage{
			ID:          id,
			Topic:       msg.Topic,
			Partition:   msg.Partition,
			KafkaOffset: msg.Offset,
			Payload:     msg.Value,
		})
	}

	// Batch insert в БД
	if len(inboxMsgs) > 0 {
		if err := ic.repo.BatchInsert(ctx, inboxMsgs); err != nil {
			ic.logger.Error(ctx, "batch insert failed", zap.Error(err))
			// НЕ очищаем batch и НЕ коммитим - retry от Kafka
			return err
		}

		ic.logger.Info(ctx, "batch inserted", zap.Int("count", len(inboxMsgs)))
	}

	// Коммитим ВСЕ сообщения из батча
	if err := ic.consumer.CommitMessages(ctx, ic.batch...); err != nil {
		ic.logger.Error(ctx, "commit failed", zap.Error(err))
		return err
	}

	// Очищаем batch
	ic.batch = ic.batch[:0]

	// Сбрасываем таймер
	if ic.batchTimer != nil {
		ic.batchTimer.Reset(ic.batchTimeout)
	}

	return nil
}
