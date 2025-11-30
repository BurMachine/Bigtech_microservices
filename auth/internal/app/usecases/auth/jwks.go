package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/models"
)

// JWKS возвращает публичные ключи для валидации JWT
func (a *AuthService) JWKS(ctx context.Context) (*models.JWKSResponse, error) {
	const api = "AuthService.JWKS"

	// 1. Получение активных публичных ключей из БД
	rsaKeys, err := a.authRepo.GetPublicKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get public keys: %w", api, err)
	}

	// 2. Конвертация в JWK формат
	jwks := make([]models.JWK, 0, len(rsaKeys))
	for _, key := range rsaKeys {
		jwk, err := convertRSAKeyToJWK(key)
		if err != nil {
			// Пропускаем невалидные ключи, логируем ошибку
			fmt.Printf("failed to convert key %s to JWK: %v\n", key.KID, err)
			continue
		}
		jwks = append(jwks, jwk)
	}

	// 3. Если нет активных ключей - это критическая ошибка
	if len(jwks) == 0 {
		return nil, fmt.Errorf("%s: no valid public keys available", api)
	}

	return &models.JWKSResponse{
		Keys: jwks,
	}, nil
}

// convertRSAKeyToJWK конвертирует RSA публичный ключ в JWK формат
func convertRSAKeyToJWK(key *models.RSAKey) (models.JWK, error) {
	// 1. Парсим PEM блок
	block, _ := pem.Decode([]byte(key.PublicKey))
	if block == nil {
		return models.JWK{}, fmt.Errorf("failed to decode PEM block")
	}

	// 2. Парсим публичный ключ
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return models.JWK{}, fmt.Errorf("failed to parse public key: %w", err)
	}

	// 3. Приводим к типу *rsa.PublicKey
	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return models.JWK{}, fmt.Errorf("not an RSA public key")
	}

	// 4. Извлекаем modulus (N) и exponent (E)
	n := rsaPubKey.N.Bytes()
	e := big.NewInt(int64(rsaPubKey.E)).Bytes()

	// 5. Кодируем в base64url (без padding)
	nBase64 := base64.RawURLEncoding.EncodeToString(n)
	eBase64 := base64.RawURLEncoding.EncodeToString(e)

	// 6. Формируем JWK
	return models.JWK{
		KID: key.KID,
		Kty: "RSA",
		Use: "sig",
		Alg: key.Algorithm, // RS256, RS384, RS512
		N:   nBase64,
		E:   eBase64,
	}, nil
}
