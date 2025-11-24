package secretslib

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	vault "github.com/hashicorp/vault/api"
)

// VaultProvider читает секреты из HashiCorp Vault
type VaultProvider struct {
	client     *vault.Client
	mountPath  string
	secretPath string
}

type VaultConfig struct {
	Address    string
	Token      string
	MountPath  string
	SecretPath string
}

// NewVaultProvider создает провайдер для работы с Vault
func NewVaultProvider(cfg *VaultConfig) (*VaultProvider, error) {
	if cfg == nil {
		return nil, ErrProviderNotConfigured
	}

	if cfg.Address == "" {
		return nil, fmt.Errorf("vault address is required")
	}

	if cfg.Token == "" {
		return nil, fmt.Errorf("vault token is required")
	}

	// Создаем конфигурацию клиента
	config := vault.DefaultConfig()
	config.Address = cfg.Address
	config.Timeout = 30 * time.Second

	// Создаем клиент Vault
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	// Устанавливаем токен для авторизации
	client.SetToken(cfg.Token)

	// Проверяем подключение к Vault
	if err := checkVaultConnection(client); err != nil {
		return nil, fmt.Errorf("failed to connect to vault: %w", err)
	}

	// Устанавливаем дефолтные значения
	mountPath := cfg.MountPath
	if mountPath == "" {
		mountPath = "secret"
	}

	secretPath := cfg.SecretPath
	if secretPath == "" {
		secretPath = "app"
	}

	return &VaultProvider{
		client:     client,
		mountPath:  mountPath,
		secretPath: secretPath,
	}, nil
}

// checkVaultConnection проверяет подключение к Vault
func checkVaultConnection(client *vault.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Проверяем статус Vault
	health, err := client.Sys().HealthWithContext(ctx)
	if err != nil {
		return fmt.Errorf("vault is unreachable: %w", err)
	}

	if health.Sealed {
		return fmt.Errorf("vault is sealed")
	}

	return nil
}

// Get возвращает секрет как строку из Vault
func (p *VaultProvider) Get(ctx context.Context, key string) (string, error) {
	// Формируем путь для KV v2: mountPath/data/secretPath
	path := fmt.Sprintf("%s/data/%s", p.mountPath, p.secretPath)

	// Читаем секрет из Vault
	secret, err := p.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return "", fmt.Errorf("failed to read secret from vault: %w", err)
	}

	if secret == nil || secret.Data == nil {
		return "", fmt.Errorf("secret not found at path: %s", path)
	}

	// Vault KV v2 хранит данные в поле "data"
	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid vault response format at path: %s", path)
	}

	// Получаем значение по ключу
	value, exists := data[key]
	if !exists {
		return "", ErrSecretNotFound
	}

	// Конвертируем в строку
	return convertToString(value, key)
}

// GetBytes возвращает секрет как байты из Vault
func (p *VaultProvider) GetBytes(ctx context.Context, key string) ([]byte, error) {
	value, err := p.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	// Пробуем декодировать из base64 (для бинарных данных)
	if decoded, err := base64.StdEncoding.DecodeString(value); err == nil {
		// Проверяем что это действительно base64
		if base64.StdEncoding.EncodeToString(decoded) == value {
			return decoded, nil
		}
	}

	// Если не base64, возвращаем как есть
	return []byte(value), nil
}

// Name возвращает имя провайдера
func (p *VaultProvider) Name() string {
	return fmt.Sprintf("vault(%s/%s)", p.mountPath, p.secretPath)
}

// GetAll возвращает все секреты по указанному пути
func (p *VaultProvider) GetAll(ctx context.Context) (map[string]string, error) {
	path := fmt.Sprintf("%s/data/%s", p.mountPath, p.secretPath)

	secret, err := p.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets from vault: %w", err)
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no secrets found at path: %s", path)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid vault response format")
	}

	result := make(map[string]string, len(data))
	for key, value := range data {
		str, err := convertToString(value, key)
		if err != nil {
			// Пропускаем ключи, которые не можем конвертировать
			continue
		}
		result[key] = str
	}

	return result, nil
}

// Close закрывает соединение с Vault
func (p *VaultProvider) Close() error {
	// Vault client не требует явного закрытия
	// Но можно очистить токен для безопасности
	p.client.ClearToken()
	return nil
}

// RenewToken обновляет токен авторизации в Vault
func (p *VaultProvider) RenewToken(ctx context.Context) error {
	secret, err := p.client.Auth().Token().RenewSelfWithContext(ctx, 0)
	if err != nil {
		return fmt.Errorf("failed to renew vault token: %w", err)
	}

	if secret == nil {
		return fmt.Errorf("received nil secret when renewing token")
	}

	return nil
}

// convertToString конвертирует значение в строку
func convertToString(value interface{}, key string) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case int:
		return fmt.Sprintf("%d", v), nil
	case int64:
		return fmt.Sprintf("%d", v), nil
	case float64:
		return fmt.Sprintf("%g", v), nil
	case bool:
		return fmt.Sprintf("%t", v), nil
	default:
		return "", fmt.Errorf("unsupported secret type for key '%s': %T", key, v)
	}
}
