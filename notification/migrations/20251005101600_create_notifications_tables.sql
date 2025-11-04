-- +goose Up
-- +goose StatementBegin
CREATE TABLE inbox_messages (
                                id            UUID         PRIMARY KEY,
                                topic         TEXT         NOT NULL,
                                partition     INT          NOT NULL,
                                kafka_offset  BIGINT       NOT NULL,      -- Переименовано из offset
                                payload       JSONB        NOT NULL,
                                status        TEXT         NOT NULL,
                                attempts      INT          NOT NULL DEFAULT 0,
                                last_error    TEXT,
                                received_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
                                processed_at  TIMESTAMPTZ,

                                CONSTRAINT valid_status CHECK (status IN ('received', 'processing', 'processed', 'failed'))
);

-- Индекс для быстрого поиска сообщений по статусу и количеству попыток
CREATE INDEX idx_inbox_messages_status_attempts
    ON inbox_messages(status, attempts)
    WHERE status IN ('received', 'failed');

-- Индекс для дедупликации и быстрого поиска по топику
CREATE INDEX idx_inbox_messages_topic_partition_offset
    ON inbox_messages(topic, partition, kafka_offset);

-- Индекс для очистки старых обработанных сообщений
CREATE INDEX idx_inbox_messages_processed_at
    ON inbox_messages(processed_at)
    WHERE processed_at IS NOT NULL;

-- Индекс для временных меток (сортировка и поиск)
CREATE INDEX idx_inbox_messages_received_at
    ON inbox_messages(received_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS inbox_messages;
-- +goose StatementEnd