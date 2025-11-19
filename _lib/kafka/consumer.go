package kafkalib

import (
	"context"
	"fmt"
	"time"

	loggerlib "github.com/Burmachine/MSA/lib/logger"
	"github.com/Burmachine/MSA/lib/metrics"
	"github.com/segmentio/kafka-go"
)

type ConsumerConfig struct {
	Brokers        []string
	GroupID        string
	Topic          string
	CommitInterval time.Duration
	ManualCommit   bool
	StartOffset    int64
}

type MessageHandler func(ctx context.Context, msg kafka.Message) error

type Consumer struct {
	reader  *kafka.Reader
	logger  *loggerlib.Logger
	metrics *metrics.Metrics
	handler MessageHandler
	stopCh  chan struct{}
	config  ConsumerConfig
}

func NewConsumer(cfg ConsumerConfig, handler MessageHandler, log *loggerlib.Logger, m *metrics.Metrics) *Consumer {
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
		MaxBytes:    10e6,
		StartOffset: cfg.StartOffset,
	}

	if !cfg.ManualCommit {
		readerCfg.CommitInterval = cfg.CommitInterval
	}

	reader := kafka.NewReader(readerCfg)

	return &Consumer{
		reader:  reader,
		logger:  log,
		metrics: m,
		handler: handler,
		stopCh:  make(chan struct{}),
		config:  cfg,
	}
}

// recordConsumeMetrics записывает метрики обработки сообщения
func (c *Consumer) recordConsumeMetrics(err error, duration time.Duration) {
	if c.metrics == nil {
		return
	}

	if err != nil {
		c.metrics.KafkaMessagesErrors.WithLabelValues(
			c.config.Topic,
			c.config.GroupID,
			"handler_error",
		).Inc()
	} else {
		c.metrics.KafkaMessagesConsumed.WithLabelValues(
			c.config.Topic,
			c.config.GroupID,
		).Inc()
	}

	c.metrics.KafkaBatchDuration.WithLabelValues(
		c.config.Topic,
		c.config.GroupID,
	).Observe(duration.Seconds())
}

// recordFetchError записывает ошибку fetch
func (c *Consumer) recordFetchError() {
	if c.metrics == nil {
		return
	}

	c.metrics.KafkaMessagesErrors.WithLabelValues(
		c.config.Topic,
		c.config.GroupID,
		"fetch_error",
	).Inc()
}

// updateLag обновляет метрику lag
func (c *Consumer) updateLag() {
	if c.metrics == nil {
		return
	}

	stats := c.reader.Stats()
	c.metrics.KafkaConsumerLag.WithLabelValues(
		c.config.Topic,
		fmt.Sprintf("%d", stats.Partition),
		c.config.GroupID,
	).Set(float64(stats.Lag))
}

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
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					c.logger.Info(ctx, "context cancelled or deadline exceeded")
					return err
				}

				c.recordFetchError()
				c.logger.Error(ctx, "error fetching message", "error", err)
				time.Sleep(time.Second)
				continue
			}

			c.logger.Debug(ctx, "received message",
				"topic", msg.Topic,
				"partition", msg.Partition,
				"offset", msg.Offset,
				"key", string(msg.Key),
			)

			start := time.Now()
			err = c.handler(ctx, msg)

			c.recordConsumeMetrics(err, time.Since(start))
			c.updateLag()

			if err != nil {
				c.logger.Error(ctx, "error handling message",
					"error", err,
					"key", string(msg.Key),
					"offset", msg.Offset,
					"partition", msg.Partition,
				)
				continue
			}

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

func (c *Consumer) Commit(ctx context.Context) error {
	c.logger.Debug(ctx, "committing current offsets")
	return c.reader.CommitMessages(ctx)
}

func (c *Consumer) Stop(ctx context.Context) error {
	c.logger.Info(ctx, "stopping kafka consumer")

	close(c.stopCh)
	time.Sleep(100 * time.Millisecond)

	if err := c.reader.CommitMessages(ctx); err != nil {
		c.logger.Error(ctx, "error committing on shutdown", "error", err)
	}

	if err := c.reader.Close(); err != nil {
		c.logger.Error(ctx, "error closing reader", "error", err)
		return fmt.Errorf("error closing reader: %w", err)
	}

	c.logger.Info(ctx, "kafka consumer stopped")
	return nil
}

func (c *Consumer) GetReader() *kafka.Reader {
	return c.reader
}
