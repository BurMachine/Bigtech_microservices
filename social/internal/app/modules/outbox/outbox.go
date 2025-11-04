package outbox

import (
	"context"
	"time"

	"github.com/BurMachine/Bigtech_microservices/social/internal/app/models"
	"github.com/google/uuid"
)

// Ports
type (
	Repo interface {
		// Чтение pending событий (неопубликованных, готовых к обработке)
		GetPendingEvents(ctx context.Context, limit int) ([]models.Event, error)

		// Отметка как опубликовано (успех)
		MarkPublished(ctx context.Context, id uuid.UUID, publishedAt time.Time) error

		// Отметка как failed (для ретрая)
		MarkFailed(ctx context.Context, id uuid.UUID, retryCount int, nextAttempt time.Time, lastError string) error

		// Очистка старых опубликованных событий
		DeleteOldPublished(ctx context.Context, before time.Time) error
	}

	EventsHandler interface {
		HandleBatch(ctx context.Context, events []*models.Event) (succeeded []uuid.UUID, failed []uuid.UUID, err error)
	}

	TransactionManager interface {
		RunReadCommitted(ctx context.Context, f func(ctx context.Context) error) error
	}
)

type Processor struct {
	outboxRepo   Repo
	eventHandler EventsHandler
	tm           TransactionManager
}

func NewProcessor(outboxRepo Repo, eventHandler EventsHandler, tm TransactionManager) *Processor {
	return &Processor{
		outboxRepo:   outboxRepo,
		eventHandler: eventHandler,
		tm:           tm,
	}
}
