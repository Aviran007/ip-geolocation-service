package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"ip-geolocation-service/internal/middleware"
	"ip-geolocation-service/internal/services"
)

// Router handles HTTP routing
type Router struct {
	ipHandler   *IPHandler
	rateLimiter interface {
		GetMapState() map[string]interface{}
	}
	logger      *slog.Logger
}

// NewRouter creates a new router
func NewRouter(ipService services.IPService, logger *slog.Logger) *Router {
	return &Router{
		ipHandler: NewIPHandler(ipService, logger),
		logger:    logger,
	}
}

// NewRouterWithRateLimiter creates a new router with rate limiter
func NewRouterWithRateLimiter(ipService services.IPService, rateLimiter interface{GetMapState() map[string]interface{}}, logger *slog.Logger) *Router {
	return &Router{
		ipHandler:   NewIPHandler(ipService, logger),
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

// SetupRoutes configures all routes
func (r *Router) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// API v1 routes
	v1 := http.NewServeMux()
	v1.HandleFunc("/find-country", r.ipHandler.FindCountry)

	// Wrap v1 routes with middleware
	mux.Handle("/v1/", http.StripPrefix("/v1", v1))

	// Health endpoint
	mux.HandleFunc("/health", r.ipHandler.HealthCheck)

	// Debug endpoint for rate limiter state
	mux.HandleFunc("/debug/rate-limiter", r.debugRateLimiter)

	// Root endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			r.ipHandler.NotFound(w, req)
		} else {
			http.NotFound(w, req)
		}
	})

	return mux
}

// debugRateLimiter shows the current state of the rate limiter
func (r *Router) debugRateLimiter(w http.ResponseWriter, req *http.Request) {
	if r.rateLimiter == nil {
		http.Error(w, "Rate limiter not available", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	
	state := r.rateLimiter.GetMapState()
	
	// Pretty print JSON
	jsonData, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		http.Error(w, "Failed to marshal state", http.StatusInternalServerError)
		return
	}
	
	w.Write(jsonData)
}

// SetupRoutesWithMiddleware configures routes with all middleware
func (r *Router) SetupRoutesWithMiddleware(rateLimiter *middleware.RateLimiter) http.Handler {
	// Create the base router
	mux := r.SetupRoutes()

	// Apply middleware in order (last applied is first executed)
	var handler http.Handler = mux

	// Security headers
	handler = middleware.SecurityHeadersMiddleware()(handler)

	// CORS
	handler = middleware.CORSMiddleware()(handler)

	// Debug rate limiting (higher limits for debug endpoints)
	handler = middleware.DebugRateLimitMiddleware(rateLimiter)(handler)

	// Regular rate limiting
	handler = middleware.RateLimitMiddleware(rateLimiter)(handler)

	// Logging
	handler = middleware.LoggingMiddleware(r.logger)(handler)

	// Recovery (should be first to catch panics)
	handler = middleware.RecoveryMiddleware(r.logger)(handler)

	return handler
}
