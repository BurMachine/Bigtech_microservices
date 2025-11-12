package workers

import (
	"context"
	"time"

	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/handler"
	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/inbox_repo"
	loggerlib "github.com/Burmachine/MSA/lib/logger"
	"github.com/opentracing/opentracing-go"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Worker struct {
	repo         inbox_repo.InboxRepo
	handler      handler.Handler
	logger       *loggerlib.Logger // Добавили только logger
	pollInterval time.Duration
	maxAttempts  int
	batchSize    int
}

func NewWorker(repo inbox_repo.InboxRepo, h handler.Handler, logger *loggerlib.Logger) *Worker {
	return &Worker{
		repo:         repo,
		handler:      h,
		logger:       logger, // Добавили logger
		pollInterval: 5 * time.Second,
		maxAttempts:  5,
		batchSize:    50,
	}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
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

	ids, err := w.repo.SelectForProcessing(ctx, w.maxAttempts, w.batchSize)
	if err != nil {
		w.logger.Error(ctx, "select failed", zap.Error(err)) // Заменили log.Printf на logger
		return
	}
	if len(ids) == 0 {
		return
	}

	for _, id := range ids {
		// Для каждого: транзакция
		err := w.processOne(ctx, id)
		if err != nil {
			w.logger.Error(ctx, "process failed", zap.String("id", id), zap.Error(err)) // Заменили log.Printf на logger
		}
	}
}

func (w *Worker) processOne(ctx context.Context, id string) error {
	// Fake: Заглушка — в реальности fetch payload
	fakeMsg := &kafka.Message{Value: []byte("payload from DB for id=" + id)} // Заменить на real fetch

	if err := w.handler.Handle(ctx, fakeMsg); err != nil {
		// Update to 'failed', last_error
		return w.repo.UpdateStatus(ctx, id, "failed", 0 /* attempts from DB +1 */, err.Error(), nil)
	}

	// 3. Успех: Update to 'processed', processed_at=now()
	now := time.Now()
	return w.repo.UpdateStatus(ctx, id, "processed", 0 /* attempts from DB +1 */, "", &now)
}
