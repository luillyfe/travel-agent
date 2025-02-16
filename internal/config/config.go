package config

type Config struct {
	ServerPort string
	LogLevel   string
}

func Load() (*Config, error) {
	return &Config{
		ServerPort: ":8080",
		LogLevel:   "info",
	}, nil
}
