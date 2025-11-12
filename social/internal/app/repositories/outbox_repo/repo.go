package outbox_repo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/BurMachine/Bigtech_microservices/social/internal/app/models"
	"github.com/Burmachine/MSA/lib/postgreslib"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

const (
	tableOutboxEvents = "outbox_events"

	colID            = "id"
	colEventID       = "event_id"
	colAggregateType = "aggregate_type"
	colAggregateID   = "aggregate_id"
	colEventType     = "event_type"
	colPayload       = "payload"
	colTopic         = "topic"
	colPartitionKey  = "partition_key"
	colPublishedAt   = "published_at"
	colRetryCount    = "retry_count"
	colNextAttemptAt = "next_attempt_at"
	colLastError     = "last_error"
	colCreatedAt     = "created_at"
	colUpdatedAt     = "updated_at"
)

const (
	aggregateTypeFriendRequest = "friend_request"
	topicSocialEvents          = "test"

	eventTypeFriendRequestCreated  = "social.friend.request.created"
	eventTypeFriendRequestAccepted = "social.friend.request.accepted"
	eventTypeFriendRequestDeclined = "social.friend.request.declined"
)

type OutboxRepo struct {
	db postgreslib.QueryEngineProvider
	qb squirrel.StatementBuilderType
}

// NewRepository конструктор Repository
func NewRepository(p postgreslib.QueryEngineProvider) *OutboxRepo {
	return &OutboxRepo{
		db: p,
		qb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *OutboxRepo) SaveFriendsRequestCreated(ctx context.Context, request models.FriendRequest) error {
	const api = "outbox_repo.OutboxRepo.SaveFriendsRequestCreated"

	// Подготовка payload
	payload := FriendRequestCreatedPayload{
		RequestID:  request.RequestID,
		FromUserID: request.FromUserID,
		ToUserID:   request.ToUserID,
		CreatedAt:  request.CreatedAt,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("%s: marshal payload: %w", api, err)
	}

	// Формирование запроса
	now := time.Now().UTC()
	eventID := uuid.New().String()

	qb := r.qb.Insert(tableOutboxEvents).
		Columns(
			colEventID,
			colAggregateType,
			colAggregateID,
			colEventType,
			colPayload,
			colTopic,
			colPartitionKey,
			colCreatedAt,
		).
		Values(
			eventID,
			aggregateTypeFriendRequest,
			request.RequestID,
			eventTypeFriendRequestCreated,
			payloadJSON,
			topicSocialEvents,
			request.ToUserID, // партиционирование по получателю
			now,
		)

	conn := r.db.GetQueryEngine(ctx)
	if _, err := conn.Execx(ctx, qb); err != nil {
		return fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	return nil
}

func (r *OutboxRepo) SaveFriendsRequestUpdated(ctx context.Context, request models.FriendRequest) error {
	const api = "outbox_repo.OutboxRepo.SaveFriendsRequestUpdated"

	// Определяем тип события на основе статуса
	eventType := eventTypeFriendRequestAccepted
	if request.Status == "DECLINED" {
		eventType = eventTypeFriendRequestDeclined
	}

	// Подготовка payload
	payload := FriendRequestUpdatedPayload{
		RequestID:  request.RequestID,
		FromUserID: request.FromUserID,
		ToUserID:   request.ToUserID,
		Status:     request.Status,
		UpdatedAt:  request.UpdatedAt,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("%s: marshal payload: %w", api, err)
	}

	// Формирование запроса
	now := time.Now().UTC()
	eventID := uuid.New().String()

	qb := r.qb.Insert(tableOutboxEvents).
		Columns(
			colEventID,
			colAggregateType,
			colAggregateID,
			colEventType,
			colPayload,
			colTopic,
			colPartitionKey,
			colCreatedAt,
		).
		Values(
			eventID,
			aggregateTypeFriendRequest,
			request.RequestID,
			eventType,
			payloadJSON,
			topicSocialEvents,
			request.FromUserID, // партиционирование по отправителю (он получит уведомление)
			now,
		)

	conn := r.db.GetQueryEngine(ctx)
	if _, err := conn.Execx(ctx, qb); err != nil {
		return fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	return nil
}

func (r *OutboxRepo) GetPendingEvents(ctx context.Context, limit int) ([]models.Event, error) {
	const api = "outbox_repo.OutboxRepo.GetPendingEvents"

	qb := r.qb.Select(
		colID,
		colEventID,
		colAggregateType,
		colAggregateID,
		colEventType,
		colPayload,
		colTopic,
		colPartitionKey,
		colPublishedAt,
		colRetryCount,
		colNextAttemptAt,
		colLastError,
		colCreatedAt,
		colUpdatedAt,
	).
		From(tableOutboxEvents).
		Where(squirrel.And{
			squirrel.Eq{colPublishedAt: nil},
			squirrel.LtOrEq{colNextAttemptAt: time.Now().UTC()},
		}).
		OrderBy(colCreatedAt + " ASC").
		Limit(uint64(limit)).
		Suffix("FOR UPDATE SKIP LOCKED")

	conn := r.db.GetQueryEngine(ctx)
	var events []models.Event
	if err := conn.Selectx(ctx, &events, qb); err != nil {
		return nil, fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	return events, nil
}

func (r *OutboxRepo) MarkPublished(ctx context.Context, id uuid.UUID, publishedAt time.Time) error {
	const api = "outbox_repo.OutboxRepo.MarkPublished"

	qb := r.qb.Update(tableOutboxEvents).
		Set(colPublishedAt, publishedAt.UTC()).
		Set(colRetryCount, 0).
		Set(colLastError, nil).
		Set(colUpdatedAt, time.Now().UTC()).
		Where(squirrel.Eq{colID: id})

	conn := r.db.GetQueryEngine(ctx)
	result, err := conn.Execx(ctx, qb)
	if err != nil {
		return fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("%s: no rows affected", api)
	}

	return nil
}

func (r *OutboxRepo) MarkFailed(ctx context.Context, id uuid.UUID, retryCount int, nextAttempt time.Time, lastError string) error {
	const api = "outbox_repo.OutboxRepo.MarkFailed"

	qb := r.qb.Update(tableOutboxEvents).
		Set(colRetryCount, retryCount).
		Set(colNextAttemptAt, nextAttempt.UTC()).
		Set(colLastError, lastError).
		Set(colUpdatedAt, time.Now().UTC()).
		Where(squirrel.Eq{colID: id})

	conn := r.db.GetQueryEngine(ctx)
	result, err := conn.Execx(ctx, qb)
	if err != nil {
		return fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("%s: no rows affected", api)
	}

	return nil
}

func (r *OutboxRepo) DeleteOldPublished(ctx context.Context, before time.Time) error {
	const api = "outbox_repo.OutboxRepo.DeleteOldPublished"

	qb := r.qb.Delete(tableOutboxEvents).
		Where(squirrel.And{
			squirrel.NotEq{colPublishedAt: nil},
			squirrel.Lt{colPublishedAt: before.UTC()},
		})

	conn := r.db.GetQueryEngine(ctx)
	if _, err := conn.Execx(ctx, qb); err != nil {
		return fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	return nil
}
