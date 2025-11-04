package inbox_repo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InboxRepo — интерфейс для операций с inbox_messages таблицей
type InboxRepo interface {
	// BatchInsert: Вставка пачки сообщений с ON CONFLICT DO NOTHING
	// Возвращает ошибку, если не удалось вставить (но игнорирует конфликты)
	BatchInsert(ctx context.Context, msgs []InboxMessage) error

	// SelectForProcessing: SELECT id... FOR UPDATE SKIP LOCKED
	// Возвращает список ID для обработки (batch_size)
	SelectForProcessing(ctx context.Context, maxAttempts int, batchSize int) ([]string, error)

	// UpdateStatus: Обновление status, attempts, last_error, processed_at для одного ID
	// Всё в одной транзакции с бизнес-логикой
	UpdateStatus(ctx context.Context, id string, status string, attempts int, lastError string, processedAt *time.Time) error
}

// InboxMessage — структура для вставки
type InboxMessage struct {
	ID          string
	Topic       string
	Partition   int
	KafkaOffset int64  // Изменено с Offset на KafkaOffset
	Payload     []byte // JSONB или BYTEA
}

// IMPLEMENTATION

type inboxRepo struct {
	db *pgxpool.Pool
}

// NewInboxRepo создает новый репозиторий для работы с inbox сообщениями
func NewInboxRepo(db *pgxpool.Pool) InboxRepo {
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

	for _, msg := range msgs {
		// Генерируем UUID если ID не валидный UUID
		id := msg.ID
		if !isValidUUID(msg.ID) {
			id = uuid.New().String()
		}

		batch.Queue(query, id, msg.Topic, msg.Partition, msg.KafkaOffset, msg.Payload, time.Now())
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

// SelectForProcessing выбирает сообщения для обработки с блокировкой
func (r *inboxRepo) SelectForProcessing(ctx context.Context, maxAttempts int, batchSize int) ([]string, error) {
	query := `
		SELECT id 
		FROM inbox_messages 
		WHERE status IN ('received', 'failed') 
		  AND attempts < $1
		ORDER BY received_at ASC
		FOR UPDATE SKIP LOCKED
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, maxAttempts, batchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to select messages for processing: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan message id: %w", err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return ids, nil
}

// UpdateStatus обновляет статус и метаданные сообщения
func (r *inboxRepo) UpdateStatus(ctx context.Context, id string, status string, attempts int, lastError string, processedAt *time.Time) error {
	query := `
		UPDATE inbox_messages 
		SET status = $1, 
		    attempts = $2, 
		    last_error = $3, 
		    processed_at = $4
		WHERE id = $5
	`

	result, err := r.db.Exec(ctx, query, status, attempts, lastError, processedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("message with id %s not found", id)
	}

	return nil
}
