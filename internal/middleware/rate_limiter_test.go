package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	tests := []struct {
		name          string
		ratePerSecond int
		burstSize     int
		windowSize    time.Duration
		clientID      string
		requests      int
		expectedAllow []bool
		description   string
	}{
		{
			name:          "Basic rate limiting",
			ratePerSecond: 1,
			burstSize:     2,
			windowSize:    time.Second,
			clientID:      "test-client-1",
			requests:      3,
			expectedAllow: []bool{true, true, false},
			description:   "Should allow burst requests then deny",
		},
		{
			name:          "High rate limiting",
			ratePerSecond: 10,
			burstSize:     5,
			windowSize:    time.Second,
			clientID:      "test-client-2",
			requests:      7,
			expectedAllow: []bool{true, true, true, true, true, false, false},
			description:   "Should allow high burst requests then deny",
		},
		{
			name:          "Low rate limiting",
			ratePerSecond: 2,
			burstSize:     1,
			windowSize:    time.Second,
			clientID:      "test-client-3",
			requests:      3,
			expectedAllow: []bool{true, false, false},
			description:   "Should allow single burst request then deny",
		},
		{
			name:          "Zero rate limiting",
			ratePerSecond: 0,
			burstSize:     1,
			windowSize:    time.Second,
			clientID:      "test-client-4",
			requests:      2,
			expectedAllow: []bool{false, false},
			description:   "Should deny all requests with zero rate",
		},
	}

	for _, tt := range tests {
		tt := tt // Capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Enable parallel execution

			rateLimiter := NewRateLimiter(tt.ratePerSecond, tt.burstSize, tt.windowSize, 1*time.Minute, 5*time.Minute)

			for i := 0; i < tt.requests; i++ {
				allowed := rateLimiter.Allow(tt.clientID)
				expected := tt.expectedAllow[i]

				if allowed != expected {
					t.Errorf("Request %d: Allow() = %v, want %v. %s",
						i+1, allowed, expected, tt.description)
				}
			}
		})
	}
}

func TestRateLimiter_GetClientID(t *testing.T) {
	rateLimiter := NewRateLimiter(1, 1, time.Second, 1*time.Minute, 5*time.Minute)

	tests := []struct {
		name           string
		headers        map[string]string
		remoteAddr     string
		expectedClient string
		description    string
	}{
		{
			name:           "X-Real-IP header",
			headers:        map[string]string{"X-Real-IP": "192.168.1.1"},
			remoteAddr:     "10.0.0.1:1234",
			expectedClient: "192.168.1.1",
			description:    "Should prioritize X-Real-IP header over RemoteAddr",
		},
		{
			name:           "X-Forwarded-For header",
			headers:        map[string]string{"X-Forwarded-For": "203.0.113.1, 70.41.3.18, 150.172.238.178"},
			remoteAddr:     "10.0.0.1:1234",
			expectedClient: "203.0.113.1",
			description:    "Should use first IP from comma-separated X-Forwarded-For",
		},
		{
			name:           "Remote address fallback",
			headers:        map[string]string{},
			remoteAddr:     "192.168.1.100:1234",
			expectedClient: "192.168.1.100",
			description:    "Should use RemoteAddr when no headers are present (port removed)",
		},
		{
			name:           "Empty headers",
			headers:        map[string]string{"X-Real-IP": "", "X-Forwarded-For": ""},
			remoteAddr:     "127.0.0.1:8080",
			expectedClient: "127.0.0.1",
			description:    "Should fallback to RemoteAddr when headers are empty (port removed)",
		},
		{
			name:           "Multiple X-Forwarded-For values",
			headers:        map[string]string{"X-Forwarded-For": "10.0.0.1, 192.168.1.1, 172.16.0.1"},
			remoteAddr:     "127.0.0.1:8080",
			expectedClient: "10.0.0.1",
			description:    "Should use first IP from comma-separated X-Forwarded-For",
		},
	}

	for _, tt := range tests {
		tt := tt // Capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Enable parallel execution

			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			clientID := rateLimiter.GetClientID(req)
			if clientID != tt.expectedClient {
				t.Errorf("GetClientID() = %v, want %v. %s",
					clientID, tt.expectedClient, tt.description)
			}
		})
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	rateLimiter := NewRateLimiter(1, 1, time.Second, 1*time.Minute, 5*time.Minute)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	middleware := RateLimitMiddleware(rateLimiter)
	wrappedHandler := middleware(handler)

	// Test allowed request
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test rate limited request - use different client to avoid timing issues
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	w2 := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w2.Code)
	}

	// Check rate limit headers
	if w2.Header().Get("X-RateLimit-Limit") == "" {
		t.Error("Expected X-RateLimit-Limit header")
	}
}

