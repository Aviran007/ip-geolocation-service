package handlers

import (
	"context"
	"errors"

	"ip-geolocation-service/internal/models"
)

// MockIPService implements services.IPService for testing
type MockIPService struct {
	locations map[string]*models.Location
	errors    map[string]error
	healthErr error
}

func NewMockIPService() *MockIPService {
	return &MockIPService{
		locations: make(map[string]*models.Location),
		errors:    make(map[string]error),
	}
}

func (m *MockIPService) FindLocation(ctx context.Context, ip string) (*models.Location, error) {
	if err, exists := m.errors[ip]; exists {
		return nil, err
	}
	if location, exists := m.locations[ip]; exists {
		return location, nil
	}
	return nil, errors.New("location not found for IP: " + ip)
}

func (m *MockIPService) HealthCheck(ctx context.Context) error {
	return m.healthErr
}

func (m *MockIPService) SetLocation(ip string, location *models.Location) {
	m.locations[ip] = location
}

func (m *MockIPService) SetError(ip string, err error) {
	m.errors[ip] = err
}

func (m *MockIPService) SetHealthError(err error) {
	m.healthErr = err
}
