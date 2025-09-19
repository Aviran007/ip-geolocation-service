package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLoggingMiddleware(t *testing.T) {
	// Create a test logger that captures output
	var logOutput strings.Builder
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create middleware
	middleware := LoggingMiddleware(logger)

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Wrap with middleware
	wrappedHandler := middleware(handler)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("User-Agent", "test-agent")

	// Create response recorder
	w := httptest.NewRecorder()

	// Execute request
	wrappedHandler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Check log output
	logStr := logOutput.String()
	if !strings.Contains(logStr, "Request completed") {
		t.Error("Expected log to contain 'Request completed'")
	}
	if !strings.Contains(logStr, "method=GET") {
		t.Error("Expected log to contain method")
	}
	if !strings.Contains(logStr, "path=/test") {
		t.Error("Expected log to contain path")
	}
	if !strings.Contains(logStr, "status=200") {
		t.Error("Expected log to contain status")
	}
	if !strings.Contains(logStr, "client_ip=192.168.1.1") {
		t.Error("Expected log to contain client IP")
	}
	if !strings.Contains(logStr, "user_agent=test-agent") {
		t.Error("Expected log to contain user agent")
	}
}

func TestLoggingMiddleware_WithRealIP(t *testing.T) {
	var logOutput strings.Builder
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	middleware := LoggingMiddleware(logger)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("POST", "/api/test", nil)
	req.RemoteAddr = "10.0.0.1:8080"
	req.Header.Set("X-Real-IP", "203.0.113.1")
	req.Header.Set("User-Agent", "curl/7.68.0")

	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	// Check that X-Real-IP is used instead of RemoteAddr
	logStr := logOutput.String()
	if !strings.Contains(logStr, "client_ip=203.0.113.1") {
		t.Error("Expected log to use X-Real-IP for client IP")
	}
	if !strings.Contains(logStr, "status=404") {
		t.Error("Expected log to contain 404 status")
	}
}

func TestLoggingMiddleware_WithForwardedFor(t *testing.T) {
	var logOutput strings.Builder
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	middleware := LoggingMiddleware(logger)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("PUT", "/api/data", nil)
	req.RemoteAddr = "10.0.0.1:8080"
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 70.41.3.18, 150.172.238.178")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	// Check that first IP from X-Forwarded-For is used
	logStr := logOutput.String()
	if !strings.Contains(logStr, "client_ip=203.0.113.1") {
		t.Error("Expected log to use first IP from X-Forwarded-For")
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name         string
		remoteAddr   string
		realIP       string
		forwardedFor string
		expectedIP   string
		description  string
	}{
		{
			name:         "Real IP header",
			remoteAddr:   "192.168.1.1:12345",
			realIP:       "203.0.113.1",
			forwardedFor: "70.41.3.18",
			expectedIP:   "203.0.113.1",
			description:  "Should use X-Real-IP when present",
		},
		{
			name:         "Forwarded For header",
			remoteAddr:   "192.168.1.1:12345",
			realIP:       "",
			forwardedFor: "203.0.113.1, 70.41.3.18",
			expectedIP:   "203.0.113.1",
			description:  "Should use first IP from X-Forwarded-For",
		},
		{
			name:         "Single Forwarded For",
			remoteAddr:   "192.168.1.1:12345",
			realIP:       "",
			forwardedFor: "203.0.113.1",
			expectedIP:   "203.0.113.1",
			description:  "Should use single IP from X-Forwarded-For",
		},
		{
			name:         "Remote Addr only",
			remoteAddr:   "192.168.1.1:12345",
			realIP:       "",
			forwardedFor: "",
			expectedIP:   "192.168.1.1",
			description:  "Should fall back to RemoteAddr",
		},
		{
			name:         "Empty Remote Addr",
			remoteAddr:   "",
			realIP:       "",
			forwardedFor: "",
			expectedIP:   "",
			description:  "Should handle empty RemoteAddr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}
			if tt.forwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.forwardedFor)
			}

			result := getClientIP(req)
			if result != tt.expectedIP {
				t.Errorf("getClientIP() = %v, want %v. %s", result, tt.expectedIP, tt.description)
			}
		})
	}
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	// Create a mock ResponseWriter
	recorder := httptest.NewRecorder()
	wrapped := &responseWriter{ResponseWriter: recorder, statusCode: http.StatusOK}

	// Test WriteHeader
	wrapped.WriteHeader(http.StatusNotFound)

	// Check that status code was captured
	if wrapped.statusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, wrapped.statusCode)
	}

	// Check that underlying ResponseWriter was called
	if recorder.Code != http.StatusNotFound {
		t.Errorf("Expected underlying ResponseWriter status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	var logOutput strings.Builder
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	middleware := RecoveryMiddleware(logger)

	// Create a handler that panics
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// This should not panic
	wrappedHandler.ServeHTTP(w, req)

	// Check that we get a 500 response
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	// Check that panic was logged
	logStr := logOutput.String()
	if !strings.Contains(logStr, "panic") {
		t.Error("Expected panic to be logged")
	}
}

func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	var logOutput strings.Builder
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	middleware := RecoveryMiddleware(logger)

	// Create a normal handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	// Check normal response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Check that no panic was logged
	logStr := logOutput.String()
	if strings.Contains(logStr, "panic") {
		t.Error("Expected no panic to be logged")
	}
}

func TestCORSMiddleware(t *testing.T) {
	middleware := CORSMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)

	// Test preflight request
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")

	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	// Check CORS headers
	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type, Authorization, X-Requested-With",
		"Access-Control-Max-Age":       "3600",
	}

	for header, expectedValue := range expectedHeaders {
		actualValue := w.Header().Get(header)
		if actualValue != expectedValue {
			t.Errorf("Expected header %s: %s, got %s", header, expectedValue, actualValue)
		}
	}

	// Test normal request
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")

	w = httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	// Check that CORS headers are still present
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected CORS headers on normal request")
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	middleware := SecurityHeadersMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	// Check security headers
	expectedHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}

	for header, expectedValue := range expectedHeaders {
		actualValue := w.Header().Get(header)
		if actualValue != expectedValue {
			t.Errorf("Expected header %s: %s, got %s", header, expectedValue, actualValue)
		}
	}
}

func TestLoggingMiddleware_Duration(t *testing.T) {
	var logOutput strings.Builder
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	middleware := LoggingMiddleware(logger)

	// Create a handler that takes some time
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	// Check that duration is logged
	logStr := logOutput.String()
	if !strings.Contains(logStr, "duration=") {
		t.Error("Expected duration to be logged")
	}
}

func TestLoggingMiddleware_ErrorStatus(t *testing.T) {
	var logOutput strings.Builder
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	middleware := LoggingMiddleware(logger)

	// Create a handler that returns an error
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	// Check that error status is logged
	logStr := logOutput.String()
	if !strings.Contains(logStr, "status=400") {
		t.Error("Expected error status to be logged")
	}
}
