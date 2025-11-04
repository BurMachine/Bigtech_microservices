package consumer

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/handler"
	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/inbox_repo"
	"github.com/segmentio/kafka-go"
)

type InboxConsumer struct {
	reader       *kafka.Reader
	repo         inbox_repo.InboxRepo
	handler      handler.Handler
	batchSize    int
	batchTimeout time.Duration
	consumerName string
}

func NewInboxConsumer(brokers []string, groupID string, consumerName string, topic string, repo inbox_repo.InboxRepo, h handler.Handler) (*InboxConsumer, error) {
	// Создаем Reader с указанием топика при инициализации
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		Topic:       topic, // Топик устанавливается здесь
		MinBytes:    10e3,  // 10KB
		MaxBytes:    10e6,  // 10MB
		StartOffset: kafka.FirstOffset,
	})

	return &InboxConsumer{
		reader:       reader,
		repo:         repo,
		handler:      h,
		batchSize:    128,
		batchTimeout: 300 * time.Millisecond,
		consumerName: consumerName,
	}, nil
}

func (c *InboxConsumer) Close() error {
	return c.reader.Close()
}

func (c *InboxConsumer) Run(ctx context.Context) error {
	batch := make([]kafka.Message, 0, c.batchSize)
	timer := time.NewTimer(c.batchTimeout)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			c.flushInbox(ctx, batch) // Flush перед выходом
			return ctx.Err()
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return err
				}
				log.Printf("[kafka-reader] error: %v", err)
				continue
			}

			batch = append(batch, msg)
			if len(batch) >= c.batchSize {
				c.flushInbox(ctx, batch)
				batch = batch[:0]
				timer.Reset(c.batchTimeout)
			}
		case <-timer.C:
			if len(batch) > 0 {
				c.flushInbox(ctx, batch)
				batch = batch[:0]
			}
			timer.Reset(c.batchTimeout)
		}
	}
}

// flushInbox: Batch insert в БД, затем commit successful
func (c *InboxConsumer) flushInbox(ctx context.Context, batch []kafka.Message) {
	if len(batch) == 0 {
		return
	}

	var inboxMsgs []inbox_repo.InboxMessage
	var toCommit []kafka.Message

	for _, msg := range batch {
		id := handler.ExtractID(&msg)
		if id == "" {
			// Без ID: скип + commit (или DLQ)
			log.Printf("skip message without ID: topic=%s p=%d off=%d", msg.Topic, msg.Partition, msg.Offset)
			toCommit = append(toCommit, msg)
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

	// Batch insert в БД (ON CONFLICT DO NOTHING)
	if len(inboxMsgs) > 0 {
		if err := c.repo.BatchInsert(ctx, inboxMsgs); err != nil {
			log.Printf("batch insert failed: %v — will retry", err)
			return // НЕ commit — retry от Kafka
		}
	}

	// Успех: commit все сообщения из батча
	if err := c.reader.CommitMessages(ctx, batch...); err != nil {
		log.Printf("commit failed: %v", err)
	}
}
