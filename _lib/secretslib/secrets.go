package secretslib

import (
	"context"
	"fmt"
	"os"
	"reflect"
)

type SecretsProvider interface {
	Get(ctx context.Context, key string) (string, error)      // строки (пароли, токены)
	GetBytes(ctx context.Context, key string) ([]byte, error) // бинарь (TLS ключи/серты)
}

// LoadSecrets загружает секреты в структуру по env тегам
func LoadSecrets[T any](ctx context.Context, env string) (*T, error) {
	secrets := new(T)

	// Создаем провайдер
	provider, err := createProvider(env)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Заполняем поля
	val := reflect.ValueOf(secrets).Elem()
	t := val.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := val.Field(i)

		if !field.IsExported() {
			continue
		}

		// Получаем имя ключа из тега env
		envName := field.Tag.Get("env")
		if envName == "" {
			continue
		}

		// Читаем секрет
		value, err := provider.Get(ctx, envName)
		if err != nil {
			continue // Пропускаем если не найден
		}

		// Устанавливаем значение
		if fieldVal.CanSet() && fieldVal.Kind() == reflect.String {
			fieldVal.SetString(value)
		}
	}

	return secrets, nil
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
