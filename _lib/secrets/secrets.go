package secretslib

import (
	"context"
	"fmt"
	"os"
)

type SecretsProvider interface {
	Get(ctx context.Context, key string) (string, error)      // строки (пароли, токены)
	GetBytes(ctx context.Context, key string) ([]byte, error) // бинарь (TLS ключи/серты)
}

// GetSecretsMap загружает секреты по списку ключей для указанного окружения
func GetSecretsMap(ctx context.Context, env string, keys []string) (map[string]string, error) {
	// Создаем провайдер в зависимости от окружения
	provider, err := createProvider(env)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Загружаем все секреты по ключам
	secretsMap := make(map[string]string, len(keys))

	for _, key := range keys {
		value, err := provider.Get(ctx, key)
		if err != nil {
			// Если секрет не найден, возвращаем ошибку
			return nil, fmt.Errorf("failed to load secret '%s': %w", key, err)
		}
		secretsMap[key] = value
	}

	return secretsMap, nil
}

// GetSecretsMapOptional загружает секреты, пропуская отсутствующие
func GetSecretsMapOptional(ctx context.Context, env string, keys []string) (map[string]string, error) {
	provider, err := createProvider(env)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	secretsMap := make(map[string]string)

	for _, key := range keys {
		value, err := provider.Get(ctx, key)
		if err != nil {
			// Пропускаем отсутствующие секреты (опциональные)
			continue
		}
		secretsMap[key] = value
	}

	return secretsMap, nil
}

// createProvider создает нужный провайдер в зависимости от окружения
func createProvider(env string) (SecretsProvider, error) {
	switch env {
	case "local":
		// Local: только файл secrets.yaml для локальной разработки
		return NewFileProvider("secrets.yaml"), nil

	case "dev":
		// Dev: только ENV переменные (из Kubernetes Secrets, Docker, etc)
		return NewEnvProvider(""), nil

	case "prod":
		// Production: только Vault
		vaultAddr := os.Getenv("VAULT_ADDR")
		if vaultAddr == "" {
			return nil, fmt.Errorf("VAULT_ADDR is required for production")
		}

		vaultToken := os.Getenv("VAULT_TOKEN")
		if vaultToken == "" {
			return nil, fmt.Errorf("VAULT_TOKEN is required for production")
		}

		return NewVaultProvider(&VaultConfig{
			Address:    vaultAddr,
			Token:      vaultToken,
			MountPath:  "secret",
			SecretPath: "myapp/production",
		})

	default:
		return nil, fmt.Errorf("unknown environment: %s (supported: local, dev, prod)", env)
	}
}

// MustGetSecretsMap загружает секреты или паникует
func MustGetSecretsMap(ctx context.Context, env string, keys []string) map[string]string {
	secretsMap, err := GetSecretsMap(ctx, env, keys)
	if err != nil {
		panic(fmt.Sprintf("failed to load secrets: %v", err))
	}
	return secretsMap
}

// ValidateSecrets проверяет наличие всех обязательных секретов
func ValidateSecrets(secretsMap map[string]string, requiredKeys []string) error {
	missing := make([]string, 0)

	for _, key := range requiredKeys {
		if value, exists := secretsMap[key]; !exists || value == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required secrets: %v", missing)
	}

	return nil
}
