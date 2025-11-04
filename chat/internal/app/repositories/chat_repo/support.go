package chat_repo

import "github.com/jackc/pgx/v5/pgconn"

// IsUniqueViolation проверяет, является ли ошибка нарушением уникальности (код 23505)
func IsUniqueViolation(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr.Code == "23505"
	}
	return false
}
