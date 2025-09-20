package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"time"
)

// LoggingMiddleware creates a middleware for request logging
func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap the ResponseWriter to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process the request
			next.ServeHTTP(wrapped, r)

			// Log the request
			duration := time.Since(start)

			// Extract client IP more cleanly
			clientIP := getClientIP(r)

			// Create a more readable log message
			logger.Info("Request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration", duration.String(),
				"client_ip", clientIP,
				"user_agent", r.UserAgent(),
			)
		})
	}
}

// getClientIP extracts the real client IP from request
func getClientIP(r *http.Request) string {
	// Check for real IP header (from reverse proxy)
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Check for forwarded IP
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		// Take the first IP from the list
		if idx := len(forwardedFor); idx > 0 {
			for i, c := range forwardedFor {
				if c == ',' {
					return forwardedFor[:i]
				}
			}
		}
		return forwardedFor
	}

	// Extract IP from RemoteAddr (remove port)
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}

	return r.RemoteAddr
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
