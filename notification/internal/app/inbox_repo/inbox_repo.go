package inbox_repo

import (
	"context"
	"fmt"
	"time"

	"github.com/Burmachine/MSA/lib/postgreslib"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// InboxRepo — интерфейс для операций с inbox_messages таблицей
type InboxRepo interface {

	// BatchInsert: Вставка пачки сообщений с ON CONFLICT DO NOTHING
	BatchInsert(ctx context.Context, msgs []InboxMessage) error

	// SelectForProcessing: SELECT id... FOR UPDATE SKIP LOCKED
	SelectForProcessing(ctx context.Context, maxAttempts int, batchSize int) ([]string, error)

	// UpdateStatus: Обновление status, attempts, last_error, processed_at для одного ID
	UpdateStatus(ctx context.Context, id string, status string, attempts int, lastError string, processedAt *time.Time) error
}

// InboxMessage — структура для вставки
type InboxMessage struct {
	ID          string
	Topic       string
	Partition   int
	KafkaOffset int64
	Payload     []byte
}

// IMPLEMENTATION

type inboxRepo struct {
	db postgreslib.QueryEngine
}

// NewInboxRepo создает новый репозиторий для работы с inbox сообщениями
func NewInboxRepo(db postgreslib.QueryEngine) InboxRepo {
	return &inboxRepo{db: db}
}

func (r *inboxRepo) BatchInsert(ctx context.Context, msgs []InboxMessage) error {
	if len(msgs) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	query := `
		INSERT INTO inbox_messages (id, topic, partition, kafka_offset, payload, status, received_at)
		VALUES ($1, $2, $3, $4, $5, 'received', $6)
		ON CONFLICT (id) DO NOTHING
	`

	now := time.Now()
	for _, msg := range msgs {
		// Генерируем UUID если ID не валидный UUID
		id := msg.ID
		if !isValidUUID(msg.ID) {
			id = uuid.New().String()
		}

		batch.Queue(query, id, msg.Topic, msg.Partition, msg.KafkaOffset, msg.Payload, now)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("failed to insert message at index %d: %w", i, err)
		}
	}

	return nil
}

// isValidUUID проверяет, является ли строка валидным UUID
func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

// inboxMessageRow структура для сканирования результата
type inboxMessageRow struct {
	ID string `db:"id"`
}

// SelectForProcessing выбирает сообщения для обработки с блокировкой
func (r *inboxRepo) SelectForProcessing(ctx context.Context, maxAttempts int, batchSize int) ([]string, error) {
	// Используем squirrel для построения запроса
	sqlizer := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("id").
		From("inbox_messages").
		Where(sq.Eq{"status": []string{"received", "failed"}}).
		Where(sq.Lt{"attempts": maxAttempts}).
		OrderBy("received_at ASC").
		Limit(uint64(batchSize)).
		Suffix("FOR UPDATE SKIP LOCKED")

	var rows []inboxMessageRow
	if err := r.db.Selectx(ctx, &rows, sqlizer); err != nil {
		return nil, fmt.Errorf("failed to select messages for processing: %w", err)
	}

	// Извлекаем только ID
	ids := make([]string, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}

	return ids, nil
}

// UpdateStatus обновляет статус и метаданные сообщения
func (r *inboxRepo) UpdateStatus(ctx context.Context, id string, status string, attempts int, lastError string, processedAt *time.Time) error {
	// Используем squirrel для UPDATE
	sqlizer := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Update("inbox_messages").
		Set("status", status).
		Set("attempts", attempts).
		Set("last_error", lastError).
		Set("processed_at", processedAt).
		Where(sq.Eq{"id": id})

	result, err := r.db.Execx(ctx, sqlizer)
	if err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("message with id %s not found", id)
	}

	return nil
}
