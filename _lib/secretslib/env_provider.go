package secretslib

import (
	"context"
	"os"
	"strings"
)

type EnvProvider struct {
	prefix string
}

func NewEnvProvider(prefix string) *EnvProvider {
	return &EnvProvider{
		prefix: prefix,
	}
}

func (p *EnvProvider) Get(ctx context.Context, key string) (string, error) {
	// Пробуем с префиксом
	if p.prefix != "" {
		envKey := p.prefix + key
		if value := os.Getenv(envKey); value != "" {
			return value, nil
		}
	}

	// Пробуем без префикса
	if value := os.Getenv(key); value != "" {
		return value, nil
	}

	return "", ErrSecretNotFound
}

func (p *EnvProvider) GetBytes(ctx context.Context, key string) ([]byte, error) {
	value, err := p.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	return []byte(value), nil
}

// Name возвращает имя провайдера
func (p *EnvProvider) Name() string {
	if p.prefix != "" {
		return "env(prefix=" + strings.TrimSuffix(p.prefix, "_") + ")"
	}
	return "env"
}
