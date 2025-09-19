package services

import (
	"context"
	"fmt"
	"time"

	"ip-geolocation-service/internal/models"
	"ip-geolocation-service/internal/repository"
)

// IPService defines the interface for IP location services
type IPService interface {
	FindLocation(ctx context.Context, ip string) (*models.Location, error)
	HealthCheck(ctx context.Context) error
}

// IPServiceImpl implements IPService
type IPServiceImpl struct {
	repository repository.IPRepository
	validator  *models.IPValidator
}

// NewIPService creates a new IP service
func NewIPService(repo repository.IPRepository) IPService {
	return &IPServiceImpl{
		repository: repo,
		validator:  models.NewIPValidator(),
	}
}

// FindLocation finds the location for a given IP address
func (s *IPServiceImpl) FindLocation(ctx context.Context, ip string) (*models.Location, error) {
	// Validate input
	if err := s.validator.ValidateIP(ip); err != nil {
		return nil, fmt.Errorf("invalid IP address: %w", err)
	}

	// Normalize IP for consistent lookup
	normalizedIP := s.validator.NormalizeIP(ip)

	// Add timeout to context if not already present
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Find location in repository
	location, err := s.repository.FindLocation(ctx, normalizedIP)
	if err != nil {
		return nil, fmt.Errorf("failed to find location: %w", err)
	}

	// Validate location data
	if err := location.ValidateLocation(); err != nil {
		return nil, fmt.Errorf("invalid location data: %w", err)
	}

	return location, nil
}

// HealthCheck checks if the service is healthy
func (s *IPServiceImpl) HealthCheck(ctx context.Context) error {
	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Check repository health
	if err := s.repository.HealthCheck(ctx); err != nil {
		return fmt.Errorf("repository health check failed: %w", err)
	}

	return nil
}
