package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Test with default values
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	// Test default values
	if cfg.Server.Port != "8080" {
		t.Errorf("Expected port 8080, got %s", cfg.Server.Port)
	}

	if cfg.Database.Type != DatabaseTypeCSV {
		t.Errorf("Expected database type %s, got %s", DatabaseTypeCSV, cfg.Database.Type)
	}

	if cfg.RateLimit.RequestsPerSecond != 20 {
		t.Errorf("Expected RPS 20, got %d", cfg.RateLimit.RequestsPerSecond)
	}
}

func TestLoadConfig_WithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("PORT", "9090")
	os.Setenv("DATABASE_TYPE", "csv")
	os.Setenv("RATE_LIMIT_RPS", "200")
	os.Setenv("LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DATABASE_TYPE")
		os.Unsetenv("RATE_LIMIT_RPS")
		os.Unsetenv("LOG_LEVEL")
	}()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	// Test environment variable values
	if cfg.Server.Port != "9090" {
		t.Errorf("Expected port 9090, got %s", cfg.Server.Port)
	}

	if cfg.Database.Type != "csv" {
		t.Errorf("Expected database type csv, got %s", cfg.Database.Type)
	}

	if cfg.RateLimit.RequestsPerSecond != 200 {
		t.Errorf("Expected RPS 200, got %d", cfg.RateLimit.RequestsPerSecond)
	}

	if cfg.Logging.Level != "debug" {
		t.Errorf("Expected log level debug, got %s", cfg.Logging.Level)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Server: ServerConfig{
					Port:         "8080",
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
					IdleTimeout:  120 * time.Second,
				},
				Database: DatabaseConfig{
					Type:     DatabaseTypeCSV,
					FilePath: "./data/test.csv",
				},
				RateLimit: RateLimitConfig{
					RequestsPerSecond: 20,
					BurstSize:         20,
				},
				Logging: LoggingConfig{
					Level:  LogLevelInfo,
					Format: LogFormatJSON,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid database type",
			config: &Config{
				Server: ServerConfig{
					Port: "8080",
				},
				Database: DatabaseConfig{
					Type: "invalid",
				},
				RateLimit: RateLimitConfig{
					RequestsPerSecond: 20,
					BurstSize:         20,
				},
				Logging: LoggingConfig{
					Level:  LogLevelInfo,
					Format: LogFormatJSON,
				},
			},
			wantErr: true,
		},
		{
			name: "missing CSV file path",
			config: &Config{
				Server: ServerConfig{
					Port: "8080",
				},
				Database: DatabaseConfig{
					Type:     DatabaseTypeCSV,
					FilePath: "",
				},
				RateLimit: RateLimitConfig{
					RequestsPerSecond: 20,
					BurstSize:         20,
				},
				Logging: LoggingConfig{
					Level:  LogLevelInfo,
					Format: LogFormatJSON,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid rate limit",
			config: &Config{
				Server: ServerConfig{
					Port: "8080",
				},
				Database: DatabaseConfig{
					Type:     DatabaseTypeCSV,
					FilePath: "./data/test.csv",
				},
				RateLimit: RateLimitConfig{
					RequestsPerSecond: 0, // Invalid
					BurstSize:         20,
				},
				Logging: LoggingConfig{
					Level:  LogLevelInfo,
					Format: LogFormatJSON,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			config: &Config{
				Server: ServerConfig{
					Port: "8080",
				},
				Database: DatabaseConfig{
					Type:     DatabaseTypeCSV,
					FilePath: "./data/test.csv",
				},
				RateLimit: RateLimitConfig{
					RequestsPerSecond: 20,
					BurstSize:         20,
				},
				Logging: LoggingConfig{
					Level:  "invalid",
					Format: LogFormatJSON,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetServerAddress(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: "8080",
		},
	}

	expected := ":8080"
	actual := cfg.GetServerAddress()

	if actual != expected {
		t.Errorf("GetServerAddress() = %v, want %v", actual, expected)
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test getEnv
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	if got := getEnv("TEST_VAR", "default"); got != "test_value" {
		t.Errorf("getEnv() = %v, want test_value", got)
	}

	if got := getEnv("NONEXISTENT_VAR", "default"); got != "default" {
		t.Errorf("getEnv() = %v, want default", got)
	}

	// Test getIntEnv
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	if got := getIntEnv("TEST_INT", 0); got != 42 {
		t.Errorf("getIntEnv() = %v, want 42", got)
	}

	if got := getIntEnv("NONEXISTENT_INT", 99); got != 99 {
		t.Errorf("getIntEnv() = %v, want 99", got)
	}

	// Test getDurationEnv
	os.Setenv("TEST_DURATION", "30s")
	defer os.Unsetenv("TEST_DURATION")

	if got := getDurationEnv("TEST_DURATION", 0); got != 30*time.Second {
		t.Errorf("getDurationEnv() = %v, want 30s", got)
	}

	if got := getDurationEnv("NONEXISTENT_DURATION", 60*time.Second); got != 60*time.Second {
		t.Errorf("getDurationEnv() = %v, want 60s", got)
	}

	// Test contains
	slice := []string{"a", "b", "c"}
	if !contains(slice, "a") {
		t.Error("contains() should return true for existing item")
	}

	if contains(slice, "d") {
		t.Error("contains() should return false for non-existing item")
	}
}
