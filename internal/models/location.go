package models

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strings"
)

// Location represents the geographical location of an IP address
type Location struct {
	Country string `json:"country"`
	City    string `json:"city"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// IPValidator provides IP address validation functionality
type IPValidator struct {
	ipv4Regex *regexp.Regexp
	ipv6Regex *regexp.Regexp
}

// NewIPValidator creates a new IP validator
func NewIPValidator() *IPValidator {
	return &IPValidator{
		ipv4Regex: regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`),
		ipv6Regex: regexp.MustCompile(`^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$`),
	}
}

// ValidateIP validates if the given string is a valid IP address
func (v *IPValidator) ValidateIP(ip string) error {
	if ip == "" {
		return fmt.Errorf("IP address cannot be empty")
	}

	// Check if it's a valid IPv4 or IPv6 address
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address format: %s", ip)
	}

	return nil
}

// IsIPv4 checks if the IP is IPv4
func (v *IPValidator) IsIPv4(ip string) bool {
	return v.ipv4Regex.MatchString(ip)
}

// IsIPv6 checks if the IP is IPv6
func (v *IPValidator) IsIPv6(ip string) bool {
	return v.ipv6Regex.MatchString(ip)
}

// NormalizeIP normalizes the IP address for consistent storage/lookup
func (v *IPValidator) NormalizeIP(ip string) string {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return ip
	}
	return parsedIP.String()
}

// ToJSON converts Location to JSON
func (l *Location) ToJSON() ([]byte, error) {
	return json.Marshal(l)
}

// ToJSON converts ErrorResponse to JSON
func (e *ErrorResponse) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// NewErrorResponse creates a new error response
func NewErrorResponse(message string) *ErrorResponse {
	return &ErrorResponse{Error: message}
}

// ValidateLocation validates location data
func (l *Location) ValidateLocation() error {
	if strings.TrimSpace(l.Country) == "" {
		return fmt.Errorf("country cannot be empty")
	}
	if strings.TrimSpace(l.City) == "" {
		return fmt.Errorf("city cannot be empty")
	}
	return nil
}
