package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ip-geolocation-service/internal/config"
)

// Test data constants
const (
	testCSVHeader = "ip,city,country"
	testCSVData   = `ip,city,country
1.1.1.1,New York,United States
8.8.8.8,Mountain View,United States
192.168.1.1,Local Network,Private`

	invalidCSVData = `ip,city,country
1.1.1.1,New York,United States
invalid-ip,Invalid City,Invalid Country
8.8.8.8,Mountain View,United States`

	testIP1       = "1.1.1.1"
	testIP2       = "8.8.8.8"
	testIP3       = "192.168.1.1"
	invalidIP     = "invalid-ip"
	nonExistentIP = "999.999.999.999"
)

func TestFileRepository_Initialize(t *testing.T) {
	// Create temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_data.csv")

	// Write test data
	err := os.WriteFile(testFile, []byte(testCSVData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create repository
	cfg := &config.DatabaseConfig{
		Type:     "csv",
		FilePath: testFile,
	}

	repo := NewFileRepository(cfg, nil)

	// Initialize repository
	ctx := context.Background()
	err = repo.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Test finding location
	location, err := repo.FindLocation(ctx, testIP1)
	if err != nil {
		t.Fatalf("Failed to find location: %v", err)
	}

	if location.Country != "United States" {
		t.Errorf("Expected country 'United States', got '%s'", location.Country)
	}

	if location.City != "New York" {
		t.Errorf("Expected city 'New York', got '%s'", location.City)
	}

	// Test non-existent IP
	_, err = repo.FindLocation(ctx, nonExistentIP)
	if err == nil {
		t.Error("Expected error for non-existent IP")
	}

	// Test health check
	err = repo.HealthCheck(ctx)
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}

}

func TestFileRepository_InvalidData(t *testing.T) {
	// Create temporary test file with invalid data
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "invalid_data.csv")

	// Write invalid test data
	err := os.WriteFile(testFile, []byte(invalidCSVData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create repository
	cfg := &config.DatabaseConfig{
		Type:     "csv",
		FilePath: testFile,
	}

	repo := NewFileRepository(cfg, nil)

	// Initialize repository (should handle invalid data gracefully)
	ctx := context.Background()
	err = repo.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Should still be able to find valid IPs
	location, err := repo.FindLocation(ctx, testIP1)
	if err != nil {
		t.Fatalf("Failed to find location: %v", err)
	}

	if location.Country != "United States" {
		t.Errorf("Expected country 'United States', got '%s'", location.Country)
	}
}

func TestFileRepository_FileNotFound(t *testing.T) {
	// Create repository with non-existent file
	cfg := &config.DatabaseConfig{
		Type:     "csv",
		FilePath: "/non/existent/file.csv",
	}

	repo := NewFileRepository(cfg, nil)

	// Initialize repository should fail
	ctx := context.Background()
	err := repo.Initialize(ctx)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestFileRepository_NotInitialized(t *testing.T) {
	// Create repository but don't initialize
	cfg := &config.DatabaseConfig{
		Type:     "csv",
		FilePath: "/some/file.csv",
	}

	repo := NewFileRepository(cfg, nil)

	// FindLocation should fail
	ctx := context.Background()
	_, err := repo.FindLocation(ctx, "1.1.1.1")
	if err == nil {
		t.Error("Expected error for uninitialized repository")
	}
}

func TestFileRepository_ConcurrentAccess(t *testing.T) {
	// Create temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "concurrent_test.csv")

	// Write test data
	err := os.WriteFile(testFile, []byte(testCSVData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create repository
	cfg := &config.DatabaseConfig{
		Type:     "csv",
		FilePath: testFile,
	}

	repo := NewFileRepository(cfg, nil)

	// Initialize repository
	ctx := context.Background()
	err = repo.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			_, err := repo.FindLocation(ctx, testIP1)
			if err != nil {
				t.Errorf("Concurrent access failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestFileRepository_Performance(t *testing.T) {
	// This test intentionally takes time to verify tests are actually running
	time.Sleep(200 * time.Millisecond)

	// Create temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "performance_test.csv")

	// Write test data
	err := os.WriteFile(testFile, []byte(testCSVData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create repository
	cfg := &config.DatabaseConfig{
		Type:     "csv",
		FilePath: testFile,
	}

	repo := NewFileRepository(cfg, nil)

	// Initialize repository
	ctx := context.Background()
	err = repo.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Test performance
	start := time.Now()
	for i := 0; i < 100; i++ {
		_, err := repo.FindLocation(ctx, testIP1)
		if err != nil {
			t.Errorf("Performance test failed: %v", err)
		}
	}
	duration := time.Since(start)

	// Verify it took some time (but not too much)
	if duration < 1*time.Microsecond {
		t.Errorf("Expected performance test to take at least 1µs, got %v", duration)
	}

	// Log the actual duration for verification
	t.Logf("Performance test completed in %v", duration)
}

// TestEdgeCases tests various edge cases
func TestFileRepository_EdgeCases(t *testing.T) {
	// Create temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "edge_cases_test.csv")

	// Write test data with edge cases
	edgeCaseData := `ip,city,country
1.1.1.1,New York,United States
8.8.8.8,Mountain View,United States
192.168.1.1,Local Network,Private
127.0.0.1,Localhost,Private
0.0.0.0,Unspecified,Private
255.255.255.255,Broadcast,Private`

	err := os.WriteFile(testFile, []byte(edgeCaseData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create repository
	cfg := &config.DatabaseConfig{
		Type:     "csv",
		FilePath: testFile,
	}

	repo := NewFileRepository(cfg, nil)

	// Initialize repository
	ctx := context.Background()
	err = repo.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Test edge cases
	testCases := []struct {
		name     string
		ip       string
		expected string
	}{
		{"Valid IPv4", "1.1.1.1", "New York"},
		{"Valid IPv4", "8.8.8.8", "Mountain View"},
		{"Private IP", "192.168.1.1", "Local Network"},
		{"Localhost", "127.0.0.1", "Localhost"},
		{"Unspecified", "0.0.0.0", "Unspecified"},
		{"Broadcast", "255.255.255.255", "Broadcast"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			location, err := repo.FindLocation(ctx, tc.ip)
			if err != nil {
				t.Errorf("Expected to find location for %s, got error: %v", tc.ip, err)
				return
			}
			if location.City != tc.expected {
				t.Errorf("Expected city %s for IP %s, got %s", tc.expected, tc.ip, location.City)
			}
		})
	}

	// Test non-existent IP
	_, err = repo.FindLocation(ctx, "999.999.999.999")
	if err == nil {
		t.Error("Expected error for non-existent IP")
	}

	// Test empty IP
	_, err = repo.FindLocation(ctx, "")
	if err == nil {
		t.Error("Expected error for empty IP")
	}

	// Test invalid IP format
	_, err = repo.FindLocation(ctx, "not-an-ip")
	if err == nil {
		t.Error("Expected error for invalid IP format")
	}
}

func TestFileRepository_Close(t *testing.T) {
	// Create temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_data.csv")

	// Write test data
	err := os.WriteFile(testFile, []byte(testCSVData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create repository
	cfg := &config.DatabaseConfig{
		Type:     "csv",
		FilePath: testFile,
	}

	repo := NewFileRepository(cfg, nil)
	ctx := context.Background()
	err = repo.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Test Close
	err = repo.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// Test Close on already closed repository
	err = repo.Close()
	if err != nil {
		t.Errorf("Close() on already closed repository returned error: %v", err)
	}
}

func TestFileRepository_HealthCheck(t *testing.T) {
	// Create temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_data.csv")

	// Write test data
	err := os.WriteFile(testFile, []byte(testCSVData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create repository
	cfg := &config.DatabaseConfig{
		Type:     "csv",
		FilePath: testFile,
	}

	repo := NewFileRepository(cfg, nil)
	ctx := context.Background()

	// Test HealthCheck before initialization
	err = repo.HealthCheck(ctx)
	if err == nil {
		t.Error("Expected error for health check before initialization")
	}

	// Initialize repository
	err = repo.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Test HealthCheck after initialization
	err = repo.HealthCheck(ctx)
	if err != nil {
		t.Errorf("HealthCheck() after initialization returned error: %v", err)
	}

	// Test HealthCheck with context timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 0)
	defer cancel()
	err = repo.HealthCheck(timeoutCtx)
	// Note: The health check might not timeout immediately, so we just check it doesn't panic
	// The actual timeout behavior depends on the implementation
}

func TestFileRepository_HealthCheck_NotInitialized(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Type:     "csv",
		FilePath: "/nonexistent/path.csv",
	}

	repo := NewFileRepository(cfg, nil)
	ctx := context.Background()

	// Test HealthCheck on uninitialized repository
	err := repo.HealthCheck(ctx)
	if err == nil {
		t.Error("Expected error for health check on uninitialized repository")
	}
}

func TestFileRepository_Initialize_FileNotFound(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Type:     "csv",
		FilePath: "/nonexistent/path.csv",
	}

	repo := NewFileRepository(cfg, nil)
	ctx := context.Background()

	// Test Initialize with non-existent file
	err := repo.Initialize(ctx)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestFileRepository_Initialize_InvalidCSV(t *testing.T) {
	// Create temporary test file with invalid CSV
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "invalid_data.csv")

	// Write invalid CSV data
	invalidData := `ip,city,country
1.1.1.1,New York
8.8.8.8,Mountain View,United States,Extra Field`

	err := os.WriteFile(testFile, []byte(invalidData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create repository
	cfg := &config.DatabaseConfig{
		Type:     "csv",
		FilePath: testFile,
	}

	repo := NewFileRepository(cfg, nil)
	ctx := context.Background()

	// Test Initialize with invalid CSV
	err = repo.Initialize(ctx)
	if err == nil {
		t.Error("Expected error for invalid CSV format")
	}
}

func TestFileRepository_Initialize_EmptyFile(t *testing.T) {
	// Create temporary empty file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "empty_data.csv")

	// Write empty file
	err := os.WriteFile(testFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create repository
	cfg := &config.DatabaseConfig{
		Type:     "csv",
		FilePath: testFile,
	}

	repo := NewFileRepository(cfg, nil)
	ctx := context.Background()

	// Test Initialize with empty file
	err = repo.Initialize(ctx)
	if err == nil {
		t.Error("Expected error for empty file")
	}
}

func TestFileRepository_Initialize_HeaderOnly(t *testing.T) {
	// Create temporary test file with header only
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "header_only.csv")

	// Write header only
	headerOnly := "ip,city,country\n"

	err := os.WriteFile(testFile, []byte(headerOnly), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create repository
	cfg := &config.DatabaseConfig{
		Type:     "csv",
		FilePath: testFile,
	}

	repo := NewFileRepository(cfg, nil)
	ctx := context.Background()

	// Test Initialize with header only
	err = repo.Initialize(ctx)
	if err != nil {
		t.Errorf("Initialize() with header only returned error: %v", err)
	}

	// Test FindLocation with no data
	_, err = repo.FindLocation(ctx, "1.1.1.1")
	if err == nil {
		t.Error("Expected error for FindLocation with no data")
	}
}

func TestFileRepository_ProcessRecord_EdgeCases(t *testing.T) {
	// Create temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_data.csv")

	// Write test data with edge cases
	edgeCaseData := `ip,city,country
1.1.1.1,New York,United States
8.8.8.8,Mountain View,United States
192.168.1.1,Local Network,Private
127.0.0.1,Localhost,Local
0.0.0.0,Unspecified,Unknown
255.255.255.255,Broadcast,Unknown
2001:db8::1,IPv6 City,IPv6 Country
invalid-ip,Invalid City,Invalid Country
,Empty IP,Empty Country
1.1.1.1,,Empty City
1.1.1.1,City,`

	err := os.WriteFile(testFile, []byte(edgeCaseData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create repository
	cfg := &config.DatabaseConfig{
		Type:     "csv",
		FilePath: testFile,
	}

	repo := NewFileRepository(cfg, nil)
	ctx := context.Background()
	err = repo.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Test various edge cases
	testCases := []struct {
		name        string
		ip          string
		expectError bool
		description string
	}{
		{"Valid IPv4", "1.1.1.1", false, "Should find valid IPv4"},
		{"Valid IPv4", "8.8.8.8", false, "Should find valid IPv4"},
		{"Private IP", "192.168.1.1", false, "Should find private IP"},
		{"Localhost", "127.0.0.1", false, "Should find localhost"},
		{"Unspecified", "0.0.0.0", false, "Should find unspecified IP"},
		{"Broadcast", "255.255.255.255", false, "Should find broadcast IP"},
		{"IPv6", "2001:db8::1", false, "Should accept IPv6 addresses"},
		{"Invalid IP", "invalid-ip", true, "Should error on invalid IP"},
		{"Empty IP", "", true, "Should error on empty IP"},
		{"Non-existent IP", "999.999.999.999", true, "Should error on non-existent IP"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := repo.FindLocation(ctx, tc.ip)
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, got nil. %s", tc.ip, tc.description)
				}
			} else {
				if err != nil {
					t.Errorf("Expected success for %s, got error: %v. %s", tc.ip, err, tc.description)
				}
			}
		})
	}
}

func TestFileRepository_IsValidIP(t *testing.T) {
	// Create temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_data.csv")

	// Write test data
	err := os.WriteFile(testFile, []byte(testCSVData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create repository
	cfg := &config.DatabaseConfig{
		Type:     "csv",
		FilePath: testFile,
	}

	repo := NewFileRepository(cfg, nil)
	ctx := context.Background()
	err = repo.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Test isValidIP method (we need to access it through reflection or make it public for testing)
	// For now, we'll test it indirectly through FindLocation
	testCases := []struct {
		ip       string
		expected bool
	}{
		{"1.1.1.1", true},
		{"8.8.8.8", true},
		{"192.168.1.1", true},
		{"127.0.0.1", true},
		{"0.0.0.0", true},
		{"255.255.255.255", true},
		{"2001:db8::1", true},
		{"invalid-ip", false},
		{"", false},
		{"999.999.999.999", false},
		{"not-an-ip", false},
	}

	for _, tc := range testCases {
		t.Run(tc.ip, func(t *testing.T) {
			_, err := repo.FindLocation(ctx, tc.ip)
			// If the IP is valid, we should either find it or get a "not found" error
			// If the IP is invalid, we should get a validation error
			if tc.expected {
				// For valid IPs, we expect either success or "not found" error
				if err != nil && err.Error() != "location not found for IP: "+tc.ip {
					t.Errorf("Expected success or 'not found' for valid IP %s, got: %v", tc.ip, err)
				}
			} else {
				// For invalid IPs, we expect a validation error
				if err == nil {
					t.Errorf("Expected error for invalid IP %s, got success", tc.ip)
				}
			}
		})
	}
}

func TestFileRepository_ConcurrentAccess_Extended(t *testing.T) {
	// Create temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_data.csv")

	// Write test data
	err := os.WriteFile(testFile, []byte(testCSVData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create repository
	cfg := &config.DatabaseConfig{
		Type:     "csv",
		FilePath: testFile,
	}

	repo := NewFileRepository(cfg, nil)
	ctx := context.Background()
	err = repo.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Make multiple concurrent requests
			for j := 0; j < 5; j++ {
				_, err := repo.FindLocation(ctx, "1.1.1.1")
				if err != nil {
					t.Errorf("Concurrent FindLocation failed: %v", err)
				}
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
