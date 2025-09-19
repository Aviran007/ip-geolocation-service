package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter implements a custom rate limiting mechanism
type RateLimiter struct {
	requestsPerSecond int
	burstSize         int

	// Token bucket implementation
	tokens     map[string]int
	lastUpdate map[string]time.Time
	mu         sync.RWMutex

	// Cleanup
	cleanupInterval   time.Duration
	inactiveThreshold time.Duration
	lastCleanup       time.Time
}

// NewRateLimiter creates a new rate limiter with optional cleanup configuration
func NewRateLimiter(requestsPerSecond, burstSize int, windowSize time.Duration, cleanupInterval, inactiveThreshold time.Duration) *RateLimiter {
	// Set defaults if not provided
	if cleanupInterval == 0 {
		cleanupInterval = 1 * time.Minute
	}
	if inactiveThreshold == 0 {
		inactiveThreshold = 5 * time.Minute
	}

	return &RateLimiter{
		requestsPerSecond: requestsPerSecond,
		burstSize:         burstSize,
		tokens:            make(map[string]int),
		lastUpdate:        make(map[string]time.Time),
		cleanupInterval:   cleanupInterval,
		inactiveThreshold: inactiveThreshold,
	}
}

// calculateCurrentTokens calculates the current number of tokens for a client
func (rl *RateLimiter) calculateCurrentTokens(clientID string, now time.Time) int {
	lastUpdate, exists := rl.lastUpdate[clientID]
	if !exists {
		return 0
	}

	timeElapsed := now.Sub(lastUpdate)
	timeElapsedSeconds := timeElapsed.Seconds()
	tokensToAdd := int(timeElapsedSeconds * float64(rl.requestsPerSecond))
	
	currentTokens := rl.tokens[clientID] + tokensToAdd
	if currentTokens > rl.burstSize {
		currentTokens = rl.burstSize
	}
	if currentTokens < 0 {
		currentTokens = 0
	}
	
	return currentTokens
}

// Allow checks if a request is allowed for the given client
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Periodic cleanup
	if now.Sub(rl.lastCleanup) > rl.cleanupInterval {
		rl.cleanup()
		rl.lastCleanup = now
	}

	// Initialize or reset client if needed
	if _, exists := rl.tokens[clientID]; !exists {
		if rl.requestsPerSecond == 0 {
			return false
		}
		rl.tokens[clientID] = rl.burstSize
		rl.lastUpdate[clientID] = now
	} else if now.Sub(rl.lastUpdate[clientID]) > rl.inactiveThreshold {
		// Reset inactive client
		rl.tokens[clientID] = rl.burstSize
		rl.lastUpdate[clientID] = now
	}

	// Calculate current tokens and update
	rl.tokens[clientID] = rl.calculateCurrentTokens(clientID, now)
	rl.lastUpdate[clientID] = now

	// Check if request is allowed and consume token
	if rl.tokens[clientID] > 0 {
		rl.tokens[clientID]--
		return true
	}

	return false
}

// GetClientID extracts client identifier from request
func (rl *RateLimiter) GetClientID(r *http.Request) string {
	// Try to get real IP from headers (for reverse proxy scenarios)
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		// Take the first IP from the comma-separated list
		if commaIdx := strings.Index(forwardedFor, ","); commaIdx > 0 {
			return strings.TrimSpace(forwardedFor[:commaIdx])
		}
		return strings.TrimSpace(forwardedFor)
	}

	// Fall back to remote address
	clientID := r.RemoteAddr
	if clientID == "" {
		clientID = "unknown"
	}

	// Extract just the IP address (remove port)
	if colonIdx := strings.LastIndex(clientID, ":"); colonIdx > 0 {
		clientID = clientID[:colonIdx]
	}

	return clientID
}

// cleanup removes old entries to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	now := time.Now()
	cutoff := now.Add(-rl.inactiveThreshold)

	for clientID, lastUpdate := range rl.lastUpdate {
		if lastUpdate.Before(cutoff) {
			delete(rl.tokens, clientID)
			delete(rl.lastUpdate, clientID)
		}
	}
}

// GetMapState returns the current state of the rate limiter maps for debugging
func (rl *RateLimiter) GetMapState() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	now := time.Now()
	clients := make(map[string]interface{})

	for clientID := range rl.tokens {
		lastUpdate, exists := rl.lastUpdate[clientID]
		if !exists {
			continue
		}

		timeSinceLastUpdate := now.Sub(lastUpdate)
		currentTokens := rl.calculateCurrentTokens(clientID, now)

		clients[clientID] = map[string]interface{}{
			"tokens":                    currentTokens,
			"last_update":               lastUpdate.Format("15:04:05.000"),
			"time_since_last_update_ms": timeSinceLastUpdate.Milliseconds(),
			"is_active":                 timeSinceLastUpdate < rl.inactiveThreshold,
		}
	}

	return map[string]interface{}{
		"total_clients": len(rl.tokens),
		"current_time":  now.Format("15:04:05.000"),
		"clients":       clients,
		"config": map[string]interface{}{
			"requests_per_second":        rl.requestsPerSecond,
			"burst_size":                 rl.burstSize,
			"inactive_threshold_minutes": rl.inactiveThreshold.Minutes(),
		},
	}
}

// RateLimitContextKey is used to store rate limit info in context
type RateLimitContextKey string

const (
	ClientIDKey RateLimitContextKey = "client_id"
)

// RateLimitMiddleware creates a middleware for rate limiting
func RateLimitMiddleware(rateLimiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientID := rateLimiter.GetClientID(r)

			// Add client ID to context
			ctx := context.WithValue(r.Context(), ClientIDKey, clientID)
			r = r.WithContext(ctx)

			if !rateLimiter.Allow(clientID) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rateLimiter.requestsPerSecond))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.WriteHeader(http.StatusTooManyRequests)

				errorResponse := `{"error": "Rate limit exceeded. Try again later."}`
				w.Write([]byte(errorResponse))
				return
			}

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rateLimiter.requestsPerSecond))

			next.ServeHTTP(w, r)
		})
	}
}

// DebugRateLimitMiddleware creates a middleware for debug endpoints with higher limits
func DebugRateLimitMiddleware(rateLimiter *RateLimiter) func(http.Handler) http.Handler {
	// Create debug rate limiter once, not on every request
	debugRateLimiter := NewRateLimiter(100, 200, 1*time.Second, 1*time.Minute, 5*time.Minute)
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only apply to debug endpoints
			if !strings.HasPrefix(r.URL.Path, "/debug/") {
				next.ServeHTTP(w, r)
				return
			}

			clientID := rateLimiter.GetClientID(r)

			// Add client ID to context
			ctx := context.WithValue(r.Context(), ClientIDKey, clientID)
			r = r.WithContext(ctx)

			if !debugRateLimiter.Allow(clientID) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-RateLimit-Limit", "100")
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.WriteHeader(http.StatusTooManyRequests)

				errorResponse := `{"error": "Debug endpoint rate limit exceeded. Try again later."}`
				w.Write([]byte(errorResponse))
				return
			}

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", "100")

			next.ServeHTTP(w, r)
		})
	}
}
