package repository

import (
	"fmt"
	"ip-geolocation-service/internal/config"
)

// RepositoryFactoryImpl implements RepositoryFactory
type RepositoryFactoryImpl struct {
	config  *config.DatabaseConfig
	metrics RepositoryMetrics
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(cfg *config.DatabaseConfig, metrics RepositoryMetrics) *RepositoryFactoryImpl {
	return &RepositoryFactoryImpl{
		config:  cfg,
		metrics: metrics,
	}
}

// CreateRepository creates a repository instance based on the database type
func (f *RepositoryFactoryImpl) CreateRepository(dbType string) (IPRepository, error) {
	switch dbType {
	case config.DatabaseTypeCSV:
		return NewFileRepository(f.config, f.metrics), nil
	case config.DatabaseTypeJSON:
		// TODO: Implement JSON file repository
		return nil, fmt.Errorf("json repository not implemented yet")
	case config.DatabaseTypeXML:
		// TODO: Implement XML file repository
		return nil, fmt.Errorf("xml repository not implemented yet")
	case config.DatabaseTypePostgres:
		// TODO: Implement PostgreSQL repository
		return nil, fmt.Errorf("postgres repository not implemented yet")
	case config.DatabaseTypeMySQL:
		// TODO: Implement MySQL repository
		return nil, fmt.Errorf("mysql repository not implemented yet")
	case config.DatabaseTypeRedis:
		// TODO: Implement Redis repository
		return nil, fmt.Errorf("redis repository not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

// CreateRepositoryFromConfig creates a repository using the factory's configuration
func (f *RepositoryFactoryImpl) CreateRepositoryFromConfig() (IPRepository, error) {
	return f.CreateRepository(f.config.Type)
}
