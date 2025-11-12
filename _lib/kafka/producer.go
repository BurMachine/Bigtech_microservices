package kafkalib

import (
	"context"
	"fmt"
	"time"

	"github.com/Burmachine/MSA/lib/logger"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/segmentio/kafka-go"
)

type Config struct {
	Brokers      []string
	Topic        string
	BatchSize    int
	BatchTimeout time.Duration
	MaxAttempts  int
	Compression  kafka.CompressionCodec // snappy, gzip, lz4, zstd
}

type Producer struct {
	writer *kafka.Writer
	logger *loggerlib.Logger
}

func NewProducer(cfg Config, log *loggerlib.Logger) *Producer {
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
		logger: log,
	}
}

func (p *Producer) PublishMessage(ctx context.Context, key, value []byte) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Kafka/Publish")
	defer span.Finish()

	// Теги
	span.SetTag("messaging.system", "kafka")
	span.SetTag("messaging.destination", p.writer.Topic)
	span.SetTag("messaging.operation", "publish")
	span.SetTag("messaging.message_id", string(key))
	ext.SpanKindProducer.Set(span)

	message := kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, message); err != nil {
		p.logger.Error(ctx, "failed to publish message",
			"error", err,
			"key", string(key),
			"topic", p.writer.Topic,
		)
		return fmt.Errorf("failed to write message: %w", err)
	}

	p.logger.Debug(ctx, "message published",
		"topic", p.writer.Topic,
		"key", string(key),
	)

	return nil
}

func (p *Producer) PublishBatch(ctx context.Context, messages []kafka.Message) error {
	if len(messages) == 0 {
		return nil
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "Kafka/Publish")
	defer span.Finish()

	// Теги
	span.SetTag("messaging.system", "kafka")
	span.SetTag("messaging.destination", p.writer.Topic)
	span.SetTag("messaging.operation", "publish_batch")
	ext.SpanKindProducer.Set(span)

	// Добавляем timestamp если не указан
	for i := range messages {
		if messages[i].Time.IsZero() {
			messages[i].Time = time.Now()
		}
	}

	if err := p.writer.WriteMessages(ctx, messages...); err != nil {
		p.logger.Error(ctx, "failed to publish batch",
			"error", err,
			"count", len(messages),
			"topic", p.writer.Topic,
		)
		return fmt.Errorf("failed to write batch: %w", err)
	}

	p.logger.Debug(ctx, "batch published",
		"topic", p.writer.Topic,
		"count", len(messages),
	)

	return nil
}

// Flush принудительно отправляет все буферизованные сообщения
func (p *Producer) Flush(ctx context.Context) error {
	p.logger.Info(ctx, "flushing kafka producer")
	// kafka-go Writer не имеет явного Flush
	// Все сообщения отправляются синхронно (Async: false)
	return nil
}

// Close корректное закрытие с автоматическим flush
func (p *Producer) Close() error {
	ctx := context.Background()
	p.logger.Info(ctx, "closing kafka producer")

	// Writer.Close() автоматически делает flush
	if err := p.writer.Close(); err != nil {
		p.logger.Error(ctx, "error closing kafka producer", "error", err)
		return err
	}

	p.logger.Info(ctx, "kafka producer closed")
	return nil
}

// Stats возвращает статистику producer'а
func (p *Producer) Stats() kafka.WriterStats {
	return p.writer.Stats()
}
