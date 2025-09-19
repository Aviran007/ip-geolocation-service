package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ip-geolocation-service/internal/models"
)

func TestNewIPHandler(t *testing.T) {
	service := NewMockIPService()
	logger := slog.Default()

	handler := NewIPHandler(service, logger)

	if handler == nil {
		t.Fatal("NewIPHandler() returned nil")
	}
	if handler.service == nil {
		t.Error("NewIPHandler() service not set correctly")
	}
	if handler.logger == nil {
		t.Error("NewIPHandler() logger not set correctly")
	}
}

func TestIPHandler_FindCountry_Success(t *testing.T) {
	service := NewMockIPService()
	logger := slog.Default()
	handler := NewIPHandler(service, logger)

	// Set up test data
	expectedLocation := &models.Location{
		Country: "United States",
		City:    "Mountain View",
	}
	service.SetLocation("8.8.8.8", expectedLocation)

	// Create request
	req := httptest.NewRequest("GET", "/v1/find-country?ip=8.8.8.8", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.FindCountry(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("FindCountry() status = %v, want %v", w.Code, http.StatusOK)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("FindCountry() Content-Type = %v, want application/json", contentType)
	}

	// Check response body
	body := w.Body.String()
	if !strings.Contains(body, "United States") {
		t.Errorf("FindCountry() body = %v, want to contain United States", body)
	}
	if !strings.Contains(body, "Mountain View") {
		t.Errorf("FindCountry() body = %v, want to contain Mountain View", body)
	}
}

func TestIPHandler_FindCountry_MissingIP(t *testing.T) {
	service := NewMockIPService()
	logger := slog.Default()
	handler := NewIPHandler(service, logger)

	// Create request without IP parameter
	req := httptest.NewRequest("GET", "/v1/find-country", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.FindCountry(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("FindCountry() status = %v, want %v", w.Code, http.StatusBadRequest)
	}

	// Check error message
	body := w.Body.String()
	if !strings.Contains(body, "Missing required parameter: ip") {
		t.Errorf("FindCountry() body = %v, want to contain missing parameter error", body)
	}
}

func TestIPHandler_FindCountry_InvalidMethod(t *testing.T) {
	service := NewMockIPService()
	logger := slog.Default()
	handler := NewIPHandler(service, logger)

	// Create POST request (should be GET)
	req := httptest.NewRequest("POST", "/v1/find-country?ip=8.8.8.8", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.FindCountry(w, req)

	// Check response
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("FindCountry() status = %v, want %v", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestIPHandler_FindCountry_InvalidIP(t *testing.T) {
	service := NewMockIPService()
	logger := slog.Default()
	handler := NewIPHandler(service, logger)

	// Set up service to return invalid IP error
	service.SetError("invalid-ip", errors.New("invalid IP address: invalid IP address format: invalid-ip"))

	// Create request with invalid IP
	req := httptest.NewRequest("GET", "/v1/find-country?ip=invalid-ip", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.FindCountry(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("FindCountry() status = %v, want %v", w.Code, http.StatusBadRequest)
	}

	// Check error message
	body := w.Body.String()
	if !strings.Contains(body, "Invalid IP address format") {
		t.Errorf("FindCountry() body = %v, want to contain invalid IP format error", body)
	}
}

func TestIPHandler_FindCountry_LocationNotFound(t *testing.T) {
	service := NewMockIPService()
	logger := slog.Default()
	handler := NewIPHandler(service, logger)

	// Set up service to return location not found error
	service.SetError("1.1.1.1", errors.New("location not found for IP: 1.1.1.1"))

	// Create request
	req := httptest.NewRequest("GET", "/v1/find-country?ip=1.1.1.1", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.FindCountry(w, req)

	// Check response
	if w.Code != http.StatusNotFound {
		t.Errorf("FindCountry() status = %v, want %v", w.Code, http.StatusNotFound)
	}

	// Check error message
	body := w.Body.String()
	if !strings.Contains(body, "Location not found for the provided IP address") {
		t.Errorf("FindCountry() body = %v, want to contain location not found error", body)
	}
}

func TestIPHandler_FindCountry_InternalError(t *testing.T) {
	service := NewMockIPService()
	logger := slog.Default()
	handler := NewIPHandler(service, logger)

	// Set up service to return internal error
	service.SetError("8.8.8.8", errors.New("database connection failed"))

	// Create request
	req := httptest.NewRequest("GET", "/v1/find-country?ip=8.8.8.8", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.FindCountry(w, req)

	// Check response
	if w.Code != http.StatusInternalServerError {
		t.Errorf("FindCountry() status = %v, want %v", w.Code, http.StatusInternalServerError)
	}

	// Check error message
	body := w.Body.String()
	if !strings.Contains(body, "Internal server error") {
		t.Errorf("FindCountry() body = %v, want to contain internal server error", body)
	}
}

func TestIPHandler_HealthCheck_Success(t *testing.T) {
	service := NewMockIPService()
	logger := slog.Default()
	handler := NewIPHandler(service, logger)

	// Create request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.HealthCheck(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("HealthCheck() status = %v, want %v", w.Code, http.StatusOK)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("HealthCheck() Content-Type = %v, want application/json", contentType)
	}

	// Check response body
	body := w.Body.String()
	if !strings.Contains(body, "healthy") {
		t.Errorf("HealthCheck() body = %v, want to contain healthy", body)
	}
}

func TestIPHandler_HealthCheck_Unhealthy(t *testing.T) {
	service := NewMockIPService()
	logger := slog.Default()
	handler := NewIPHandler(service, logger)

	// Set up service to return health check error
	service.SetHealthError(errors.New("database connection failed"))

	// Create request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.HealthCheck(w, req)

	// Check response
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("HealthCheck() status = %v, want %v", w.Code, http.StatusServiceUnavailable)
	}

	// Check response body
	body := w.Body.String()
	if !strings.Contains(body, "unhealthy") {
		t.Errorf("HealthCheck() body = %v, want to contain unhealthy", body)
	}
}

func TestIPHandler_NotFound(t *testing.T) {
	service := NewMockIPService()
	logger := slog.Default()
	handler := NewIPHandler(service, logger)

	// Create request
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.NotFound(w, req)

	// Check response
	if w.Code != http.StatusNotFound {
		t.Errorf("NotFound() status = %v, want %v", w.Code, http.StatusNotFound)
	}

	// Check error message
	body := w.Body.String()
	if !strings.Contains(body, "Not found") {
		t.Errorf("NotFound() body = %v, want to contain not found error", body)
	}
}

func TestIPHandler_MethodNotAllowed(t *testing.T) {
	service := NewMockIPService()
	logger := slog.Default()
	handler := NewIPHandler(service, logger)

	// Create request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.MethodNotAllowed(w, req)

	// Check response
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("MethodNotAllowed() status = %v, want %v", w.Code, http.StatusMethodNotAllowed)
	}

	// Check error message
	body := w.Body.String()
	if !strings.Contains(body, "Method not allowed") {
		t.Errorf("MethodNotAllowed() body = %v, want to contain method not allowed error", body)
	}
}
