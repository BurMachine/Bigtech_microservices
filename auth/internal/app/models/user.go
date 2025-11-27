package models

import "time"

// ============================================
// RESPONSE MODELS
// ============================================

// User - пользователь (возвращается после регистрации)
type User struct {
	ID        string
	Email     string
	CreatedAt time.Time
}

// UserToken - токены пользователя (возвращается после Login/Refresh)
type UserToken struct {
	UserID       string
	AccessToken  string
	RefreshToken string
	ExpiresInS   int64 // время жизни access токена в секундах (например, 900 = 15 минут)
}

// ============================================
// JWKS MODELS
// ============================================

// JWK - публичный ключ в формате JSON Web Key
type JWK struct {
	KID string `json:"kid"` // Key ID
	Kty string `json:"kty"` // Key Type (обычно "RSA")
	Use string `json:"use"` // Use (обычно "sig" для подписи)
	Alg string `json:"alg"` // Algorithm (RS256, RS384, RS512)
	N   string `json:"n"`   // Modulus (base64url)
	E   string `json:"e"`   // Exponent (base64url)
}

// JWKSResponse - ответ с набором публичных ключей
type JWKSResponse struct {
	Keys []JWK `json:"keys"`
}
