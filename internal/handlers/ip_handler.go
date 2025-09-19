package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"ip-geolocation-service/internal/middleware"
	"ip-geolocation-service/internal/models"
	"ip-geolocation-service/internal/services"
)

// IPHandler handles IP location requests
type IPHandler struct {
	service services.IPService
	logger  *slog.Logger
}

// NewIPHandler creates a new IP handler
func NewIPHandler(service services.IPService, logger *slog.Logger) *IPHandler {
	return &IPHandler{
		service: service,
		logger:  logger,
	}
}

// FindCountry handles GET /v1/find-country requests
func (h *IPHandler) FindCountry(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Only allow GET requests
	if r.Method != http.MethodGet {
		h.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get IP from query parameter
	ip := r.URL.Query().Get("ip")
	if ip == "" {
		h.sendError(w, "Missing required parameter: ip", http.StatusBadRequest)
		return
	}

	// Add request context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Add client ID to context if available
	var clientID interface{}
	if clientID = r.Context().Value(middleware.ClientIDKey); clientID != nil {
		ctx = context.WithValue(ctx, middleware.ClientIDKey, clientID)
	}

	// Log the request
	h.logger.Info("🔍 Processing IP lookup request",
		"ip", ip,
		"client_id", clientID,
	)

	// Find location
	location, err := h.service.FindLocation(ctx, ip)
	if err != nil {
		h.logger.Error("❌ Failed to find location",
			"ip", ip,
			"error", err,
		)

		// Determine appropriate error response based on error type
		switch {
		case strings.Contains(err.Error(), "location not found"):
			h.sendError(w, "Location not found for the provided IP address", http.StatusNotFound)
		case strings.Contains(err.Error(), "invalid IP address"):
			h.sendError(w, "Invalid IP address format", http.StatusBadRequest)
		case strings.Contains(err.Error(), "invalid location data"):
			h.sendError(w, "Invalid location data", http.StatusInternalServerError)
		default:
			h.sendError(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Send successful response
	h.sendSuccess(w, location)
}

// sendSuccess sends a successful response
func (h *IPHandler) sendSuccess(w http.ResponseWriter, location *models.Location) {
	w.WriteHeader(http.StatusOK)

	response, err := location.ToJSON()
	if err != nil {
		h.logger.Error("Failed to marshal location response", "error", err)
		h.sendError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Write(response)
}

// sendError sends an error response
func (h *IPHandler) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)

	errorResp := models.NewErrorResponse(message)
	response, err := errorResp.ToJSON()
	if err != nil {
		h.logger.Error("Failed to marshal error response", "error", err)
		// Fallback to plain text
		w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, message)))
		return
	}

	w.Write(response)
}

// HealthCheck handles health check requests
func (h *IPHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	// Check service health
	if err := h.service.HealthCheck(ctx); err != nil {
		h.logger.Error("Health check failed", "error", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status": "unhealthy", "error": "` + err.Error() + `"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}

// NotFound handles 404 requests
func (h *IPHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	h.sendError(w, "Not found", http.StatusNotFound)
}

// MethodNotAllowed handles 405 requests
func (h *IPHandler) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	h.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
}
