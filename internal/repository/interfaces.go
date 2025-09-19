package repository

import (
	"context"
	"ip-geolocation-service/internal/models"
)

// IPRepository defines the interface for IP location data access
type IPRepository interface {
	// FindLocation finds the location for a given IP address
	FindLocation(ctx context.Context, ip string) (*models.Location, error)

	// Initialize initializes the repository (loads data, connects to DB, etc.)
	Initialize(ctx context.Context) error

	// Close closes the repository and cleans up resources
	Close() error

	// HealthCheck checks if the repository is healthy
	HealthCheck(ctx context.Context) error
}

// RepositoryFactory creates repository instances based on configuration
type RepositoryFactory interface {
	CreateRepository(dbType string) (IPRepository, error)
}

