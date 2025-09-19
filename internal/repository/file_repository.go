package repository

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"ip-geolocation-service/internal/config"
	"ip-geolocation-service/internal/models"
)

// FileRepository implements IPRepository using a file-based storage (CSV format)
type FileRepository struct {
	config   *config.DatabaseConfig
	data     map[string]*models.Location
	mu       sync.RWMutex
	loaded   bool
	loadTime time.Time
	metrics  RepositoryMetrics
}

// NewFileRepository creates a new file-based repository (CSV format)
func NewFileRepository(cfg *config.DatabaseConfig, metrics RepositoryMetrics) *FileRepository {
	return &FileRepository{
		config:  cfg,
		data:    make(map[string]*models.Location),
		metrics: metrics,
	}
}

// Initialize loads the CSV data into memory
func (r *FileRepository) Initialize(ctx context.Context) error {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RecordLookupTime(time.Since(start).Seconds())
		}
	}()

	file, err := os.Open(r.config.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open data file %s: %w", r.config.FilePath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 3 // ip, city, country

	// Skip header if it exists
	firstRecord, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read first record: %w", err)
	}

	// Check if first record is a header (contains non-IP values)
	if !isValidIP(firstRecord[0]) {
		// This is a header, continue reading
	} else {
		// This is data, process it
		if err := r.processRecord(firstRecord); err != nil {
			return fmt.Errorf("failed to process first record: %w", err)
		}
	}

	// Read remaining records
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read record: %w", err)
		}

		if err := r.processRecord(record); err != nil {
			// Log error but continue processing
			fmt.Printf("Warning: failed to process record %v: %v\n", record, err)
			continue
		}
	}

	r.mu.Lock()
	r.loaded = true
	r.loadTime = time.Now()
	r.mu.Unlock()

	return nil
}

// processRecord processes a single CSV record
func (r *FileRepository) processRecord(record []string) error {
	if len(record) != 3 {
		return fmt.Errorf("invalid record format, expected 3 fields, got %d", len(record))
	}

	ip := strings.TrimSpace(record[0])
	city := strings.TrimSpace(record[1])
	country := strings.TrimSpace(record[2])

	if ip == "" || city == "" || country == "" {
		return fmt.Errorf("empty fields in record: %v", record)
	}

	// Validate IP format
	if !isValidIP(ip) {
		return fmt.Errorf("invalid IP address: %s", ip)
	}

	location := &models.Location{
		Country: country,
		City:    city,
	}

	if err := location.ValidateLocation(); err != nil {
		return fmt.Errorf("invalid location data: %w", err)
	}

	r.mu.Lock()
	r.data[ip] = location
	r.mu.Unlock()

	return nil
}

// FindLocation finds the location for a given IP address
func (r *FileRepository) FindLocation(ctx context.Context, ip string) (*models.Location, error) {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RecordLookupTime(time.Since(start).Seconds())
		}
	}()

	r.mu.RLock()
	loaded := r.loaded
	r.mu.RUnlock()

	if !loaded {
		return nil, fmt.Errorf("repository not initialized")
	}

	// Normalize IP for lookup
	normalizedIP := normalizeIP(ip)

	r.mu.RLock()
	location, exists := r.data[normalizedIP]
	r.mu.RUnlock()

	if !exists {
		if r.metrics != nil {
			r.metrics.RecordLookupCount(false)
		}
		return nil, fmt.Errorf("location not found for IP: %s", ip)
	}

	if r.metrics != nil {
		r.metrics.RecordLookupCount(true)
	}

	return location, nil
}

// Close cleans up resources
func (r *FileRepository) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data = nil
	r.loaded = false
	return nil
}

// HealthCheck checks if the repository is healthy
func (r *FileRepository) HealthCheck(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.loaded {
		return fmt.Errorf("repository not loaded")
	}

	// Check if data file still exists and is readable
	if _, err := os.Stat(r.config.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("data file does not exist: %s", r.config.FilePath)
	}

	return nil
}

// Helper functions

func isValidIP(ip string) bool {
	// Use Go's built-in IP parsing for proper validation
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil
}

func normalizeIP(ip string) string {
	// Simple normalization - just trim whitespace
	// In a real implementation, you might want to handle IPv6 normalization
	return strings.TrimSpace(ip)
}
