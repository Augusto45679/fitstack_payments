// Package config provides database configuration (placeholder).
package config

// DatabaseConfig holds database connection settings.
// This is a placeholder for future MySQL/Redis integration.
type DatabaseConfig struct {
	MySQL MySQLConfig
	Redis RedisConfig
}

// MySQLConfig holds MySQL connection settings.
type MySQLConfig struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// LoadDatabaseConfig loads database configuration.
func LoadDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		MySQL: MySQLConfig{
			Host:     getEnv("MYSQL_HOST", "localhost"),
			Port:     getEnv("MYSQL_PORT", "3306"),
			Database: getEnv("MYSQL_DATABASE", "fitstack_payments"),
			User:     getEnv("MYSQL_USER", "root"),
			Password: getEnv("MYSQL_PASSWORD", ""),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
	}
}
