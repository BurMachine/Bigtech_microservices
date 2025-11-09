package platform_middleware

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Client struct {
		GRPC ClientGRPCConfig `yaml:"grpc"`
	} `yaml:"client"`
	Server   ServerConfig `yaml:"server"`
	Shutdown struct {
		GracePeriod string `yaml:"gracePeriod"` // e.g., "30s" (добавьте для shutdown)
	} `yaml:"shutdown"`
}

// ClientGRPCConfig как в моём предыдущем ответе (timeout, retry, circuitBreaker, metrics)
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

// ServerConfig_RateLimit вложенный тип для rate limit конфига
type ServerConfig_RateLimit struct {
	Enabled   bool     `yaml:"enabled"`
	Ignore    []string `yaml:"ignore"`
	ReqPerSec int      `yaml:"reqPerSec"`
	Paths     []struct {
		Path      string `yaml:"path"`
		ReqPerSec int    `yaml:"reqPerSec"`
	} `yaml:"paths"`
}

// ServerConfig как в ДЗ (timeout, rateLimit)
type ServerConfig struct {
	Timeout   ServerConfig_Timeout   `yaml:"timeout"`
	RateLimit ServerConfig_RateLimit `yaml:"rateLimit"`
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
	if cfg.Client.GRPC.Timeout == "" {
		cfg.Client.GRPC.Timeout = "5s"
	}

	if cfg.Shutdown.GracePeriod == "" {
		cfg.Shutdown.GracePeriod = "30s"
	}
}
