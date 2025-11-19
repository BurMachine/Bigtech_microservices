package kafkalib

import (
	"context"
	"fmt"
	"time"

	loggerlib "github.com/Burmachine/MSA/lib/logger"
	"github.com/Burmachine/MSA/lib/metrics"
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
	Compression  kafka.CompressionCodec
}

type Producer struct {
	writer  *kafka.Writer
	logger  *loggerlib.Logger
	metrics *metrics.Metrics
}

func NewProducer(cfg Config, log *loggerlib.Logger, m *metrics.Metrics) *Producer {
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
		writer:  writer,
		logger:  log,
		metrics: m,
	}
}

// recordPublishMetrics записывает метрики публикации
func (p *Producer) recordPublishMetrics(messageCount int, err error, duration time.Duration) {
	if p.metrics == nil {
		return
	}

	if err != nil {
		p.metrics.KafkaMessagesErrors.WithLabelValues(
			p.writer.Topic,
			"producer",
			"write_error",
		).Add(float64(messageCount))
	} else {
		p.metrics.KafkaMessagesConsumed.WithLabelValues(
			p.writer.Topic,
			"producer",
		).Add(float64(messageCount))
	}

	p.metrics.KafkaBatchDuration.WithLabelValues(
		p.writer.Topic,
		"producer",
	).Observe(duration.Seconds())
}

func (p *Producer) PublishMessage(ctx context.Context, key, value []byte) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Kafka/Publish")
	defer span.Finish()

	span.SetTag("messaging.system", "kafka")
	span.SetTag("messaging.destination", p.writer.Topic)
	span.SetTag("messaging.operation", "publish")
	span.SetTag("messaging.message_id", string(key))
	ext.SpanKindProducer.Set(span)

	start := time.Now()

	message := kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now(),
	}

	err := p.writer.WriteMessages(ctx, message)

	p.recordPublishMetrics(1, err, time.Since(start))

	if err != nil {
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

	span, ctx := opentracing.StartSpanFromContext(ctx, "Kafka/PublishBatch")
	defer span.Finish()

	span.SetTag("messaging.system", "kafka")
	span.SetTag("messaging.destination", p.writer.Topic)
	span.SetTag("messaging.operation", "publish_batch")
	span.SetTag("messaging.batch_size", len(messages))
	ext.SpanKindProducer.Set(span)

	for i := range messages {
		if messages[i].Time.IsZero() {
			messages[i].Time = time.Now()
		}
	}

	start := time.Now()

	err := p.writer.WriteMessages(ctx, messages...)

	p.recordPublishMetrics(len(messages), err, time.Since(start))

	if err != nil {
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

func (p *Producer) Flush(ctx context.Context) error {
	p.logger.Info(ctx, "flushing kafka producer")
	return nil
}

func (p *Producer) Close() error {
	ctx := context.Background()
	p.logger.Info(ctx, "closing kafka producer")

	if err := p.writer.Close(); err != nil {
		p.logger.Error(ctx, "error closing kafka producer", "error", err)
		return err
	}

	p.logger.Info(ctx, "kafka producer closed")
	return nil
}

func (p *Producer) Stats() kafka.WriterStats {
	return p.writer.Stats()
}
