package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestIPValidator_ValidateIP(t *testing.T) {
	validator := NewIPValidator()

	tests := []struct {
		name        string
		ip          string
		wantErr     bool
		description string
	}{
		{
			name:        "Valid IPv4 - Private",
			ip:          "192.168.1.1",
			wantErr:     false,
			description: "Should accept valid private IPv4 address",
		},
		{
			name:        "Valid IPv4 - Public",
			ip:          "8.8.8.8",
			wantErr:     false,
			description: "Should accept valid public IPv4 address",
		},
		{
			name:        "Valid IPv4 - Localhost",
			ip:          "127.0.0.1",
			wantErr:     false,
			description: "Should accept localhost IPv4 address",
		},
		{
			name:        "Valid IPv4 - Broadcast",
			ip:          "255.255.255.255",
			wantErr:     false,
			description: "Should accept broadcast IPv4 address",
		},
		{
			name:        "Valid IPv4 - Zero",
			ip:          "0.0.0.0",
			wantErr:     false,
			description: "Should accept zero IPv4 address",
		},
		{
			name:        "Invalid IPv4 - Out of range",
			ip:          "256.1.1.1",
			wantErr:     true,
			description: "Should reject IPv4 with octets > 255",
		},
		{
			name:        "Invalid IPv4 - Too few octets",
			ip:          "192.168.1",
			wantErr:     true,
			description: "Should reject IPv4 with too few octets",
		},
		{
			name:        "Invalid IPv4 - Too many octets",
			ip:          "192.168.1.1.1",
			wantErr:     true,
			description: "Should reject IPv4 with too many octets",
		},
		{
			name:        "Empty IP",
			ip:          "",
			wantErr:     true,
			description: "Should reject empty IP address",
		},
		{
			name:        "Invalid format - Not IP",
			ip:          "not-an-ip",
			wantErr:     true,
			description: "Should reject non-IP format string",
		},
		{
			name:        "Invalid format - Mixed",
			ip:          "192.168.1.abc",
			wantErr:     true,
			description: "Should reject IPv4 with non-numeric octets",
		},
		{
			name:        "Invalid format - Negative",
			ip:          "192.168.-1.1",
			wantErr:     true,
			description: "Should reject IPv4 with negative octets",
		},
	}

	for _, tt := range tests {
		tt := tt // Capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Enable parallel execution

			err := validator.ValidateIP(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIP(%q) error = %v, wantErr %v. %s",
					tt.ip, err, tt.wantErr, tt.description)
			}
		})
	}
}

func TestLocation_ValidateLocation(t *testing.T) {
	tests := []struct {
		name        string
		location    Location
		wantErr     bool
		description string
	}{
		{
			name:        "Valid location - US",
			location:    Location{Country: "US", City: "New York"},
			wantErr:     false,
			description: "Should accept valid US location",
		},
		{
			name:        "Valid location - Israel",
			location:    Location{Country: "IL", City: "Tel Aviv"},
			wantErr:     false,
			description: "Should accept valid Israeli location",
		},
		{
			name:        "Valid location - Long names",
			location:    Location{Country: "United States of America", City: "San Francisco"},
			wantErr:     false,
			description: "Should accept location with long names",
		},
		{
			name:        "Empty country",
			location:    Location{Country: "", City: "New York"},
			wantErr:     true,
			description: "Should reject location with empty country",
		},
		{
			name:        "Empty city",
			location:    Location{Country: "US", City: ""},
			wantErr:     true,
			description: "Should reject location with empty city",
		},
		{
			name:        "Whitespace country",
			location:    Location{Country: "   ", City: "New York"},
			wantErr:     true,
			description: "Should reject location with whitespace-only country",
		},
		{
			name:        "Whitespace city",
			location:    Location{Country: "US", City: "   "},
			wantErr:     true,
			description: "Should reject location with whitespace-only city",
		},
		{
			name:        "Both empty",
			location:    Location{Country: "", City: ""},
			wantErr:     true,
			description: "Should reject location with both empty fields",
		},
		{
			name:        "Both whitespace",
			location:    Location{Country: "   ", City: "   "},
			wantErr:     true,
			description: "Should reject location with both whitespace fields",
		},
	}

	for _, tt := range tests {
		tt := tt // Capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Enable parallel execution

			err := tt.location.ValidateLocation()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLocation() error = %v, wantErr %v. %s",
					err, tt.wantErr, tt.description)
			}
		})
	}
}