// TestRateLimiter_EdgeCases tests various edge cases
func TestRateLimiter_EdgeCases(t *testing.T) {
	// Test with zero rate limit
	rateLimiter := NewRateLimiter(0, 1, time.Second, 1*time.Minute, 5*time.Minute)
	clientID := "test-client"

	// Should not allow any requests
	if rateLimiter.Allow(clientID) {
		t.Error("Expected rate limiter with 0 RPS to deny all requests")
	}

	// Test with very high burst size
	rateLimiter = NewRateLimiter(1, 1000, time.Second, 1*time.Minute, 5*time.Minute)

	// Should allow burst requests
	for i := 0; i < 1000; i++ {
		if !rateLimiter.Allow(clientID) {
			t.Errorf("Expected to allow burst request %d", i)
		}
	}

	// Should deny after burst
	if rateLimiter.Allow(clientID) {
		t.Error("Expected to deny request after burst")
	}
}

// TestRateLimiter_TimeBased tests time-based rate limiting
func TestRateLimiter_TimeBased(t *testing.T) {
	rateLimiter := NewRateLimiter(10, 2, time.Second, 1*time.Minute, 5*time.Minute) // 10 RPS, burst 2, 1 second window
	clientID := "test-client"

	// Should allow burst
	if !rateLimiter.Allow(clientID) {
		t.Error("Expected to allow first request")
	}
	if !rateLimiter.Allow(clientID) {
		t.Error("Expected to allow second request")
	}

	// Should deny immediately after burst
	if rateLimiter.Allow(clientID) {
		t.Error("Expected to deny request immediately after burst")
	}

	// Wait for rate limit to reset (100ms should be enough for 10 RPS)
	time.Sleep(150 * time.Millisecond)

	// Should allow requests again
	if !rateLimiter.Allow(clientID) {
		t.Error("Expected to allow request after rate limit reset")
	}

	// Test that we can get another token after waiting
	time.Sleep(100 * time.Millisecond)
	if !rateLimiter.Allow(clientID) {
		t.Error("Expected to allow request after additional wait")
	}
}

// TestRateLimiter_ConcurrentAccess tests concurrent access to rate limiter
func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	rateLimiter := NewRateLimiter(100, 200, time.Second, 1*time.Minute, 5*time.Minute)
	clientID := "test-client"

	// Test concurrent access
	done := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func() {
			rateLimiter.Allow(clientID)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 100; i++ {
		<-done
	}

	// Should not panic or cause race conditions
	// Test completed successfully
}

// TestRateLimiter_DifferentClients tests rate limiting for different clients
func TestRateLimiter_DifferentClients(t *testing.T) {
	rateLimiter := NewRateLimiter(1, 1, time.Second, 1*time.Minute, 5*time.Minute)

	// Each client should have independent rate limits
	clients := []string{"client1", "client2", "client3"}

	for _, client := range clients {
		if !rateLimiter.Allow(client) {
			t.Errorf("Expected to allow request for %s", client)
		}
		if rateLimiter.Allow(client) {
			t.Errorf("Expected to deny second request for %s", client)
		}
	}
}

