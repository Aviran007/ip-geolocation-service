package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"ip-geolocation-service/internal/config"
	"ip-geolocation-service/internal/handlers"
	"ip-geolocation-service/internal/middleware"
	"ip-geolocation-service/internal/repository"
	"ip-geolocation-service/internal/services"
)

// App represents the application and its dependencies
type App struct {
	config      *config.Config
	logger      *slog.Logger
	server      *http.Server
	repository  repository.IPRepository
	ipService   services.IPService
	rateLimiter *middleware.RateLimiter
}

// NewApp creates a new application instance with all dependencies
func NewApp(cfg *config.Config) (*App, error) {
	logger := setupLogger(cfg.Logging)

	// Create repository factory
	repoFactory := repository.NewRepositoryFactory(&cfg.Database)

	// Create repository
	repo, err := repoFactory.CreateRepositoryFromConfig()
	if err != nil {
		return nil, err
	}

	// Initialize repository
	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		return nil, err
	}

	// Create service
	ipService := services.NewIPService(repo)

	// Create rate limiter
	rateLimiter := middleware.NewRateLimiter(
		cfg.RateLimit.RequestsPerSecond,
		cfg.RateLimit.BurstSize,
		1*time.Second, // windowSize - not used but required for compatibility
		cfg.RateLimit.CleanupInterval,
		cfg.RateLimit.InactiveThreshold,
	)

	// Create router with rate limiter
	router := handlers.NewRouterWithRateLimiter(ipService, rateLimiter, logger)

	// Setup routes with middleware
	handler := router.SetupRoutesWithMiddleware(rateLimiter)

	// Create server
	server := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &App{
		config:      cfg,
		logger:      logger,
		server:      server,
		repository:  repo,
		ipService:   ipService,
		rateLimiter: rateLimiter,
	}, nil
}

// Start starts the application server
func (a *App) Start() error {
	a.logger.Info("🚀 Starting IP Geolocation Service",
		"port", a.config.Server.Port,
		"database_type", a.config.Database.Type,
		"rate_limit_rps", a.config.RateLimit.RequestsPerSecond,
		"log_level", a.config.Logging.Level,
	)

	// Start server in a goroutine
	go func() {
		a.logger.Info("🌐 Server starting", "addr", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("❌ Server failed to start", "error", err)
		}
	}()

	return nil
}

// Stop gracefully stops the application
func (a *App) Stop() error {
	a.logger.Info("🛑 Shutting down server...")

	// Close repository
	if err := a.repository.Close(); err != nil {
		a.logger.Error("Failed to close repository", "error", err)
	}

	// Create a deadline for shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := a.server.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("❌ Server forced to shutdown", "error", err)
		return err
	}

	a.logger.Info("✅ Server exited gracefully")
	return nil
}
