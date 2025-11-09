package kafkalib

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type ConsumerConfig struct {
	Brokers        []string
	GroupID        string
	Topic          string
	CommitInterval time.Duration // Интервал автокоммита (0 = manual commit)
	ManualCommit   bool          // Если true - не коммитить автоматически
}

type MessageHandler func(ctx context.Context, msg kafka.Message) error

type Consumer struct {
	reader  *kafka.Reader
	logger  *zap.Logger
	handler MessageHandler
	stopCh  chan struct{}
	config  ConsumerConfig
}

func NewConsumer(cfg ConsumerConfig, handler MessageHandler, logger *zap.Logger) *Consumer {
	// Дефолты
	if cfg.CommitInterval == 0 && !cfg.ManualCommit {
		cfg.CommitInterval = time.Second
	}

	readerCfg := kafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		GroupID:     cfg.GroupID,
		Topic:       cfg.Topic,
		MinBytes:    1,
		MaxBytes:    10e6, // 10MB
		StartOffset: kafka.FirstOffset,
	}

	// Если manual commit - не устанавливаем CommitInterval
	if !cfg.ManualCommit {
		readerCfg.CommitInterval = cfg.CommitInterval
	}

	reader := kafka.NewReader(readerCfg)

	return &Consumer{
		reader:  reader,
		logger:  logger,
		handler: handler,
		stopCh:  make(chan struct{}),
		config:  cfg,
	}
}

// Start запускает consumer (вызывать в горутине)
func (c *Consumer) Start(ctx context.Context) error {
	c.logger.Info("starting kafka consumer",
		zap.String("topic", c.reader.Config().Topic),
		zap.Bool("manualCommit", c.config.ManualCommit),
	)

	for {
		// Убираем select с default - FetchMessage сам блокируется
		select {
		case <-ctx.Done():
			c.logger.Info("consumer context cancelled")
			return ctx.Err()
		case <-c.stopCh:
			c.logger.Info("consumer stopped")
			return nil
		default:
			// FetchMessage блокируется до получения сообщения или ошибки
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					c.logger.Info("context cancelled or deadline exceeded")
					return err
				}
				c.logger.Error("error fetching message", zap.Error(err))
				time.Sleep(time.Second) // backoff при ошибке
				continue
			}

			c.logger.Debug("received message",
				zap.String("topic", msg.Topic),
				zap.Int("partition", msg.Partition),
				zap.Int64("offset", msg.Offset),
			)

			// Обрабатываем сообщение
			if err := c.handler(ctx, msg); err != nil {
				c.logger.Error("error handling message",
					zap.Error(err),
					zap.String("key", string(msg.Key)),
					zap.Int64("offset", msg.Offset),
				)
				// При ошибке НЕ коммитим - Kafka retry'нет
				continue
			}

			// Коммитим только если не manual commit
			if !c.config.ManualCommit {
				if err := c.reader.CommitMessages(ctx, msg); err != nil {
					c.logger.Error("error committing message", zap.Error(err))
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

	c.logger.Debug("committing messages", zap.Int("count", len(msgs)))
	return c.reader.CommitMessages(ctx, msgs...)
}

// Commit коммитит текущие оффсеты без указания конкретных сообщений
func (c *Consumer) Commit(ctx context.Context) error {
	c.logger.Debug("committing current offsets")
	return c.reader.CommitMessages(ctx)
}

// Stop graceful остановка consumer
func (c *Consumer) Stop(ctx context.Context) error {
	c.logger.Info("stopping kafka consumer")

	// Сигналим о остановке
	close(c.stopCh)

	// Даем время на завершение текущего FetchMessage
	time.Sleep(100 * time.Millisecond)

	// Коммитим текущие оффсеты
	if err := c.reader.CommitMessages(ctx); err != nil {
		c.logger.Error("error committing on shutdown", zap.Error(err))
	}

	// Закрываем reader
	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("error closing reader: %w", err)
	}

	c.logger.Info("kafka consumer stopped")
	return nil
}

// GetReader возвращает внутренний reader (для продвинутого использования)
func (c *Consumer) GetReader() *kafka.Reader {
	return c.reader
}
