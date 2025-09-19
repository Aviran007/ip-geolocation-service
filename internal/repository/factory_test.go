package repository

import (
	"testing"

	"ip-geolocation-service/internal/config"
)

func TestNewRepositoryFactory(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Type:     "csv",
		FilePath: "/test/path.csv",
	}

	factory := NewRepositoryFactory(cfg)

	if factory == nil {
		t.Fatal("NewRepositoryFactory() returned nil")
	}

	if factory.config != cfg {
		t.Error("Factory config not set correctly")
	}
}

func TestRepositoryFactory_CreateRepository(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Type:     "csv",
		FilePath: "/test/path.csv",
	}

		factory := NewRepositoryFactory(cfg)

	tests := []struct {
		name        string
		dbType      string
		expectError bool
		description string
	}{
		{
			name:        "Valid CSV type",
			dbType:      "csv",
			expectError: false,
			description: "Should create CSV repository successfully",
		},
		{
			name:        "Invalid type",
			dbType:      "invalid",
			expectError: true,
			description: "Should return error for invalid database type",
		},
		{
			name:        "Empty type",
			dbType:      "",
			expectError: true,
			description: "Should return error for empty database type",
		},
		{
			name:        "PostgreSQL type (not implemented)",
			dbType:      "postgres",
			expectError: true,
			description: "Should return error for unimplemented database type",
		},
		{
			name:        "MySQL type (not implemented)",
			dbType:      "mysql",
			expectError: true,
			description: "Should return error for unimplemented database type",
		},
		{
			name:        "Redis type (not implemented)",
			dbType:      "redis",
			expectError: true,
			description: "Should return error for unimplemented database type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := factory.CreateRepository(tt.dbType)

			if tt.expectError {
				if err == nil {
					t.Errorf("CreateRepository() expected error, got nil. %s", tt.description)
				}
				if repo != nil {
					t.Errorf("CreateRepository() expected nil repository on error, got %v. %s", repo, tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("CreateRepository() unexpected error: %v. %s", err, tt.description)
				}
				if repo == nil {
					t.Errorf("CreateRepository() expected repository, got nil. %s", tt.description)
				}
			}
		})
	}
}

func TestRepositoryFactory_CreateRepositoryFromConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.DatabaseConfig
		expectError bool
		description string
	}{
		{
			name: "Valid CSV config",
			config: &config.DatabaseConfig{
				Type:     "csv",
				FilePath: "/test/path.csv",
			},
			expectError: false,
			description: "Should create CSV repository from config",
		},
		{
			name: "Invalid database type",
			config: &config.DatabaseConfig{
				Type:     "invalid",
				FilePath: "/test/path.csv",
			},
			expectError: true,
			description: "Should return error for invalid database type",
		},
		{
			name: "Empty config",
			config: &config.DatabaseConfig{
				Type:     "",
				FilePath: "",
			},
			expectError: true,
			description: "Should return error for empty config",
		},
		{
			name:        "Nil config",
			config:      nil,
			expectError: true,
			description: "Should return error for nil config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		factory := NewRepositoryFactory(tt.config)

			// Handle nil config case specially
			if tt.config == nil {
				// This will panic, so we expect it to panic
				defer func() {
					if r := recover(); r == nil {
						t.Error("Expected panic for nil config, but no panic occurred")
					}
				}()
				factory.CreateRepositoryFromConfig()
				return
			}

			repo, err := factory.CreateRepositoryFromConfig()

			if tt.expectError {
				if err == nil {
					t.Errorf("CreateRepositoryFromConfig() expected error, got nil. %s", tt.description)
				}
				if repo != nil {
					t.Errorf("CreateRepositoryFromConfig() expected nil repository on error, got %v. %s", repo, tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("CreateRepositoryFromConfig() unexpected error: %v. %s", err, tt.description)
				}
				if repo == nil {
					t.Errorf("CreateRepositoryFromConfig() expected repository, got nil. %s", tt.description)
				}
			}
		})
	}
}