// TestRateLimiter_Benchmark tests performance of rate limiter
func TestRateLimiter_Benchmark(t *testing.T) {
	rateLimiter := NewRateLimiter(1000, 1000, time.Second, 1*time.Minute, 5*time.Minute)
	clientID := "benchmark-client"

	// Benchmark rate limiter performance
	start := time.Now()
	for i := 0; i < 1000; i++ {
		rateLimiter.Allow(clientID)
	}
	duration := time.Since(start)

	// Should complete within reasonable time
	if duration > 100*time.Millisecond {
		t.Errorf("Rate limiter took too long: %v", duration)
	}

	t.Logf("Processed %d requests in %v", 1000, duration)
}

// TestRateLimiter_Stress tests rate limiter under stress
func TestRateLimiter_Stress(t *testing.T) {
	rateLimiter := NewRateLimiter(100, 200, time.Second, 1*time.Minute, 5*time.Minute)
	clientID := "stress-client"

	// Test with many requests
	allowed := 0
	denied := 0

	for i := 0; i < 1000; i++ {
		if rateLimiter.Allow(clientID) {
			allowed++
		} else {
			denied++
		}
	}

	// Should have allowed burst + some additional requests
	if allowed < 200 {
		t.Errorf("Expected at least 200 allowed requests, got %d", allowed)
	}

	// Should have denied some requests
	if denied == 0 {
		t.Error("Expected some requests to be denied")
	}

	t.Logf("Stress test: %d allowed, %d denied", allowed, denied)
}

