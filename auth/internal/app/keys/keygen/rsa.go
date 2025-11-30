// internal/app/keygen/rsa.go

package keygen

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"
)

// GenerateRSAKeyPair генерирует пару RSA ключей
func GenerateRSAKeyPair(bits int) (privateKeyPEM, publicKeyPEM string, err error) {
	// Генерация приватного ключа
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}

	// Кодирование приватного ключа в PEM
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}))

	// Кодирование публичного ключа в PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal public key: %w", err)
	}
	publicKeyPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}))

	return privateKeyPEM, publicKeyPEM, nil
}

// GenerateKID генерирует Key ID на основе текущей даты
func GenerateKID() string {
	return fmt.Sprintf("key-%s", time.Now().Format("2006-01-02-150405"))
}
