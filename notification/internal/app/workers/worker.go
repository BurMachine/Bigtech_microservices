package workers

import (
	"context"
	"log"
	"time"

	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/handler"
	"github.com/BurMachine/Bigtech_microservices/notification/internal/app/inbox_repo"
	"github.com/segmentio/kafka-go"
)

type Worker struct {
	repo         inbox_repo.InboxRepo
	handler      handler.Handler
	pollInterval time.Duration // e.g. 5s
	maxAttempts  int           // e.g. 5
	batchSize    int           // e.g. 50
}

func NewWorker(repo inbox_repo.InboxRepo, h handler.Handler) *Worker {
	return &Worker{
		repo:         repo,
		handler:      h,
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
	// SELECT ids для обработки
	ids, err := w.repo.SelectForProcessing(ctx, w.maxAttempts, w.batchSize)
	if err != nil {
		log.Printf("select failed: %v", err)
		return
	}
	if len(ids) == 0 {
		return
	}

	for _, id := range ids {
		// Для каждого: транзакция
		err := w.processOne(ctx, id)
		if err != nil {
			log.Printf("process id=%s failed: %v", id, err)
		}
	}
}

func (w *Worker) processOne(ctx context.Context, id string) error {
	// Здесь: BEGIN транзакция в БД (в репо.UpdateStatus — сделать транзакционным)

	// 1. Update to 'processing', attempts++
	// (В репо: SELECT attempts FROM ... FOR UPDATE; затем UPDATE)
	// Для простоты предполагаем, что UpdateStatus делает всё в tx

	// 2. Выполнить handler (идемпотентно)
	// Но handler принимает *kafka.Message — нужно восстановить из БД?
	// Да: В реальности — добавить метод repo.GetMessage(id) -> InboxMessage, затем создать fake kafka.Message из него.
	// Для примера: Предположим, мы имеем msg (payload из БД)

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
