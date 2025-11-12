package platform_middleware

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Client struct {
		GRPC ClientGRPCConfig `yaml:"grpc"`
	} `yaml:"client"`
	Server   ServerConfig  `yaml:"server"`
	Tracing  TracingConfig `yaml:"tracing"`
	Shutdown struct {
		GracePeriod string `yaml:"gracePeriod"`
	} `yaml:"shutdown"`
}

// ClientGRPCConfig конфигурация для gRPC клиента
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

// ServerConfig_Timeout конфигурация timeout для сервера
type ServerConfig_Timeout struct {
	Enabled   bool     `yaml:"enabled"`
	Ignore    []string `yaml:"ignore"`
	TimeoutMs int      `yaml:"timeoutMs"`
	Paths     []struct {
		Path      string `yaml:"path"`
		TimeoutMs int    `yaml:"timeoutMs"`
	} `yaml:"paths"`
}

// ServerConfig_RateLimit конфигурация rate limit для сервера
type ServerConfig_RateLimit struct {
	Enabled   bool     `yaml:"enabled"`
	Ignore    []string `yaml:"ignore"`
	ReqPerSec int      `yaml:"reqPerSec"`
	Paths     []struct {
		Path      string `yaml:"path"`
		ReqPerSec int    `yaml:"reqPerSec"`
	} `yaml:"paths"`
}

// ServerConfig конфигурация для gRPC/HTTP сервера
type ServerConfig struct {
	Timeout   ServerConfig_Timeout   `yaml:"timeout"`
	RateLimit ServerConfig_RateLimit `yaml:"rateLimit"`
}

// TracingConfig конфигурация для Jaeger трейсинга
type TracingConfig struct {
	Enabled      bool    `yaml:"enabled"`
	AgentHost    string  `yaml:"agentHost"`  // UDP порт для spans (6831)
	HealthHost   string  `yaml:"healthHost"` // TCP порт для health check (14271)
	SamplerType  string  `yaml:"samplerType"`
	SamplerParam float64 `yaml:"samplerParam"`
	LogSpans     bool    `yaml:"logSpans"`
}

func Load(filePath string) (*Config, error) {
	var cfg Config

	ApplyDefaults(&cfg)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return &cfg, nil
	}

	// Unmarshal YAML
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
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

	// Shutdown defaults
	if cfg.Shutdown.GracePeriod == "" {
		cfg.Shutdown.GracePeriod = "30s"
	}
}
