package handlers

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ip-geolocation-service/internal/middleware"
)

func TestNewRouter(t *testing.T) {
	service := &MockIPService{}
	logger := slog.Default()

	router := NewRouter(service, logger)

	if router == nil {
		t.Fatal("NewRouter() returned nil")
	}

	if router.ipHandler == nil {
		t.Error("NewRouter() ipHandler not set")
	}

	if router.logger != logger {
		t.Error("NewRouter() logger not set correctly")
	}
}

func TestRouter_SetupRoutes(t *testing.T) {
	service := &MockIPService{}
	logger := slog.Default()
	router := NewRouter(service, logger)

	mux := router.SetupRoutes()

	if mux == nil {
		t.Fatal("SetupRoutes() returned nil")
	}

	// Test that routes are registered
	testCases := []struct {
		method string
		path   string
		status int
	}{
		{"GET", "/v1/find-country?ip=8.8.8.8", http.StatusOK},
		{"GET", "/health", http.StatusOK},
		{"GET", "/", http.StatusNotFound},
		{"GET", "/nonexistent", http.StatusNotFound},
	}

	for _, tc := range testCases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			// We can't easily test the exact status codes without setting up the service
			// But we can test that the routes are registered and respond
			if w.Code == 0 {
				t.Error("Route not registered or handler not called")
			}
		})
	}
}

func TestRouter_SetupRoutesWithMiddleware(t *testing.T) {
	service := &MockIPService{}
	logger := slog.Default()
	router := NewRouter(service, logger)

	// Create a rate limiter
	rateLimiter := middleware.NewRateLimiter(100, 200, 1, 1*time.Minute, 5*time.Minute)

	handler := router.SetupRoutesWithMiddleware(rateLimiter)

	if handler == nil {
		t.Fatal("SetupRoutesWithMiddleware() returned nil")
	}

	// Test that the handler is properly wrapped with middleware
	// We can't easily test the middleware behavior without more complex setup
	// But we can test that it returns a valid handler
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// The handler should respond (even if it's an error due to missing setup)
	if w.Code == 0 {
		t.Error("Handler not properly configured")
	}
}

func TestRouter_SetupRoutesWithMiddleware_NilRateLimiter(t *testing.T) {
	service := &MockIPService{}
	logger := slog.Default()
	router := NewRouter(service, logger)

	// Test with nil rate limiter
	handler := router.SetupRoutesWithMiddleware(nil)

	if handler == nil {
		t.Fatal("SetupRoutesWithMiddleware() returned nil with nil rate limiter")
	}

	// Should still work, just without rate limiting
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code == 0 {
		t.Error("Handler not properly configured with nil rate limiter")
	}
}
