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
// RSA KEY METHODS
// ============================================

// CreateRSAKey создает новый RSA ключ
func (r *RepositoryImpl) CreateRSAKey(ctx context.Context, kid, privateKey, publicKey, algorithm, status string) error {
	const api = "auth_repo.Repository.CreateRSAKey"

	keyID := uuid.New().String()
	now := time.Now()

	qb := r.qb.Insert(tableRSAKeys).
		Columns(colRSAKeyID, colRSAKeyKID, colRSAKeyPrivate, colRSAKeyPublic, colRSAKeyAlgorithm, colRSAKeyStatus, colRSAKeyCreatedAt).
		Values(keyID, kid, privateKey, publicKey, algorithm, status, now)

	conn := r.db.GetQueryEngine(ctx)
	if _, err := conn.Execx(ctx, qb); err != nil {
		return fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	return nil
}

// GetActiveRSAKey получает активный RSA ключ
func (r *RepositoryImpl) GetActiveRSAKey(ctx context.Context) (*models.RSAKey, error) {
	const api = "auth_repo.Repository.GetActiveRSAKey"

	var key models.RSAKey
	qb := r.qb.Select(
		colRSAKeyID, colRSAKeyKID, colRSAKeyPrivate, colRSAKeyPublic,
		colRSAKeyAlgorithm, colRSAKeyStatus, colRSAKeyCreatedAt, colRSAKeyActivatedAt, colRSAKeyRetiredAt,
	).From(tableRSAKeys).
		Where(squirrel.Eq{colRSAKeyStatus: "active"}).
		Limit(1)

	conn := r.db.GetQueryEngine(ctx)
	if err := conn.Getx(ctx, &key, qb); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: no active key found", api)
		}
		return nil, fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	return &key, nil
}

// GetPublicKeys получает все публичные ключи для JWKS
func (r *RepositoryImpl) GetPublicKeys(ctx context.Context) ([]*models.RSAKey, error) {
	const api = "auth_repo.Repository.GetPublicKeys"

	var keys []*models.RSAKey
	qb := r.qb.Select(
		colRSAKeyID, colRSAKeyKID, colRSAKeyPublic, colRSAKeyAlgorithm,
		colRSAKeyStatus, colRSAKeyCreatedAt, colRSAKeyActivatedAt, colRSAKeyRetiredAt,
	).From(tableRSAKeys).
		Where(squirrel.Or{
			squirrel.Eq{colRSAKeyStatus: "active"},
			squirrel.Eq{colRSAKeyStatus: "next"},
		})

	conn := r.db.GetQueryEngine(ctx)
	if err := conn.Selectx(ctx, &keys, qb); err != nil {
		return nil, fmt.Errorf("%s: %w", api, postgreslib.ConvertPGError(err))
	}

	return keys, nil
}
