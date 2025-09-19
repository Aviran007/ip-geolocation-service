package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ip-geolocation-service/internal/config"
	"ip-geolocation-service/internal/handlers"
	"ip-geolocation-service/internal/middleware"
	"ip-geolocation-service/internal/repository"
	"ip-geolocation-service/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Setup logger
	logger := setupLogger(cfg.Logging)

	logger.Info("🚀 Starting IP Geolocation Service",
		"port", cfg.Server.Port,
		"database_type", cfg.Database.Type,
		"rate_limit_rps", cfg.RateLimit.RequestsPerSecond,
		"log_level", cfg.Logging.Level,
	)

	// Create repository factory
	repoFactory := repository.NewRepositoryFactory(&cfg.Database, nil)

	// Create repository
	repo, err := repoFactory.CreateRepositoryFromConfig()
	if err != nil {
		logger.Error("Failed to create repository", "error", err)
		os.Exit(1)
	}

	// Initialize repository
	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		logger.Error("Failed to initialize repository", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			logger.Error("Failed to close repository", "error", err)
		}
	}()

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

	// Start server in a goroutine
	go func() {
		logger.Info("🌐 Server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("❌ Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("🛑 Shutting down server...")

	// Create a deadline for shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("❌ Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("✅ Server exited gracefully")
}

// setupLogger configures the logger based on configuration
func setupLogger(cfg config.LoggingConfig) *slog.Logger {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	}

	return slog.New(handler)
}
