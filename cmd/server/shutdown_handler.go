package main

import (
	"os"
	"os/signal"
	"syscall"
)

// waitForShutdownSignal waits for interrupt signals to gracefully shutdown the server
// Handles SIGINT (Ctrl+C) and SIGTERM (Docker/Systemd termination)
func waitForShutdownSignal() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
