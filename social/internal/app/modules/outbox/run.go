package outbox

import (
	"context"
	"time"

	"github.com/BurMachine/Bigtech_microservices/social/internal/app/models"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// recordBatchMetrics записывает метрики обработки батча
func (p *Processor) recordBatchMetrics(processedCount, failedCount int, err error, duration time.Duration) {
	if p.metrics == nil {
		return
	}

	// Количество обработанных батчей
	if err != nil {
		p.metrics.WorkerBatchProcessed.WithLabelValues("outbox", "error").Inc()
		p.metrics.WorkerBatchErrors.WithLabelValues("outbox", "processing_error").Inc()
	} else {
		p.metrics.WorkerBatchProcessed.WithLabelValues("outbox", "success").Inc()
	}

	// Длительность обработки батча
	p.metrics.WorkerBatchDuration.WithLabelValues("outbox").Observe(duration.Seconds())

	// Обновляем размер очереди (приблизительно)
	// В реальности нужно получать из БД: SELECT COUNT(*) FROM outbox WHERE status = 'pending'
	// Для простоты используем количество необработанных событий в последнем батче
	remainingEvents := failedCount
	p.metrics.WorkerQueueSize.WithLabelValues("outbox").Set(float64(remainingEvents))
}

func (p *Processor) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	p.logger.Info(ctx, "outbox worker started", "interval", "1m")

	for {
		select {
		case <-ctx.Done():
			p.logger.Info(ctx, "outbox worker stopped")
			return
		case <-ticker.C:
			span, batchCtx := opentracing.StartSpanFromContext(ctx, "Outbox/ProcessBatch")
			span.SetTag("component", "outbox-worker")

			start := time.Now() // ← Засекаем время

			var processedCount, failedCount int

			err := p.tm.RunReadCommitted(batchCtx, func(txCtx context.Context) error {
				events, err := p.outboxRepo.GetPendingEvents(txCtx, 100)
				if err != nil {
					p.logger.Error(txCtx, "failed to get pending events", "error", err)
					return err
				}

				if len(events) == 0 {
					p.logger.Debug(txCtx, "no pending events")
					return nil
				}

				p.logger.Info(txCtx, "processing outbox batch", "count", len(events))

				var eventPtrs []*models.Event
				for i := range events {
					eventPtrs = append(eventPtrs, &events[i])
				}

				succeeded, failed, handleErr := p.eventHandler.HandleBatch(txCtx, eventPtrs)
				now := time.Now()

				processedCount = len(succeeded) // ← Считаем успешные
				failedCount = len(failed)       // ← Считаем неудачные

				// Обработка успешных
				for _, id := range succeeded {
					if err := p.outboxRepo.MarkPublished(txCtx, id, now); err != nil {
						p.logger.Error(txCtx, "failed to mark published", "event_id", id, "error", err)
						return err
					}
				}

				// Обработка неудачных (ретрей)
				for _, id := range failed {
					var ev models.Event
					for _, e := range events {
						if e.ID == id {
							ev = e
							break
						}
					}

					retryCount := ev.RetryCount + 1
					backoff := time.Minute * time.Duration(1<<retryCount) // exponential: 1m, 2m, 4m...
					nextAttempt := now.Add(backoff)
					lastErrMsg := "batch processing error"
					if handleErr != nil {
						lastErrMsg = handleErr.Error()
					}

					if err := p.outboxRepo.MarkFailed(txCtx, id, retryCount, nextAttempt, lastErrMsg); err != nil {
						p.logger.Error(txCtx, "failed to mark failed", "event_id", id, "error", err)
						return err
					}
				}

				if handleErr != nil {
					p.logger.Error(txCtx, "batch processing error", "error", handleErr)
				}

				return handleErr
			})

			// ← Записываем метрики
			duration := time.Since(start)
			p.recordBatchMetrics(processedCount, failedCount, err, duration)

			if err != nil {
				p.logger.Error(batchCtx, "failed to process outbox events", "error", err)
				ext.Error.Set(span, true)
				span.LogKV("error", err.Error())
			} else {
				p.logger.Info(batchCtx, "outbox batch processed",
					"processed", processedCount,
					"failed", failedCount,
					"duration_ms", duration.Milliseconds(),
				)
			}

			span.Finish()
		}
	}
}

func (p *Processor) Cleanup(ctx context.Context) error {
	span := opentracing.StartSpan("Outbox/Cleanup")
	defer span.Finish()

	ctx = opentracing.ContextWithSpan(ctx, span)
	span.SetTag("component", "outbox-worker")

	p.logger.Info(ctx, "running outbox cleanup")

	before := time.Now().Add(-7 * 24 * time.Hour)

	err := p.outboxRepo.DeleteOldPublished(ctx, before)
	if err != nil {
		p.logger.Error(ctx, "outbox cleanup failed", "error", err)
		ext.Error.Set(span, true)
		span.LogKV("error", err.Error())
	} else {
		p.logger.Info(ctx, "outbox cleanup completed", "before", before)
	}

	return err
}
