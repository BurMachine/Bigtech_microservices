package kafkalib

import (
	"context"
	"fmt"
	"time"

	"github.com/Burmachine/MSA/lib/logger"
	"github.com/segmentio/kafka-go"
)

type ConsumerConfig struct {
	Brokers        []string
	GroupID        string
	Topic          string
	CommitInterval time.Duration // Интервал автокоммита (0 = manual commit)
	ManualCommit   bool          // Если true - не коммитить автоматически
	StartOffset    int64         // kafka.FirstOffset, kafka.LastOffset
}

type MessageHandler func(ctx context.Context, msg kafka.Message) error

type Consumer struct {
	reader  *kafka.Reader
	logger  *loggerlib.Logger
	handler MessageHandler
	stopCh  chan struct{}
	config  ConsumerConfig
}

func NewConsumer(cfg ConsumerConfig, handler MessageHandler, log *loggerlib.Logger) *Consumer {
	// Дефолты
	if cfg.CommitInterval == 0 && !cfg.ManualCommit {
		cfg.CommitInterval = time.Second
	}
	if cfg.StartOffset == 0 {
		cfg.StartOffset = kafka.FirstOffset
	}

	readerCfg := kafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		GroupID:     cfg.GroupID,
		Topic:       cfg.Topic,
		MinBytes:    1,
		MaxBytes:    10e6, // 10MB
		StartOffset: cfg.StartOffset,
	}

	// Если manual commit - не устанавливаем CommitInterval
	if !cfg.ManualCommit {
		readerCfg.CommitInterval = cfg.CommitInterval
	}

	reader := kafka.NewReader(readerCfg)

	return &Consumer{
		reader:  reader,
		logger:  log,
		handler: handler,
		stopCh:  make(chan struct{}),
		config:  cfg,
	}
}

// Start запускает consumer (вызывать в горутине)
func (c *Consumer) Start(ctx context.Context) error {
	c.logger.Info(ctx, "starting kafka consumer",
		"topic", c.reader.Config().Topic,
		"group_id", c.reader.Config().GroupID,
		"manual_commit", c.config.ManualCommit,
	)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info(ctx, "consumer context cancelled")
			return ctx.Err()
		case <-c.stopCh:
			c.logger.Info(ctx, "consumer stopped")
			return nil
		default:
			// FetchMessage блокируется до получения сообщения
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					c.logger.Info(ctx, "context cancelled or deadline exceeded")
					return err
				}
				c.logger.Error(ctx, "error fetching message", "error", err)
				time.Sleep(time.Second) // backoff при ошибке
				continue
			}

			c.logger.Debug(ctx, "received message",
				"topic", msg.Topic,
				"partition", msg.Partition,
				"offset", msg.Offset,
				"key", string(msg.Key),
			)

			// Обрабатываем сообщение
			if err := c.handler(ctx, msg); err != nil {
				c.logger.Error(ctx, "error handling message",
					"error", err,
					"key", string(msg.Key),
					"offset", msg.Offset,
					"partition", msg.Partition,
				)
				// При ошибке НЕ коммитим - Kafka retry'нет
				continue
			}

			// Коммитим только если не manual commit
			if !c.config.ManualCommit {
				if err := c.reader.CommitMessages(ctx, msg); err != nil {
					c.logger.Error(ctx, "error committing message",
						"error", err,
						"offset", msg.Offset,
					)
				} else {
					c.logger.Debug(ctx, "message committed",
						"offset", msg.Offset,
						"partition", msg.Partition,
					)
				}
			}
		}
	}
}

// CommitMessages явный коммит (для manual commit mode)
func (c *Consumer) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	if len(msgs) == 0 {
		return nil
	}

	c.logger.Debug(ctx, "committing messages", "count", len(msgs))

	if err := c.reader.CommitMessages(ctx, msgs...); err != nil {
		c.logger.Error(ctx, "error committing messages",
			"error", err,
			"count", len(msgs),
		)
		return err
	}

	c.logger.Debug(ctx, "messages committed successfully", "count", len(msgs))
	return nil
}

// Commit коммитит текущие оффсеты без указания конкретных сообщений
func (c *Consumer) Commit(ctx context.Context) error {
	c.logger.Debug(ctx, "committing current offsets")
	return c.reader.CommitMessages(ctx)
}

// Stop graceful остановка consumer
func (c *Consumer) Stop(ctx context.Context) error {
	c.logger.Info(ctx, "stopping kafka consumer")

	// Сигналим о остановке
	close(c.stopCh)

	// Даем время на завершение текущего FetchMessage
	time.Sleep(100 * time.Millisecond)

	// Коммитим текущие оффсеты
	if err := c.reader.CommitMessages(ctx); err != nil {
		c.logger.Error(ctx, "error committing on shutdown", "error", err)
	}

	// Закрываем reader
	if err := c.reader.Close(); err != nil {
		c.logger.Error(ctx, "error closing reader", "error", err)
		return fmt.Errorf("error closing reader: %w", err)
	}

	c.logger.Info(ctx, "kafka consumer stopped")
	return nil
}

// GetReader возвращает внутренний reader (для продвинутого использования)
func (c *Consumer) GetReader() *kafka.Reader {
	return c.reader
}
