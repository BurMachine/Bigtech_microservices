package keygen

import (
	"context"
	"fmt"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth"
	loggerlib "github.com/Burmachine/MSA/lib/logger"
)

type KeyInitializer struct {
	repo   auth.AuthRepository
	logger *loggerlib.Logger
}

func NewKeyInitializer(repo auth.AuthRepository, logger *loggerlib.Logger) *KeyInitializer {
	return &KeyInitializer{
		repo:   repo,
		logger: logger,
	}
}

// InitializeKeys проверяет и создаёт RSA ключ если его нет
func (k *KeyInitializer) InitializeKeys(ctx context.Context) error {
	k.logger.Info(ctx, "checking for active RSA key")

	// 1. Проверяем есть ли активный ключ
	existingKey, err := k.repo.GetActiveRSAKey(ctx)
	if err == nil {
		// Ключ уже есть
		k.logger.Info(ctx, "active RSA key found", "kid", existingKey.KID)
		return nil
	}

	// 2. Генерируем новый ключ
	k.logger.Warn(ctx, "no active RSA key found, generating new one")

	kid := GenerateKID()
	privateKeyPEM, publicKeyPEM, err := GenerateRSAKeyPair(2048)
	if err != nil {
		k.logger.Error(ctx, "failed to generate RSA key pair", "error", err)
		return fmt.Errorf("failed to generate RSA key pair: %w", err)
	}

	// 3. Сохраняем в базу
	err = k.repo.CreateRSAKey(ctx, kid, privateKeyPEM, publicKeyPEM, "RS256", "active")
	if err != nil {
		k.logger.Error(ctx, "failed to save RSA key to database", "kid", kid, "error", err)
		return fmt.Errorf("failed to save RSA key: %w", err)
	}

	k.logger.Info(ctx, "RSA key created successfully", "kid", kid, "algorithm", "RS256", "bits", 2048)
	return nil
}

// RotateKeys выполняет ротацию ключей (для будущего использования)
func (k *KeyInitializer) RotateKeys(ctx context.Context) error {
	k.logger.Info(ctx, "starting key rotation")

	// 1. Получаем текущий активный ключ
	currentKey, err := k.repo.GetActiveRSAKey(ctx)
	if err != nil {
		k.logger.Error(ctx, "failed to get current active key", "error", err)
		return fmt.Errorf("failed to get current active key: %w", err)
	}

	k.logger.Debug(ctx, "current active key retrieved", "kid", currentKey.KID)

	// 2. Генерируем новый ключ
	kid := GenerateKID()
	privateKeyPEM, publicKeyPEM, err := GenerateRSAKeyPair(2048)
	if err != nil {
		k.logger.Error(ctx, "failed to generate new RSA key pair", "error", err)
		return fmt.Errorf("failed to generate RSA key pair: %w", err)
	}

	// 3. Сохраняем новый как active
	err = k.repo.CreateRSAKey(ctx, kid, privateKeyPEM, publicKeyPEM, "RS256", "active")
	if err != nil {
		k.logger.Error(ctx, "failed to save new RSA key", "kid", kid, "error", err)
		return fmt.Errorf("failed to save RSA key: %w", err)
	}

	k.logger.Info(ctx, "key rotation completed successfully",
		"old_kid", currentKey.KID,
		"new_kid", kid,
	)

	return nil
}
