package outbox

import (
	"context"
	"time"

	"github.com/BurMachine/Bigtech_microservices/social/internal/app/models"
	"github.com/google/martian/log"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func (p *Processor) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute) // polling interval
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			span, batchCtx := opentracing.StartSpanFromContext(ctx, "Outbox/ProcessBatch")
			//batchCtx := opentracing.ContextWithSpan(ctx, span)

			span.SetTag("component", "outbox-worker")

			err := p.tm.RunReadCommitted(batchCtx, func(txCtx context.Context) error {
				events, err := p.outboxRepo.GetPendingEvents(txCtx, 100) // батч лимит
				if err != nil {
					return err // или log и continue
				}
				if len(events) == 0 {
					return nil // пусто — выход
				}

				// Конвертируем в []*models.Event (если нужно; предполагаем []models.Event уже)
				var eventPtrs []*models.Event
				for i := range events {
					eventPtrs = append(eventPtrs, &events[i])
				}

				succeeded, failed, handleErr := p.eventHandler.HandleBatch(txCtx, eventPtrs)
				now := time.Now()

				// Обработка успешных
				for _, id := range succeeded {
					if err := p.outboxRepo.MarkPublished(txCtx, id, now); err != nil {
						return err
					}
				}

				// Обработка неудачных (ретрей)
				for _, id := range failed {
					// Найти событие по ID для retry_count
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
						return err
					}
				}

				return handleErr // если общая err — rollback
			})
			if err != nil {
				log.Errorf("%v failed to process outbox events: %w", err)
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

	before := time.Now().Add(-7 * 24 * time.Hour)

	err := p.outboxRepo.DeleteOldPublished(ctx, before)
	if err != nil {
		ext.Error.Set(span, true)
		span.LogKV("error", err.Error())
	}

	return err
}
