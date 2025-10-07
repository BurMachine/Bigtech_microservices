-- +goose Up
-- +goose StatementBegin

-- Расширение для полнотекстового поиска должно быть установлено ПЕРВЫМ
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS user_profiles (
                                             id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    nickname VARCHAR(20) UNIQUE NOT NULL,
    bio TEXT,
    avatar_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Ограничения
    CONSTRAINT check_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT check_nickname_format CHECK (nickname ~* '^[a-z0-9_]{3,20}$')
    );

-- Индексы для быстрого поиска
CREATE INDEX idx_user_profiles_email ON user_profiles(email);
CREATE INDEX idx_user_profiles_nickname ON user_profiles(nickname);
CREATE INDEX idx_user_profiles_nickname_trgm ON user_profiles USING gin(nickname gin_trgm_ops);

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггер для автоматического обновления updated_at
CREATE TRIGGER update_user_profiles_updated_at BEFORE UPDATE ON user_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_user_profiles_updated_at ON user_profiles;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP INDEX IF EXISTS idx_user_profiles_nickname_trgm;
DROP INDEX IF EXISTS idx_user_profiles_nickname;
DROP INDEX IF EXISTS idx_user_profiles_email;
DROP TABLE IF EXISTS user_profiles;
DROP EXTENSION IF EXISTS pg_trgm;
-- +goose StatementEnd