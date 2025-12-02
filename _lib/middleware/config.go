package platform_middleware

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Client struct {
		GRPC ClientGRPCConfig `yaml:"grpc"`
	} `yaml:"client"`
	Server   ServerConfig  `yaml:"server"`
	Tracing  TracingConfig `yaml:"tracing"`
	Admin    AdminConfig   `yaml:"admin"`
	Auth     *AuthConfig   `yaml:"auth,omitempty"` // Опциональный блок
	Shutdown struct {
		GracePeriod string `yaml:"gracePeriod"`
	} `yaml:"shutdown"`
}

// AuthConfig конфигурация для auth middleware
type AuthConfig struct {
	Enabled  bool             `yaml:"enabled"`
	Issuer   string           `yaml:"issuer"`
	Audience []string         `yaml:"audience"`
	JWKS     AuthJWKSConfig   `yaml:"jwks"`
	Required bool             `yaml:"required"`
	Public   AuthPublicConfig `yaml:"public"`

	// Parsed values (заполняются после парсинга)
	cacheTTLDuration       time.Duration
	refreshTimeoutDuration time.Duration
}

// AuthJWKSConfig конфигурация JWKS клиента
type AuthJWKSConfig struct {
	URL            string `yaml:"url"`
	CacheTTL       string `yaml:"cacheTtl"`
	RefreshTimeout string `yaml:"refreshTimeout"`
}

// AuthPublicConfig конфигурация публичных методов (без токена)
type AuthPublicConfig struct {
	Methods []string `yaml:"methods"` // gRPC методы
	Paths   []string `yaml:"paths"`   // HTTP пути
}

type ClientGRPCConfig struct {
	Timeout string `yaml:"timeout"`
	Retry   struct {
		MaxAttempts uint `yaml:"maxAttempts"`
		Backoff     struct {
			Base   string `yaml:"base"`
			Max    string `yaml:"max"`
			Jitter bool   `yaml:"jitter"`
		} `yaml:"backoff"`
		RetryableCodes []string `yaml:"retryableCodes"`
	} `yaml:"retry"`
	CircuitBreaker struct {
		FailuresForOpen  int    `yaml:"failuresForOpen"`
		Window           string `yaml:"window"`
		HalfOpenMaxCalls int    `yaml:"halfOpenMaxCalls"`
		OpenStateFor     string `yaml:"openStateFor"`
	} `yaml:"circuitBreaker"`
	Metrics bool `yaml:"metrics"`
}

type ServerConfig_Timeout struct {
	Enabled   bool     `yaml:"enabled"`
	Ignore    []string `yaml:"ignore"`
	TimeoutMs int      `yaml:"timeoutMs"`
	Paths     []struct {
		Path      string `yaml:"path"`
		TimeoutMs int    `yaml:"timeoutMs"`
	} `yaml:"paths"`
}

type ServerConfig_RateLimit struct {
	Enabled   bool     `yaml:"enabled"`
	Ignore    []string `yaml:"ignore"`
	ReqPerSec int      `yaml:"reqPerSec"`
	Paths     []struct {
		Path      string `yaml:"path"`
		ReqPerSec int    `yaml:"reqPerSec"`
	} `yaml:"paths"`
}

type ServerConfig struct {
	Timeout   ServerConfig_Timeout   `yaml:"timeout"`
	RateLimit ServerConfig_RateLimit `yaml:"rateLimit"`
}

type TracingConfig struct {
	Enabled      bool    `yaml:"enabled"`
	AgentHost    string  `yaml:"agentHost"`
	HealthHost   string  `yaml:"healthHost"`
	SamplerType  string  `yaml:"samplerType"`
	SamplerParam float64 `yaml:"samplerParam"`
	LogSpans     bool    `yaml:"logSpans"`
}

type AdminConfig struct {
	Enabled bool               `yaml:"enabled"`
	Host    string             `yaml:"host"`
	Port    int                `yaml:"port"`
	Metrics AdminMetricsConfig `yaml:"metrics"`
	Pprof   AdminPprofConfig   `yaml:"pprof"`
}

type AdminMetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

type AdminPprofConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

