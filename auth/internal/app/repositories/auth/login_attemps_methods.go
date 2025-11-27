package auth_repo

import (
	"context"
	"fmt"
	"time"

	"github.com/Burmachine/MSA/lib/postgreslib"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

// ============================================
// LOGIN ATTEMPTS METHODS
// ============================================

// RecordLoginAttempt записывает попытку входа
func (r *RepositoryImpl) RecordLoginAttempt(ctx context.Context, email, ipAddress string, success bool) error {
	const api = "auth_repo.Repository.RecordLoginAttempt"

	attemptID := uuid.New().String()

	qb := r.qb.Insert(tableLoginAttempts).
		Columns(colLoginAttemptID, colLoginAttemptEmail, colLoginAttemptIP, colLoginAttemptSuccess, colLoginAttemptTimestamp).
		Values(attemptID, email, ipAddress, success, time.Now())

	conn := r.db.GetQueryEngine(ctx)
	if _, err := conn.Execx(ctx, qb); err != nil {
		return fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	return nil
}

// GetFailedLoginAttempts получает количество неудачных попыток входа за период
func (r *RepositoryImpl) GetFailedLoginAttempts(ctx context.Context, email string, since time.Time) (int, error) {
	const api = "auth_repo.Repository.GetFailedLoginAttempts"

	var count int
	qb := r.qb.Select("COUNT(*)").
		From(tableLoginAttempts).
		Where(squirrel.Eq{colLoginAttemptEmail: email}).
		Where(squirrel.Eq{colLoginAttemptSuccess: false}).
		Where(squirrel.GtOrEq{colLoginAttemptTimestamp: since})

	conn := r.db.GetQueryEngine(ctx)
	if err := conn.Getx(ctx, &count, qb); err != nil {
		return 0, fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	return count, nil
}
