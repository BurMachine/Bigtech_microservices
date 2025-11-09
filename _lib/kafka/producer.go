package kafkalib

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Config struct {
	Brokers      []string
	Topic        string
	BatchSize    int
	BatchTimeout time.Duration
	MaxAttempts  int
}

type Producer struct {
	writer *kafka.Writer
	logger *zap.Logger
}

func NewProducer(cfg Config, logger *zap.Logger) *Producer {
	// Применяем дефолты
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 100
	}
	if cfg.BatchTimeout == 0 {
		cfg.BatchTimeout = 10 * time.Millisecond
	}
	if cfg.MaxAttempts == 0 {
		cfg.MaxAttempts = 3
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.Hash{},
		BatchSize:    cfg.BatchSize,
		BatchTimeout: cfg.BatchTimeout,
		MaxAttempts:  cfg.MaxAttempts,
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	return &Producer{
		writer: writer,
		logger: logger,
	}
}

func (p *Producer) PublishMessage(ctx context.Context, key, value []byte) error {
	message := kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, message); err != nil {
		p.logger.Error("failed to publish message",
			zap.Error(err),
			zap.String("key", string(key)),
		)
		return fmt.Errorf("failed to write message: %w", err)
	}

	p.logger.Debug("message published",
		zap.String("topic", p.writer.Topic),
		zap.String("key", string(key)),
	)

	return nil
}

func (p *Producer) PublishBatch(ctx context.Context, messages []kafka.Message) error {
	if len(messages) == 0 {
		return nil
	}

	// Добавляем timestamp если не указан
	for i := range messages {
		if messages[i].Time.IsZero() {
			messages[i].Time = time.Now()
		}
	}

	if err := p.writer.WriteMessages(ctx, messages...); err != nil {
		p.logger.Error("failed to publish batch",
			zap.Error(err),
			zap.Int("count", len(messages)),
		)
		return fmt.Errorf("failed to write batch: %w", err)
	}

	p.logger.Debug("batch published",
		zap.String("topic", p.writer.Topic),
		zap.Int("count", len(messages)),
	)

	return nil
}

// Flush принудительно отправляет все буферизованные сообщения
func (p *Producer) Flush(ctx context.Context) error {
	p.logger.Info("flushing kafka producer")

	// kafka-go Writer не имеет явного Flush, но Close делает это автоматически
	// Можно отправить пустой батч для синхронизации
	return nil
}

// Close корректное закрытие с автоматическим flush
func (p *Producer) Close() error {
	p.logger.Info("closing kafka producer")

	// Writer.Close() автоматически делает flush всех буферизованных сообщений
	if err := p.writer.Close(); err != nil {
		p.logger.Error("error closing kafka producer", zap.Error(err))
		return err
	}

	p.logger.Info("kafka producer closed")
	return nil
}

// Stats возвращает статистику producer'а
func (p *Producer) Stats() kafka.WriterStats {
	return p.writer.Stats()
}
