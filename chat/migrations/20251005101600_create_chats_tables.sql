-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE chats (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE chat_members (
                              chat_id UUID NOT NULL,
                              user_id UUID NOT NULL,  -- UUID от user-сервиса, без FK
                              UNIQUE(chat_id, user_id)
);

CREATE TABLE messages (
                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          chat_id UUID NOT NULL,
                          sender_id UUID NOT NULL,  -- UUID от user-сервиса, без FK
                          text TEXT NOT NULL,
                          created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_messages_chat_created ON messages(chat_id, created_at);
CREATE INDEX idx_chat_members_chat ON chat_members(chat_id);
CREATE INDEX idx_chat_members_user ON chat_members(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_chat_members_user;
DROP INDEX IF EXISTS idx_chat_members_chat;
DROP INDEX IF EXISTS idx_messages_chat_created;

DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS chat_members;
DROP TABLE IF EXISTS chats;

DROP EXTENSION IF EXISTS "uuid-ossp";
-- +goose StatementEnd