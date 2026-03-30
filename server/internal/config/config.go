package config

import "os"

type Config struct {
	DBServiceBaseURL string
}

func Load() *Config {
	baseURL := os.Getenv("DB_SERVICE_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}

	return &Config{
		DBServiceBaseURL: baseURL,
	}
}
