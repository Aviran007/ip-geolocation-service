package services

import (
	"context"
	"errors"

	"ip-geolocation-service/internal/models"
)

// MockRepository implements repository.IPRepository for testing
type MockRepository struct {
	locations map[string]*models.Location
	initErr   error
	closeErr  error
	healthErr error
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		locations: make(map[string]*models.Location),
	}
}

func (m *MockRepository) FindLocation(ctx context.Context, ip string) (*models.Location, error) {
	if location, exists := m.locations[ip]; exists {
		return location, nil
	}
	return nil, errors.New("location not found for IP: " + ip)
}

func (m *MockRepository) Initialize(ctx context.Context) error {
	return m.initErr
}

func (m *MockRepository) Close() error {
	return m.closeErr
}

func (m *MockRepository) HealthCheck(ctx context.Context) error {
	return m.healthErr
}

func (m *MockRepository) SetLocation(ip string, location *models.Location) {
	m.locations[ip] = location
}

func (m *MockRepository) SetInitError(err error) {
	m.initErr = err
}

func (m *MockRepository) SetCloseError(err error) {
	m.closeErr = err
}

func (m *MockRepository) SetHealthError(err error) {
	m.healthErr = err
}