func Load(filePath string) (*Config, error) {
	var cfg Config

	ApplyDefaults(&cfg)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", filePath)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Валидация конфига
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// Validate валидирует весь конфиг
func (c *Config) Validate() error {
	// Валидация auth конфига (если он есть)
	if c.Auth != nil {
		if err := c.validateAuth(); err != nil {
			return fmt.Errorf("auth: %w", err)
		}
	}

	return nil
}

// validateAuth валидирует auth конфигурацию
func (c *Config) validateAuth() error {
	if !c.Auth.Enabled {
		return nil // Если отключено, валидация не нужна
	}

	// Проверка обязательных полей
	if c.Auth.Issuer == "" {
		return fmt.Errorf("issuer is required")
	}
	if c.Auth.JWKS.URL == "" {
		return fmt.Errorf("jwks.url is required")
	}

	// Парсинг duration для JWKS
	if c.Auth.JWKS.CacheTTL != "" {
		cacheTTL, err := time.ParseDuration(c.Auth.JWKS.CacheTTL)
		if err != nil {
			return fmt.Errorf("invalid jwks.cacheTtl: %w", err)
		}
		c.Auth.cacheTTLDuration = cacheTTL
	} else {
		c.Auth.cacheTTLDuration = 5 * time.Minute // default
	}

	if c.Auth.JWKS.RefreshTimeout != "" {
		refreshTimeout, err := time.ParseDuration(c.Auth.JWKS.RefreshTimeout)
		if err != nil {
			return fmt.Errorf("invalid jwks.refreshTimeout: %w", err)
		}
		c.Auth.refreshTimeoutDuration = refreshTimeout
	} else {
		c.Auth.refreshTimeoutDuration = 2 * time.Second // default
	}

	return nil
}

func ApplyDefaults(cfg *Config) {
	// Client defaults
	if cfg.Client.GRPC.Timeout == "" {
		cfg.Client.GRPC.Timeout = "5s"
	}

	// Tracing defaults
	if cfg.Tracing.AgentHost == "" {
		cfg.Tracing.AgentHost = "localhost:6831"
	}
	if cfg.Tracing.HealthHost == "" {
		cfg.Tracing.HealthHost = "localhost:14271"
	}
	if cfg.Tracing.SamplerType == "" {
		cfg.Tracing.SamplerType = "const"
	}
	if cfg.Tracing.SamplerParam == 0 {
		cfg.Tracing.SamplerParam = 1.0
	}

	// Admin defaults
	if cfg.Admin.Host == "" {
		cfg.Admin.Host = "0.0.0.0"
	}
	if cfg.Admin.Port == 0 {
		cfg.Admin.Port = 9090
	}
	if cfg.Admin.Metrics.Path == "" {
		cfg.Admin.Metrics.Path = "/metrics"
	}
	if cfg.Admin.Pprof.Path == "" {
		cfg.Admin.Pprof.Path = "/debug/pprof"
	}

	// Shutdown defaults
	if cfg.Shutdown.GracePeriod == "" {
		cfg.Shutdown.GracePeriod = "30s"
	}

	// Auth defaults (если блок существует)
	if cfg.Auth != nil {
		if cfg.Auth.JWKS.CacheTTL == "" {
			cfg.Auth.JWKS.CacheTTL = "5m"
		}
		if cfg.Auth.JWKS.RefreshTimeout == "" {
			cfg.Auth.JWKS.RefreshTimeout = "2s"
		}
		if cfg.Auth.Audience == nil {
			cfg.Auth.Audience = []string{}
		}
		if cfg.Auth.Public.Methods == nil {
			cfg.Auth.Public.Methods = []string{}
		}
	}
}

// IsAuthEnabled проверяет, включена ли аутентификация
func (c *Config) IsAuthEnabled() bool {
	return c.Auth != nil && c.Auth.Enabled
}

// GetAuthCacheTTL возвращает распарсенный cache TTL
func (c *AuthConfig) GetCacheTTL() time.Duration {
	if c.cacheTTLDuration == 0 {
		return 5 * time.Minute
	}
	return c.cacheTTLDuration
}

// GetAuthRefreshTimeout возвращает распарсенный refresh timeout
func (c *AuthConfig) GetRefreshTimeout() time.Duration {
	if c.refreshTimeoutDuration == 0 {
		return 2 * time.Second
	}
	return c.refreshTimeoutDuration
}

// IsPublicMethod проверяет, является ли метод публичным
func (c *AuthConfig) IsPublicMethod(fullMethod string) bool {
	for _, method := range c.Public.Methods {
		if method == fullMethod {
			return true
		}
	}
	return false
}

func (c *AuthConfig) IsPublicPath(path string) bool {
	for _, publicPath := range c.Public.Paths {
		if path == publicPath || strings.HasPrefix(path, publicPath) {
			return true
		}
	}
	return false
}