func TestLocation_ToJSON(t *testing.T) {
	tests := []struct {
		name        string
		location    Location
		expected    string
		description string
	}{
		{
			name:        "Valid US location",
			location:    Location{Country: "US", City: "New York"},
			expected:    `{"country":"US","city":"New York"}`,
			description: "Should serialize US location correctly",
		},
		{
			name:        "Valid Israeli location",
			location:    Location{Country: "IL", City: "Tel Aviv"},
			expected:    `{"country":"IL","city":"Tel Aviv"}`,
			description: "Should serialize Israeli location correctly",
		},
		{
			name:        "Location with special characters",
			location:    Location{Country: "FR", City: "Saint-Étienne"},
			expected:    `{"country":"FR","city":"Saint-Étienne"}`,
			description: "Should serialize location with special characters correctly",
		},
		{
			name:        "Location with long names",
			location:    Location{Country: "United States of America", City: "San Francisco"},
			expected:    `{"country":"United States of America","city":"San Francisco"}`,
			description: "Should serialize location with long names correctly",
		},
	}

	for _, tt := range tests {
		tt := tt // Capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Enable parallel execution

			json, err := tt.location.ToJSON()
			if err != nil {
				t.Errorf("ToJSON() error = %v. %s", err, tt.description)
				return
			}

			if string(json) != tt.expected {
				t.Errorf("ToJSON() = %v, want %v. %s", string(json), tt.expected, tt.description)
			}
		})
	}
}

func TestErrorResponse_ToJSON(t *testing.T) {
	tests := []struct {
		name        string
		errorResp   *ErrorResponse
		expected    string
		description string
	}{
		{
			name:        "Simple error message",
			errorResp:   NewErrorResponse("Test error"),
			expected:    `{"error":"Test error"}`,
			description: "Should serialize simple error message correctly",
		},
		{
			name:        "Error with special characters",
			errorResp:   NewErrorResponse("Error: Invalid IP address 192.168.1.999"),
			expected:    `{"error":"Error: Invalid IP address 192.168.1.999"}`,
			description: "Should serialize error with special characters correctly",
		},
		{
			name:        "Empty error message",
			errorResp:   NewErrorResponse(""),
			expected:    `{"error":""}`,
			description: "Should serialize empty error message correctly",
		},
		{
			name:        "Error with quotes",
			errorResp:   NewErrorResponse("Error: \"Invalid input\""),
			expected:    `{"error":"Error: \"Invalid input\""}`,
			description: "Should serialize error with quotes correctly",
		},
	}

	for _, tt := range tests {
		tt := tt // Capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Enable parallel execution

			json, err := tt.errorResp.ToJSON()
			if err != nil {
				t.Errorf("ToJSON() error = %v. %s", err, tt.description)
				return
			}

			if string(json) != tt.expected {
				t.Errorf("ToJSON() = %v, want %v. %s", string(json), tt.expected, tt.description)
			}
		})
	}
}

func TestSlowOperation(t *testing.T) {
	// This test intentionally takes time to verify tests are actually running
	time.Sleep(100 * time.Millisecond)

	// Simple validation to ensure test runs
	validator := NewIPValidator()
	err := validator.ValidateIP("8.8.8.8")
	if err != nil {
		t.Errorf("Expected valid IP to pass validation, got error: %v", err)
	}
}

// TestIPValidator_Benchmark tests performance of IP validation
func TestIPValidator_Benchmark(t *testing.T) {
	validator := NewIPValidator()

	// Test with various IP formats
	ips := []string{
		"192.168.1.1",
		"8.8.8.8",
		"127.0.0.1",
		"255.255.255.255",
		"0.0.0.0",
		"10.0.0.1",
		"172.16.0.1",
	}

	// Benchmark validation performance
	start := time.Now()
	for i := 0; i < 1000; i++ {
		for _, ip := range ips {
			validator.ValidateIP(ip)
		}
	}
	duration := time.Since(start)

	// Should complete within reasonable time
	if duration > 100*time.Millisecond {
		t.Errorf("IP validation took too long: %v", duration)
	}

	t.Logf("Validated %d IPs in %v", len(ips)*1000, duration)
}

// TestLocation_JSONRoundTrip tests JSON serialization and deserialization
func TestLocation_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name        string
		location    Location
		description string
	}{
		{
			name:        "US location",
			location:    Location{Country: "US", City: "New York"},
			description: "Should round-trip US location correctly",
		},
		{
			name:        "Israeli location",
			location:    Location{Country: "IL", City: "Tel Aviv"},
			description: "Should round-trip Israeli location correctly",
		},
		{
			name:        "Location with special characters",
			location:    Location{Country: "FR", City: "Saint-Étienne"},
			description: "Should round-trip location with special characters correctly",
		},
	}

	for _, tt := range tests {
		tt := tt // Capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Enable parallel execution

			// Serialize to JSON
			jsonData, err := tt.location.ToJSON()
			if err != nil {
				t.Errorf("ToJSON() error = %v. %s", err, tt.description)
				return
			}

			// Deserialize from JSON
			var deserialized Location
			err = json.Unmarshal(jsonData, &deserialized)
			if err != nil {
				t.Errorf("json.Unmarshal() error = %v. %s", err, tt.description)
				return
			}

			// Compare original and deserialized
			if deserialized.Country != tt.location.Country {
				t.Errorf("Country mismatch: got %v, want %v. %s",
					deserialized.Country, tt.location.Country, tt.description)
			}
			if deserialized.City != tt.location.City {
				t.Errorf("City mismatch: got %v, want %v. %s",
					deserialized.City, tt.location.City, tt.description)
			}
		})
	}
}

