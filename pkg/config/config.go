package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server         ServerConfig         `yaml:"server"`
	Providers      []ProviderConfig     `yaml:"providers"`
	Health         HealthConfig         `yaml:"health"`
	Routing        RoutingConfig        `yaml:"routing"`
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`
	Redis          RedisConfig          `yaml:"redis"`
	Caching        CachingConfig        `yaml:"caching"`
}

// ServerConfig contains server settings
type ServerConfig struct {
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// ProviderConfig contains provider settings
type ProviderConfig struct {
	Name           string  `yaml:"name"`
	URL            string  `yaml:"url"`
	Priority       int     `yaml:"priority"`
	CostPerRequest float64 `yaml:"cost_per_request"`
}

// HealthConfig contains health check settings
type HealthConfig struct {
	CheckInterval      time.Duration `yaml:"check_interval"`
	Timeout            time.Duration `yaml:"timeout"`
	UnhealthyThreshold int           `yaml:"unhealthy_threshold"`
}

// RoutingConfig contains routing settings
type RoutingConfig struct {
	Strategy     string        `yaml:"strategy"`
	MaxRetries   int           `yaml:"max_retries"`
	RetryBackoff time.Duration `yaml:"retry_backoff"`
}

// CircuitBreakerConfig contains circuit breaker settings
type CircuitBreakerConfig struct {
	MaxRequests uint32        `yaml:"max_requests"`
	Timeout     time.Duration `yaml:"timeout"`
}

// RedisConfig contains Redis settings
type RedisConfig struct {
	URL string `yaml:"url"`
	DB  int    `yaml:"db"`
}

// CachingConfig contains settings for request caching
type CachingConfig struct {
	Enabled bool                     `yaml:"enabled"`
	Methods map[string]time.Duration `yaml:"methods"`
}

// Load reads and parses the configuration file
func Load(configPath string) (*Config, error) {
	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables
	expanded := os.ExpandEnv(string(data))

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if len(c.Providers) == 0 {
		return fmt.Errorf("at least one provider must be configured")
	}

	for i, p := range c.Providers {
		if p.Name == "" {
			return fmt.Errorf("provider %d: name is required", i)
		}
		if p.URL == "" {
			return fmt.Errorf("provider %s: URL is required", p.Name)
		}
		if !strings.HasPrefix(p.URL, "http://") && !strings.HasPrefix(p.URL, "https://") {
			return fmt.Errorf("provider %s: URL must start with http:// or https://", p.Name)
		}
		if p.CostPerRequest < 0 {
			return fmt.Errorf("provider %s: cost_per_request must be non-negative", p.Name)
		}
	}

	if c.Routing.MaxRetries < 0 {
		return fmt.Errorf("max_retries must be non-negative")
	}

	return nil
}
