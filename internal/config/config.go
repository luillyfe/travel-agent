package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Config struct {
	ServerPort string
	LogLevel   string
	AIProvider AIProviderConfig
}

type AIProviderConfig struct {
	APIKey         string        `json:"api_key" envconfig:"AI_PROVIDER_API_KEY" required:"true"`
	Timeout        time.Duration `json:"timeout" envconfig:"AI_PROVIDER_TIMEOUT" default:"30s"`
	RetryAttempts  int           `json:"retry_attempts" envconfig:"AI_PROVIDER_RETRY_ATTEMPTS" default:"3"`
	RetryDelay     time.Duration `json:"retry_delay" envconfig:"AI_PROVIDER_RETRY_DELAY" default:"1s"`
	MaxConcurrency int           `json:"max_concurrency" envconfig:"AI_PROVIDER_MAX_CONCURRENCY" default:"10"`
	Model          string        `json:"model" envconfig:"AI_PROVIDER_MODEL" default:"gpt-4"`
	Temperature    float64       `json:"temperature" envconfig:"AI_PROVIDER_TEMPERATURE" default:"0.7"`
}

func Load(filename string) (*Config, error) {
	// Read the configuration file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	// Parse JSON into Config struct
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	// Set defaults if not specified
	if cfg.ServerPort == "" {
		cfg.ServerPort = ":8080"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	// Set defaults for AIProvider
	setDefaultsAIProvider(&cfg)

	return &cfg, nil
}

// setDefaultsAIProvider ensures the AIProvider config has appropriate default values
func setDefaultsAIProvider(config *Config) {
	if config.AIProvider.Model == "" {
		config.AIProvider.Model = "gpt-4"
	}

	if config.AIProvider.Timeout == 0 {
		config.AIProvider.Timeout = 30 * time.Second
	}

	if config.AIProvider.RetryAttempts == 0 {
		config.AIProvider.RetryAttempts = 3
	}

	if config.AIProvider.RetryDelay == 0 {
		config.AIProvider.RetryDelay = time.Second
	}

	if config.AIProvider.MaxConcurrency == 0 {
		config.AIProvider.MaxConcurrency = 10
	}

	if config.AIProvider.Temperature == 0 {
		config.AIProvider.Temperature = 0.7
	}
}

// Example JSON configuration file (config.json)
// const exampleConfig = `{
//     "environment": "development",
//     "aiprovider": {
//         "api_key": "your-api-key-here",
//         "model": "gpt-4",
//         "timeout": "30s",
//         "retry_attempts": 3,
//         "retry_delay": "1s",
//         "max_concurrency": 10,
//         "request_rate_limit": 3000,
//         "token_rate_limit": 250000,
//         "max_tokens_per_req": 4000,
//         "temperature": 0.7
//     }
//     // ... other config sections ...
// }`
