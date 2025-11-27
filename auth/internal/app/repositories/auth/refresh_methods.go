package auth_repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/models"
	"github.com/Burmachine/MSA/lib/postgreslib"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

// ============================================
// REFRESH TOKEN METHODS
// ============================================

// CreateRefreshToken создает новый refresh токен
func (r *RepositoryImpl) CreateRefreshToken(ctx context.Context, userID, token string, deviceID *string, expiresAt time.Time) error {
	const api = "auth_repo.Repository.CreateRefreshToken"

	if _, err := uuid.Parse(userID); err != nil {
		return fmt.Errorf("%s: invalid user_id format: %w", api, err)
	}

	// Хешируем токен
	tokenHash := hashToken(token)
	tokenID := uuid.New().String()

	qb := r.qb.Insert(tableRefreshTokens).
		Columns(colRefreshTokenID, colRefreshTokenUserID, colRefreshTokenHash, colRefreshTokenDeviceID, colRefreshTokenExpires, colRefreshTokenCreated).
		Values(tokenID, userID, tokenHash, deviceID, expiresAt, time.Now())

	conn := r.db.GetQueryEngine(ctx)
	if _, err := conn.Execx(ctx, qb); err != nil {
		return fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	return nil
}

// GetRefreshToken получает refresh токен по хешу
func (r *RepositoryImpl) GetRefreshToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	const api = "auth_repo.Repository.GetRefreshToken"

	tokenHash := hashToken(token)

	var rt models.RefreshToken
	qb := r.qb.Select(
		colRefreshTokenID, colRefreshTokenUserID, colRefreshTokenHash, colRefreshTokenDeviceID,
		colRefreshTokenExpires, colRefreshTokenCreated, colRefreshTokenUsedAt, colRefreshTokenRevoked,
	).From(tableRefreshTokens).
		Where(squirrel.Eq{colRefreshTokenHash: tokenHash})

	conn := r.db.GetQueryEngine(ctx)
	if err := conn.Getx(ctx, &rt, qb); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: token not found", api)
		}
		return nil, fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	return &rt, nil
}

// MarkTokenAsUsed помечает токен как использованный
func (r *RepositoryImpl) MarkTokenAsUsed(ctx context.Context, tokenID string) error {
	const api = "auth_repo.Repository.MarkTokenAsUsed"

	if _, err := uuid.Parse(tokenID); err != nil {
		return fmt.Errorf("%s: invalid token_id format: %w", api, err)
	}

	qb := r.qb.Update(tableRefreshTokens).
		Set(colRefreshTokenUsedAt, time.Now()).
		Where(squirrel.Eq{colRefreshTokenID: tokenID})

	conn := r.db.GetQueryEngine(ctx)
	if _, err := conn.Execx(ctx, qb); err != nil {
		return fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	return nil
}

// RevokeToken отзывает токен
func (r *RepositoryImpl) RevokeToken(ctx context.Context, token string) error {
	const api = "auth_repo.Repository.RevokeToken"

	tokenHash := hashToken(token)

	qb := r.qb.Update(tableRefreshTokens).
		Set(colRefreshTokenRevoked, time.Now()).
		Where(squirrel.Eq{colRefreshTokenHash: tokenHash}).
		Where(squirrel.Eq{colRefreshTokenRevoked: nil})

	conn := r.db.GetQueryEngine(ctx)
	result, err := conn.Execx(ctx, qb)
	if err != nil {
		return fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	rowsAffected := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", api, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: token not found or already revoked", api)
	}

	return nil
}

// DeleteAllUserTokens удаляет все токены пользователя (при обнаружении reuse)
func (r *RepositoryImpl) DeleteAllUserTokens(ctx context.Context, userID string) error {
	const api = "auth_repo.Repository.DeleteAllUserTokens"

	if _, err := uuid.Parse(userID); err != nil {
		return fmt.Errorf("%s: invalid user_id format: %w", api, err)
	}

	qb := r.qb.Delete(tableRefreshTokens).
		Where(squirrel.Eq{colRefreshTokenUserID: userID})

	conn := r.db.GetQueryEngine(ctx)
	if _, err := conn.Execx(ctx, qb); err != nil {
		return fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	return nil
}
