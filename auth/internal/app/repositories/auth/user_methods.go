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
// USER METHODS
// ============================================

// CreateUserWithHash создает пользователя с уже захешированным паролем
func (r *RepositoryImpl) CreateUserWithHash(ctx context.Context, email, passwordHash string) (*models.UserRepo, error) {
	const api = "auth_repo.Repository.CreateUserWithHash"

	userID := uuid.New().String()
	now := time.Now()

	qb := r.qb.Insert(tableUsers).
		Columns(colUserID, colUserEmail, colUserPasswordHash, colUserCreatedAt).
		Values(userID, email, passwordHash, now).
		Suffix("RETURNING id, email, created_at, is_active")

	conn := r.db.GetQueryEngine(ctx)

	var user models.UserRepo
	if err := conn.Getx(ctx, &user, qb); err != nil {
		return nil, fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	return &user, nil
}

// GetUserByEmail получает пользователя по email
func (r *RepositoryImpl) GetUserByEmail(ctx context.Context, email string) (*models.UserRepo, error) {
	const api = "auth_repo.Repository.GetUserByEmail"

	var user models.UserRepo
	qb := r.qb.Select(
		colUserID, colUserEmail, colUserPasswordHash, colUserCreatedAt, colUserIsActive,
	).From(tableUsers).
		Where(squirrel.Eq{colUserEmail: email})

	conn := r.db.GetQueryEngine(ctx)
	if err := conn.Getx(ctx, &user, qb); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: user not found", api)
		}
		return nil, fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	return &user, nil
}

// GetUserByID получает пользователя по ID
func (r *RepositoryImpl) GetUserByID(ctx context.Context, userID string) (*models.UserRepo, error) {
	const api = "auth_repo.Repository.GetUserByID"

	if _, err := uuid.Parse(userID); err != nil {
		return nil, fmt.Errorf("%s: invalid user_id format: %w", api, err)
	}

	var user models.UserRepo
	qb := r.qb.Select(
		colUserID, colUserEmail, colUserPasswordHash, colUserCreatedAt, colUserIsActive,
	).From(tableUsers).
		Where(squirrel.Eq{colUserID: userID})

	conn := r.db.GetQueryEngine(ctx)
	if err := conn.Getx(ctx, &user, qb); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: user not found", api)
		}
		return nil, fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	return &user, nil
}

// CheckUserExistsByEmail проверяет существование пользователя по email
func (r *RepositoryImpl) CheckUserExistsByEmail(ctx context.Context, email string) (bool, error) {
	const api = "auth_repo.Repository.CheckUserExistsByEmail"

	var count int
	qb := r.qb.Select("COUNT(*)").
		From(tableUsers).
		Where(squirrel.Eq{colUserEmail: email})

	conn := r.db.GetQueryEngine(ctx)
	if err := conn.Getx(ctx, &count, qb); err != nil {
		return false, fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	return count > 0, nil
}

// DeleteUser удаляет пользователя (для компенсации транзакций в saga pattern)
func (r *RepositoryImpl) DeleteUser(ctx context.Context, userID string) error {
	const api = "auth_repo.Repository.DeleteUser"

	if _, err := uuid.Parse(userID); err != nil {
		return fmt.Errorf("%s: invalid user_id format: %w", api, err)
	}

	qb := r.qb.Delete(tableUsers).
		Where(squirrel.Eq{colUserID: userID})

	conn := r.db.GetQueryEngine(ctx)
	if _, err := conn.Execx(ctx, qb); err != nil {
		return fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	return nil
}
