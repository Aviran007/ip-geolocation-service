package main

import (
	"fmt"
	"os"

	"ip-geolocation-service/internal/config"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Create application
	app, err := NewApp(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create application: %v\n", err)
		os.Exit(1)
	}

	// Start application
	if err := app.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start application: %v\n", err)
		os.Exit(1)
	}

	// Wait for shutdown signal
	waitForShutdownSignal()

	// Stop application gracefully
	if err := app.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stop application: %v\n", err)
		os.Exit(1)
	}
}
