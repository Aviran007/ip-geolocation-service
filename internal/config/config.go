package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	RateLimit RateLimitConfig
	Logging   LoggingConfig
}

// Database types
const (
	DatabaseTypeCSV      = "csv"
	DatabaseTypeJSON     = "json"
	DatabaseTypeXML      = "xml"
	DatabaseTypePostgres = "postgres"
	DatabaseTypeMySQL    = "mysql"
	DatabaseTypeRedis    = "redis"
)

// Log levels
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// Log formats
const (
	LogFormatJSON = "json"
	LogFormatText = "text"
)

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Type     string
	FilePath string
	Host     string
	Port     int
	Username string
	Password string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond int
	BurstSize         int
	// Cleanup configuration
	CleanupInterval   time.Duration // How often to run cleanup (default: 1 minute)
	InactiveThreshold time.Duration // How long before client is considered inactive (default: 5 minutes)
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  getDurationEnv("READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDurationEnv("IDLE_TIMEOUT", 120*time.Second),
		},
		Database: DatabaseConfig{
			Type:     getEnv("DATABASE_TYPE", DatabaseTypeCSV),
			FilePath: getEnv("DATABASE_FILE_PATH", "./data/ip_locations.csv"),
			Host:     getEnv("DATABASE_HOST", "localhost"),
			Port:     getIntEnv("DATABASE_PORT", 5432),
			Username: getEnv("DATABASE_USERNAME", ""),
			Password: getEnv("DATABASE_PASSWORD", ""),
		},
		RateLimit: RateLimitConfig{
			RequestsPerSecond: getIntEnv("RATE_LIMIT_RPS", 20),
			BurstSize:         getIntEnv("RATE_LIMIT_BURST", 20),
			CleanupInterval:   getDurationEnv("RATE_LIMIT_CLEANUP_INTERVAL", 1*time.Minute),
			InactiveThreshold: getDurationEnv("RATE_LIMIT_INACTIVE_THRESHOLD", 5*time.Minute),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", LogLevelInfo),
			Format: getEnv("LOG_FORMAT", LogFormatJSON),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server config
	if c.Server.Port == "" {
		return fmt.Errorf("server port cannot be empty")
	}

	// Validate database config
	validDBTypes := []string{DatabaseTypeCSV, DatabaseTypePostgres, DatabaseTypeMySQL, DatabaseTypeRedis}
	if !contains(validDBTypes, c.Database.Type) {
		return fmt.Errorf("invalid database type: %s, must be one of: %s",
			c.Database.Type, strings.Join(validDBTypes, ", "))
	}

	if c.Database.Type == "csv" && c.Database.FilePath == "" {
		return fmt.Errorf("database file path is required when using CSV database")
	}

	// Validate rate limit config
	if c.RateLimit.RequestsPerSecond <= 0 {
		return fmt.Errorf("rate limit requests per second must be positive")
	}

	if c.RateLimit.BurstSize <= 0 {
		return fmt.Errorf("rate limit burst size must be positive")
	}

	// Validate logging config
	validLogLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLogLevels, c.Logging.Level) {
		return fmt.Errorf("invalid log level: %s, must be one of: %s",
			c.Logging.Level, strings.Join(validLogLevels, ", "))
	}

	validLogFormats := []string{"json", "text"}
	if !contains(validLogFormats, c.Logging.Format) {
		return fmt.Errorf("invalid log format: %s, must be one of: %s",
			c.Logging.Format, strings.Join(validLogFormats, ", "))
	}

	return nil
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetServerAddress returns the full server address
func (c *Config) GetServerAddress() string {
	return ":" + c.Server.Port
}
