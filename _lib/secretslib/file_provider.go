package secretslib

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// FileProvider читает секреты из YAML файла
type FileProvider struct {
	filePath string
	secrets  map[string]interface{}
	mu       sync.RWMutex
	loaded   bool
}

// NewFileProvider создает провайдер для чтения из файла
// filePath - путь к YAML файлу с секретами
func NewFileProvider(filePath string) *FileProvider {
	return &FileProvider{
		filePath: filePath,
		secrets:  make(map[string]interface{}),
	}
}

// loadSecrets загружает секреты из файла (ленивая загрузка)
func (p *FileProvider) loadSecrets() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.loaded {
		return nil
	}

	// Проверяем существование файла
	if _, err := os.Stat(p.filePath); os.IsNotExist(err) {
		return fmt.Errorf("secrets file not found: %s", p.filePath)
	}

	// Читаем файл
	data, err := os.ReadFile(p.filePath)
	if err != nil {
		return fmt.Errorf("failed to read secrets file: %w", err)
	}

	// Проверяем что файл не пустой
	if len(data) == 0 {
		return fmt.Errorf("secrets file is empty: %s", p.filePath)
	}

	// Парсим YAML в map[string]interface{}
	var secrets map[string]interface{}
	if err := yaml.Unmarshal(data, &secrets); err != nil {
		// Добавляем контекст к ошибке
		return fmt.Errorf("failed to parse secrets file %s: %w\nFile content preview: %s",
			p.filePath, err, string(data[:min(len(data), 100)]))
	}

	// Проверяем что получили map
	if secrets == nil {
		return fmt.Errorf("secrets file %s parsed to nil, expected key-value pairs", p.filePath)
	}

	p.secrets = secrets
	p.loaded = true
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Get возвращает секрет как строку из файла
func (p *FileProvider) Get(ctx context.Context, key string) (string, error) {
	if err := p.loadSecrets(); err != nil {
		return "", err
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	value, exists := p.secrets[key]
	if !exists {
		return "", ErrSecretNotFound
	}

	// Конвертируем в строку
	return p.convertToString(value, key)
}

// GetBytes возвращает секрет как байты из файла
// Поддерживает как обычные строки, так и base64-encoded значения
func (p *FileProvider) GetBytes(ctx context.Context, key string) ([]byte, error) {
	if err := p.loadSecrets(); err != nil {
		return nil, err
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	value, exists := p.secrets[key]
	if !exists {
		return nil, ErrSecretNotFound
	}

	switch v := value.(type) {
	case string:
		// Пробуем декодировать из base64 (для бинарных данных)
		if decoded, err := base64.StdEncoding.DecodeString(v); err == nil {
			// Проверяем что это действительно base64, а не случайная строка
			if base64.StdEncoding.EncodeToString(decoded) == v {
				return decoded, nil
			}
		}
		// Если не base64, возвращаем как есть
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		return nil, fmt.Errorf("cannot convert secret '%s' of type %T to bytes", key, v)
	}
}

// Name возвращает имя провайдера
func (p *FileProvider) Name() string {
	return fmt.Sprintf("file(%s)", p.filePath)
}

// Reload перезагружает секреты из файла
func (p *FileProvider) Reload() error {
	p.mu.Lock()
	p.loaded = false
	p.secrets = make(map[string]interface{})
	p.mu.Unlock()
	return p.loadSecrets()
}

// GetAll возвращает все секреты из файла
func (p *FileProvider) GetAll(ctx context.Context) (map[string]string, error) {
	if err := p.loadSecrets(); err != nil {
		return nil, err
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[string]string, len(p.secrets))
	for key, value := range p.secrets {
		str, err := p.convertToString(value, key)
		if err != nil {
			// Пропускаем ключи, которые не можем конвертировать
			continue
		}
		result[key] = str
	}

	return result, nil
}

// convertToString конвертирует значение в строку
func (p *FileProvider) convertToString(value interface{}, key string) (string, error) {
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

// Keys возвращает список всех ключей в файле
func (p *FileProvider) Keys() ([]string, error) {
	if err := p.loadSecrets(); err != nil {
		return nil, err
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	keys := make([]string, 0, len(p.secrets))
	for key := range p.secrets {
		keys = append(keys, key)
	}

	return keys, nil
}
