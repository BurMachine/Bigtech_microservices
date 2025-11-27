package models

import "time"

// User представляет пользователя для аутентификации
type UserRepo struct {
	ID           string    `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	IsActive     bool      `db:"is_active"`
}

// RefreshToken представляет refresh токен
type RefreshToken struct {
	ID        string     `db:"id"`
	UserID    string     `db:"user_id"`
	TokenHash string     `db:"token_hash"`
	DeviceID  *string    `db:"device_id"`
	ExpiresAt time.Time  `db:"expires_at"`
	CreatedAt time.Time  `db:"created_at"`
	UsedAt    *time.Time `db:"used_at"`
	RevokedAt *time.Time `db:"revoked_at"`
}

// RSAKey представляет RSA ключ для подписи JWT
type RSAKey struct {
	ID          string     `db:"id"`
	KID         string     `db:"kid"`
	PrivateKey  string     `db:"private_key"`
	PublicKey   string     `db:"public_key"`
	Algorithm   string     `db:"algorithm"`
	Status      string     `db:"status"` // active, next, retired
	CreatedAt   time.Time  `db:"created_at"`
	ActivatedAt *time.Time `db:"activated_at"`
	RetiredAt   *time.Time `db:"retired_at"`
}

// LoginAttempt представляет попытку входа
type LoginAttempt struct {
	ID          string    `db:"id"`
	Email       string    `db:"email"`
	IPAddress   string    `db:"ip_address"`
	Success     bool      `db:"success"`
	AttemptedAt time.Time `db:"attempted_at"`
}
