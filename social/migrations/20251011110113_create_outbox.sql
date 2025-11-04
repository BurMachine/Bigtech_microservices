-- +goose Up
-- +goose StatementBegin
CREATE TABLE outbox_events (
    -- Идентификация события
                               id                  UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
                               event_id            UUID            NOT NULL UNIQUE,  -- для идемпотентности (дедупликация)

    -- Метаданные агрегата
                               aggregate_type      VARCHAR(50)     NOT NULL,         -- 'order', 'friend_request', 'user' и т.д.
                               aggregate_id        UUID            NOT NULL,         -- ID агрегата (order_id, user_id и т.д.)

    -- Тип и данные события
                               event_type          VARCHAR(100)    NOT NULL,         -- 'social.friend.updated', 'chat.message.sent'
                               payload             JSONB           NOT NULL,         -- данные события (можно заменить на BYTEA для protobuf)

    -- Метаданные публикации
                               topic               VARCHAR(100)    NOT NULL,         -- Kafka топик назначения
                               partition_key       VARCHAR(255),                     -- ключ партиционирования (chat_id, user_id)

    -- Статус обработки
                               published_at        TIMESTAMPTZ,                      -- NULL = не опубликовано
                               retry_count         INT             NOT NULL DEFAULT 0,
                               next_attempt_at     TIMESTAMPTZ     NOT NULL DEFAULT now(),
                               last_error          TEXT,

    -- Временные метки
                               created_at          TIMESTAMPTZ     NOT NULL DEFAULT now(),
                               updated_at          TIMESTAMPTZ     NOT NULL DEFAULT now()
);

-- Индексы для эффективной работы воркера
CREATE INDEX idx_outbox_pending ON outbox_events(next_attempt_at, published_at)
    WHERE published_at IS NULL;  -- частичный индекс только для неопубликованных

CREATE INDEX idx_outbox_aggregate ON outbox_events(aggregate_type, aggregate_id, created_at DESC);

CREATE INDEX idx_outbox_cleanup ON outbox_events(published_at, created_at)
    WHERE published_at IS NOT NULL;  -- для очистки старых записей

-- Автообновление updated_at
CREATE OR REPLACE FUNCTION update_outbox_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER outbox_events_updated_at
    BEFORE UPDATE ON outbox_events
    FOR EACH ROW
    EXECUTE FUNCTION update_outbox_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS outbox_events_updated_at ON outbox_events;
DROP FUNCTION IF EXISTS update_outbox_updated_at();
DROP TABLE IF EXISTS outbox_events;
-- +goose StatementEnd