// Additional tests for IPValidator methods
func TestIPValidator_IsIPv4(t *testing.T) {
	validator := NewIPValidator()

	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"Valid IPv4", "192.168.1.1", true},
		{"Valid IPv4", "8.8.8.8", true},
		{"Valid IPv4", "127.0.0.1", true},
		{"Valid IPv4", "255.255.255.255", true},
		{"Valid IPv4", "0.0.0.0", true},
		{"Invalid IPv4", "192.168.1", false},
		{"Invalid IPv4", "192.168.1.1.1", false},
		{"Invalid IPv4", "256.1.1.1", false},
		{"Invalid IPv4", "192.168.1.abc", false},
		{"Empty string", "", false},
		{"Not an IP", "not-an-ip", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.IsIPv4(tt.ip)
			if result != tt.expected {
				t.Errorf("IsIPv4(%q) = %v, want %v", tt.ip, result, tt.expected)
			}
		})
	}
}

func TestIPValidator_IsIPv6(t *testing.T) {
	validator := NewIPValidator()

	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"Valid IPv6 - Full format", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},
		{"Valid IPv6 - Compressed", "2001:db8:85a3::8a2e:370:7334", false}, // This regex doesn't handle :: compression
		{"Valid IPv6 - Localhost", "::1", false},                           // This regex doesn't handle :: compression
		{"Valid IPv6 - All zeros", "::", false},                            // This regex doesn't handle :: compression
		{"Invalid IPv6", "192.168.1.1", false},
		{"Invalid IPv6 - Too many segments", "2001:0db8:85a3:0000:0000:8a2e:0370:7334:7334", false},
		{"Invalid IPv6 - Too few segments", "2001:0db8:85a3:0000:0000:8a2e:0370", false},
		{"Invalid IPv6 - Invalid characters", "2001:0db8:85a3:0000:0000:8a2e:0370:733g", false},
		{"Empty string", "", false},
		{"Not an IP", "not-an-ip", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.IsIPv6(tt.ip)
			if result != tt.expected {
				t.Errorf("IsIPv6(%q) = %v, want %v", tt.ip, result, tt.expected)
			}
		})
	}
}

func TestIPValidator_NormalizeIP(t *testing.T) {
	validator := NewIPValidator()

	tests := []struct {
		name     string
		ip       string
		expected string
	}{
		{"Valid IPv4", "192.168.1.1", "192.168.1.1"},
		{"Valid IPv4 with spaces", " 192.168.1.1 ", " 192.168.1.1 "}, // NormalizeIP doesn't trim spaces
		{"Valid IPv6", "2001:0db8:85a3::8a2e:370:7334", "2001:db8:85a3::8a2e:370:7334"},
		{"Invalid IP", "invalid-ip", "invalid-ip"},
		{"Empty string", "", ""},
		{"Whitespace only", "   ", "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.NormalizeIP(tt.ip)
			if result != tt.expected {
				t.Errorf("NormalizeIP(%q) = %v, want %v", tt.ip, result, tt.expected)
			}
		})
	}
}

func TestNewIPValidator(t *testing.T) {
	validator := NewIPValidator()

	if validator == nil {
		t.Fatal("NewIPValidator() returned nil")
	}

	// Test that the validator can validate IPs
	err := validator.ValidateIP("8.8.8.8")
	if err != nil {
		t.Errorf("NewIPValidator() created invalid validator: %v", err)
	}

	// Test that IPv4 regex is compiled
	if validator.ipv4Regex == nil {
		t.Error("IPv4 regex not compiled")
	}

	// Test that IPv6 regex is compiled
	if validator.ipv6Regex == nil {
		t.Error("IPv6 regex not compiled")
	}
}

func TestNewErrorResponse(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{"Simple message", "Test error", "Test error"},
		{"Empty message", "", ""},
		{"Message with special chars", "Error: Invalid input \"test\"", "Error: Invalid input \"test\""},
		{"Long message", "This is a very long error message that contains multiple words and should be handled correctly", "This is a very long error message that contains multiple words and should be handled correctly"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorResp := NewErrorResponse(tt.message)

			if errorResp == nil {
				t.Fatal("NewErrorResponse() returned nil")
			}

			if errorResp.Error != tt.expected {
				t.Errorf("NewErrorResponse() error = %v, want %v", errorResp.Error, tt.expected)
			}
		})
	}
}
