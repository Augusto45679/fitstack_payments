// Package config handles loading and managing application configuration.
package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the application.
type Config struct {
	// Server configuration
	Server ServerConfig

	// FitStack Core API configuration
	Core CoreConfig

	// Security settings
	Security SecurityConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port    string
	GinMode string // "debug", "release", or "test"
}

// CoreConfig holds FitStack Core API configuration.
type CoreConfig struct {
	BaseURL string
	APIKey  string
}

// SecurityConfig holds security-related configuration.
type SecurityConfig struct {
	EncryptionKey    string
	WebhookSecret    string
	JWTValidationURL string // URL to validate JWT tokens (optional)
}

// Load reads configuration from environment variables.
// Returns a Config struct with all settings populated.
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:    getEnv("PORT", "8080"),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		Core: CoreConfig{
			BaseURL: getEnv("FITSTACK_CORE_URL", "http://localhost:8000"),
			APIKey:  getEnv("FITSTACK_CORE_API_KEY", ""),
		},
		Security: SecurityConfig{
			EncryptionKey:    getEnv("ENCRYPTION_KEY", ""),
			WebhookSecret:    getEnv("MP_WEBHOOK_SECRET", ""),
			JWTValidationURL: getEnv("JWT_VALIDATION_URL", ""),
		},
	}
}

// getEnv retrieves an environment variable with a fallback default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt retrieves an environment variable as an integer with a fallback.
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvBool retrieves an environment variable as a boolean with a fallback.
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
