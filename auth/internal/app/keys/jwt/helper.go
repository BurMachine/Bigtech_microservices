// internal/app/jwt/jwt.go

package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// Claims представляет JWT claims для access токена
type Claims struct {
	jwt.RegisteredClaims
}

// GenerateAccessToken создаёт JWT access токен
func GenerateAccessToken(
	userID string,
	kid string,
	privateKeyPEM string,
	issuer string,
	audience []string,
	expiresIn time.Duration,
) (string, error) {
	// 1. Парсим приватный ключ из PEM
	privateKey, err := parseRSAPrivateKey(privateKeyPEM)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// 2. Создаём claims
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,                                 // "https://auth.example.com"
			Subject:   userID,                                 // user ID
			Audience:  audience,                               // ["chat-service", "user-service"]
			ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)), // exp
			IssuedAt:  jwt.NewNumericDate(now),                // iat
			NotBefore: jwt.NewNumericDate(now),                // nbf
			ID:        uuid.New().String(),                    // jti (unique token ID)
		},
	}

	// 3. Создаём токен с RS256 алгоритмом
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// 4. Добавляем kid в header
	token.Header["kid"] = kid

	// 5. Подписываем токен приватным ключом
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// parseRSAPrivateKey парсит RSA приватный ключ из PEM формата
func parseRSAPrivateKey(privateKeyPEM string) (*rsa.PrivateKey, error) {
	// Декодируем PEM блок
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Парсим приватный ключ
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return privateKey, nil
}

// ValidateAccessToken валидирует JWT токен (для других сервисов)
func ValidateAccessToken(tokenString string, publicKeyPEM string) (*Claims, error) {
	// Парсим публичный ключ
	publicKey, err := parseRSAPublicKey(publicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	// Парсим и валидируем токен
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Извлекаем claims
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// parseRSAPublicKey парсит RSA публичный ключ из PEM формата
func parseRSAPublicKey(publicKeyPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPubKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPubKey, nil
}