// TestRateLimiter_ConcurrentStress tests rate limiter under concurrent stress
func TestRateLimiter_ConcurrentStress(t *testing.T) {
	rateLimiter := NewRateLimiter(100, 200, time.Second, 1*time.Minute, 5*time.Minute)
	clientID := "concurrent-stress-client"

	// Test with concurrent requests
	done := make(chan bool, 100)
	allowed := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func() {
			result := rateLimiter.Allow(clientID)
			allowed <- result
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 100; i++ {
		<-done
	}

	// Count results
	allowedCount := 0
	for i := 0; i < 100; i++ {
		if <-allowed {
			allowedCount++
		}
	}

	// Should have allowed some requests
	if allowedCount == 0 {
		t.Error("Expected some requests to be allowed")
	}

	t.Logf("Concurrent stress test: %d allowed out of 100", allowedCount)
}

func TestRateLimitMiddlewareWithContext(t *testing.T) {
	rateLimiter := NewRateLimiter(2, 1, time.Second, 1*time.Minute, 5*time.Minute)
	middleware := RateLimitMiddleware(rateLimiter)

	// Create a handler that checks context
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID := r.Context().Value(ClientIDKey)
		if clientID == nil {
			t.Error("Expected client ID in context")
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	wrappedHandler := middleware(handler)

	// Test first request (should succeed)
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Test second request (should be rate limited)
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	w = httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rateLimiter := NewRateLimiter(10, 5, 100*time.Millisecond, 1*time.Minute, 1*time.Minute)

	// Add some test clients
	rateLimiter.tokens["client1"] = 3
	rateLimiter.tokens["client2"] = 2
	rateLimiter.lastUpdate["client1"] = time.Now().Add(-2 * time.Minute)  // Old (2 minutes ago)
	rateLimiter.lastUpdate["client2"] = time.Now().Add(-30 * time.Second) // Recent (30 seconds ago)

	// Force cleanup
	rateLimiter.cleanup()

	// Check that old client was removed
	if _, exists := rateLimiter.tokens["client1"]; exists {
		t.Error("Expected old client to be cleaned up")
	}
	if _, exists := rateLimiter.lastUpdate["client1"]; exists {
		t.Error("Expected old client lastUpdate to be cleaned up")
	}

	// Check that recent client still exists
	if _, exists := rateLimiter.tokens["client2"]; !exists {
		t.Error("Expected recent client to remain")
	}
	if _, exists := rateLimiter.lastUpdate["client2"]; !exists {
		t.Error("Expected recent client lastUpdate to remain")
	}
}

func TestRateLimiter_ZeroRate(t *testing.T) {
	rateLimiter := NewRateLimiter(0, 5, time.Second, 1*time.Minute, 5*time.Minute)

	// All requests should be denied
	for i := 0; i < 5; i++ {
		allowed := rateLimiter.Allow("test-client")
		if allowed {
			t.Errorf("Request %d should have been denied with zero rate", i+1)
		}
	}
}

func TestRateLimiter_ClientIDEdgeCases(t *testing.T) {
	rateLimiter := NewRateLimiter(1, 1, time.Second, 1*time.Minute, 5*time.Minute)

	// Test empty client ID
	allowed := rateLimiter.Allow("")
	if !allowed {
		t.Error("Expected empty client ID to be allowed initially")
	}

	// Test very long client ID
	longClientID := strings.Repeat("a", 1000)
	allowed = rateLimiter.Allow(longClientID)
	if !allowed {
		t.Error("Expected long client ID to be allowed initially")
	}

	// Test special characters in client ID
	specialClientID := "client@#$%^&*()_+-=[]{}|;':\",./<>?"
	allowed = rateLimiter.Allow(specialClientID)
	if !allowed {
		t.Error("Expected special character client ID to be allowed initially")
	}
}

// TestRateLimiter_DifferentPorts tests that different ports from same IP are treated as same client
func TestRateLimiter_DifferentPorts(t *testing.T) {
	rateLimiter := NewRateLimiter(1, 2, time.Second, 1*time.Minute, 5*time.Minute)

	// Test that different ports from same IP are treated as same client
	// This is the bug that was fixed - before the fix, each port would be treated as different client

	// First request from port 1234
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	clientID1 := rateLimiter.GetClientID(req1)

	// Second request from port 1235 (same IP, different port)
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:1235"
	clientID2 := rateLimiter.GetClientID(req2)

	// Both should have same client ID (IP only, no port)
	if clientID1 != clientID2 {
		t.Errorf("Expected same client ID for different ports, got %s and %s", clientID1, clientID2)
	}

	// Test rate limiting with same IP but different ports
	// First request should be allowed
	if !rateLimiter.Allow(clientID1) {
		t.Error("Expected first request to be allowed")
	}

	// Second request should be allowed (burst size is 2)
	if !rateLimiter.Allow(clientID2) {
		t.Error("Expected second request to be allowed (burst size 2)")
	}

	// Third request should be denied (rate limit exceeded)
	if rateLimiter.Allow(clientID1) {
		t.Error("Expected third request to be denied (rate limit exceeded)")
	}
}

func TestRateLimiter_TokenRefill(t *testing.T) {
	rateLimiter := NewRateLimiter(2, 2, time.Second, 1*time.Minute, 5*time.Minute)

	// Use up all tokens
	rateLimiter.Allow("client1")
	rateLimiter.Allow("client1")

	// Should be rate limited
	if rateLimiter.Allow("client1") {
		t.Error("Expected to be rate limited after using all tokens")
	}

	// Wait for token refill
	time.Sleep(600 * time.Millisecond)

	// Should be allowed again
	if !rateLimiter.Allow("client1") {
		t.Error("Expected to be allowed after token refill")
	}
}

func TestRateLimiter_BurstSize(t *testing.T) {
	rateLimiter := NewRateLimiter(1, 5, time.Second, 1*time.Minute, 5*time.Minute)

	// Should be able to make burst requests
	for i := 0; i < 5; i++ {
		if !rateLimiter.Allow("client1") {
			t.Errorf("Expected burst request %d to be allowed", i+1)
		}
	}

	// Should be rate limited after burst
	if rateLimiter.Allow("client1") {
		t.Error("Expected to be rate limited after burst")
	}
}
