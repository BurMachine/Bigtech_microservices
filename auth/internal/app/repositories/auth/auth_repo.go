package auth_repo

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/Burmachine/MSA/lib/postgreslib"
	"github.com/Masterminds/squirrel"
)

type RepositoryImpl struct {
	db postgreslib.QueryEngineProvider
	qb squirrel.StatementBuilderType
}

// NewRepository конструктор Repository
func NewRepository(p postgreslib.QueryEngineProvider) *RepositoryImpl {
	return &RepositoryImpl{
		db: p,
		qb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Константы для названий таблиц и колонок
const (
	// refresh_tokens
	tableRefreshTokens      = "refresh_tokens"
	colRefreshTokenID       = "id"
	colRefreshTokenUserID   = "user_id"
	colRefreshTokenHash     = "token_hash"
	colRefreshTokenDeviceID = "device_id"
	colRefreshTokenExpires  = "expires_at"
	colRefreshTokenCreated  = "created_at"
	colRefreshTokenUsedAt   = "used_at"
	colRefreshTokenRevoked  = "revoked_at"

	// rsa_keys
	tableRSAKeys         = "rsa_keys"
	colRSAKeyID          = "id"
	colRSAKeyKID         = "kid"
	colRSAKeyPrivate     = "private_key"
	colRSAKeyPublic      = "public_key"
	colRSAKeyAlgorithm   = "algorithm"
	colRSAKeyStatus      = "status"
	colRSAKeyCreatedAt   = "created_at"
	colRSAKeyActivatedAt = "activated_at"
	colRSAKeyRetiredAt   = "retired_at"

	// login_attempts
	tableLoginAttempts       = "login_attempts"
	colLoginAttemptID        = "id"
	colLoginAttemptEmail     = "email"
	colLoginAttemptIP        = "ip_address"
	colLoginAttemptSuccess   = "success"
	colLoginAttemptTimestamp = "attempted_at"

	tableUsers          = "users"
	colUserID           = "id"
	colUserEmail        = "email"
	colUserPasswordHash = "password_hash"
	colUserCreatedAt    = "created_at"
	colUserIsActive     = "is_active"
)

// ============================================
// HELPER FUNCTIONS
// ============================================

// hashToken создает SHA-256 хеш токена
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
