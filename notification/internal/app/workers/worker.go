package workers

import (
	"context"
	"time"

	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/handler"
	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/inbox_repo"
	loggerlib "github.com/Burmachine/MSA/lib/logger"
	"github.com/Burmachine/MSA/lib/metrics"
	"github.com/opentracing/opentracing-go"
	"github.com/segmentio/kafka-go"
)

type Worker struct {
	repo         inbox_repo.InboxRepo
	handler      handler.Handler
	logger       *loggerlib.Logger
	metrics      *metrics.Metrics
	pollInterval time.Duration
	maxAttempts  int
	batchSize    int
}

func NewWorker(
	repo inbox_repo.InboxRepo,
	h handler.Handler,
	logger *loggerlib.Logger,
	m *metrics.Metrics,
) *Worker {
	return &Worker{
		repo:         repo,
		handler:      h,
		logger:       logger,
		metrics:      m,
		pollInterval: 5 * time.Second,
		maxAttempts:  5,
		batchSize:    50,
	}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	w.logger.Info(ctx, "inbox worker started", "interval", w.pollInterval)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info(ctx, "inbox worker stopped")
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *Worker) processBatch(ctx context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "InboxWorker")
	span.SetTag("component", "inbox-worker")
	defer span.Finish()

	start := time.Now()

	ids, err := w.repo.SelectForProcessing(ctx, w.maxAttempts, w.batchSize)
	if err != nil {
		w.logger.Error(ctx, "select failed", "error", err)

		// Записываем метрики ошибки
		if w.metrics != nil {
			w.metrics.WorkerBatchProcessed.WithLabelValues("inbox", "error").Inc()
			w.metrics.WorkerBatchErrors.WithLabelValues("inbox", "select_error").Inc()
			w.metrics.WorkerBatchDuration.WithLabelValues("inbox").Observe(time.Since(start).Seconds())
		}
		return
	}

	if len(ids) == 0 {
		w.logger.Debug(ctx, "no pending inbox records")
		return
	}

	w.logger.Info(ctx, "processing inbox batch", "count", len(ids))

	var processedCount, failedCount int

	for _, id := range ids {
		err := w.processOne(ctx, id)
		if err != nil {
			w.logger.Error(ctx, "process failed", "id", id, "error", err)
			failedCount++
		} else {
			processedCount++
		}
	}

	duration := time.Since(start)

	// Записываем метрики
	if w.metrics != nil {
		// Счётчик успешных батчей
		w.metrics.WorkerBatchProcessed.WithLabelValues("inbox", "success").Inc()

		// Время обработки батча
		w.metrics.WorkerBatchDuration.WithLabelValues("inbox").Observe(duration.Seconds())

		// Размер очереди (приблизительно по размеру текущего батча)
		// Если len(ids) == batchSize, значит в очереди ещё есть записи
		// Если len(ids) < batchSize, значит это последние записи
		w.metrics.WorkerQueueSize.WithLabelValues("inbox").Set(float64(len(ids)))
	}

	w.logger.Info(ctx, "inbox batch processed",
		"processed", processedCount,
		"failed", failedCount,
		"duration_ms", duration.Milliseconds(),
	)
}

func (w *Worker) processOne(ctx context.Context, id string) error {
	// TODO: В реальности получить payload из БД
	fakeMsg := &kafka.Message{Value: []byte("payload from DB for id=" + id)}

	if err := w.handler.Handle(ctx, fakeMsg); err != nil {
		return w.repo.UpdateStatus(ctx, id, "failed", 0, err.Error(), nil)
	}

	now := time.Now()
	return w.repo.UpdateStatus(ctx, id, "processed", 0, "", &now)
}
