package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	ServerPort string
	LogLevel   string
	AIProvider AIProviderConfig
}

type AIProviderConfig struct {
	APIKey string `json:"api_key" required:"true"`
}

func Load(filename string) (*Config, error) {
	// Check if API key is set in environment
	apiKey := os.Getenv("AI_PROVIDER_API_KEY")

	// Read the configuration file
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// If file doesn't exist, return default config
			cfg := &Config{
				ServerPort: ":8080",
				LogLevel:   "info",
				AIProvider: AIProviderConfig{
					APIKey: apiKey,
				},
			}
			return cfg, nil
		}
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
	if cfg.AIProvider.APIKey == "" {
		cfg.AIProvider.APIKey = apiKey
	}

	return &cfg, nil
}

// Example usage of config.json:
/*
{
    "ServerPort": ":8080",           // The port the server will listen on
    "LogLevel": "info",              // Logging level (debug, info, warn, error)
    "AIProvider": {
        "api_key": ""                // AI Provider API key
    }
}

Configuration can be provided via:
1. config.json file
2. Environment variables:
   - AI_PROVIDER_API_KEY: Override the API key from config.json

Default values:
- ServerPort: ":8080"
- LogLevel: "info"
- AIProvider.api_key: Must be provided either in config.json or via environment variable
*/
