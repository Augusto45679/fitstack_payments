// Package config handles application configuration.
package config

import "os"

// Config holds all configuration values.
type Config struct {
	Server  ServerConfig
	Django  DjangoConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port    string
	GinMode string
}

// DjangoConfig holds Django backend configuration.
type DjangoConfig struct {
	BaseURL string
	APIKey  string
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:    getEnv("PORT", "8080"),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		Django: DjangoConfig{
			BaseURL: getEnv("DJANGO_BACKEND_URL", "http://localhost:8000"),
			APIKey:  getEnv("DJANGO_API_KEY", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
