-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE inbox_messages (
                                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                message_key VARCHAR(255) NOT NULL,  -- Ключ для дедупликации/идентификации сообщения
                                consumer VARCHAR(100) NOT NULL,     -- Имя потребителя (напр. 'social-notifier')
                                processed_at TIMESTAMP WITH TIME ZONE NULL,  -- NULL = не обработано
                                created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_inbox_messages_consumer ON inbox_messages(consumer);
CREATE INDEX idx_inbox_messages_processed ON inbox_messages(processed_at);
CREATE INDEX idx_inbox_messages_key ON inbox_messages(message_key);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_inbox_messages_key;
DROP INDEX IF EXISTS idx_inbox_messages_processed;
DROP INDEX IF EXISTS idx_inbox_messages_consumer;

DROP TABLE IF EXISTS inbox_messages;

DROP EXTENSION IF EXISTS "uuid-ossp";
-- +goose StatementEnd