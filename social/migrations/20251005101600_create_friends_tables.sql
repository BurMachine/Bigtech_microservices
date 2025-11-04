-- +goose Up
-- +goose StatementBegin

-- Таблица для заявок в друзья
CREATE TABLE IF NOT EXISTS friend_requests (
                                               id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                               from_user_id UUID NOT NULL,
                                               to_user_id UUID NOT NULL,
                                               status VARCHAR(20) NOT NULL CHECK (status IN ('PENDING', 'ACCEPTED', 'DECLINED')),
                                               created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                                               updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
                                               CONSTRAINT unique_request UNIQUE (from_user_id, to_user_id)
);

-- Таблица для дружеских связей
CREATE TABLE IF NOT EXISTS friends (
                                       user_id UUID NOT NULL,
                                       friend_user_id UUID NOT NULL,
                                       created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                                       CONSTRAINT unique_friendship UNIQUE (user_id, friend_user_id),
                                       CONSTRAINT check_different_users CHECK (user_id != friend_user_id)
);

-- Индексы для оптимизации запросов
CREATE INDEX idx_friend_requests_to_user_id ON friend_requests(to_user_id, status, created_at DESC);
CREATE INDEX idx_friends_user_id ON friends(user_id, created_at DESC);

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггер для автоматического обновления updated_at в friend_requests
CREATE TRIGGER update_friend_requests_updated_at BEFORE UPDATE ON friend_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_friend_requests_updated_at ON friend_requests;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP INDEX IF EXISTS idx_friend_requests_to_user_id;
DROP INDEX IF EXISTS idx_friends_user_id;
DROP TABLE IF EXISTS friends;
DROP TABLE IF EXISTS friend_requests;
-- +goose StatementEnd