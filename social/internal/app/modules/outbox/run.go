package outbox

import (
	"context"
	"time"

	"github.com/BurMachine/Bigtech_microservices/social/internal/app/models"
	"github.com/google/martian/log"
)

func (p *Processor) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute) // polling interval
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := p.tm.RunReadCommitted(ctx, func(txCtx context.Context) error {
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
		}
	}
}

func (p *Processor) Cleanup(ctx context.Context) error {
	before := time.Now().Add(-7 * 24 * time.Hour) // retention 7 дней
	return p.outboxRepo.DeleteOldPublished(ctx, before)
}
