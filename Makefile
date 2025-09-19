# Makefile for IP Geolocation Service

.PHONY: help build run test test-coverage clean docker-build docker-run docker-compose-up docker-compose-down lint fmt vet test-3-clients test-rate-limit-single test-api load-test run-dev run-prod

# Default target
help: ## Show this help message
	@echo "IP Geolocation Service - Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build the application
build: ## Build the application
	@echo "🔨 Building IP Geolocation Service..."
	@echo "=========================================="
	go build -o bin/ip-geolocation-service cmd/server/main.go
	@echo "=========================================="
	@echo "✅ Build completed! Binary: bin/ip-geolocation-service"

# Run the application
run: ## Run the application locally
	@echo "🚀 Running IP Geolocation Service..."
	@echo "=========================================="
	go run cmd/server/main.go

# Run tests
test: ## Run all tests
	@echo "🧪 Running tests..."
	@echo "=========================================="
	go test -v ./...
	@echo "=========================================="
	@echo "✅ All tests completed!"

# Run tests with coverage
test-coverage: ## Run tests with coverage report
	@echo "📊 Running tests with coverage..."
	@echo "=========================================="
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "=========================================="
	@echo "📈 Coverage report generated: coverage.html"
	@echo "🌐 Open coverage.html in your browser to view detailed coverage"

# Run benchmarks
benchmark: ## Run performance benchmarks
	@echo "⚡ Running benchmarks..."
	@echo "=========================================="
	go test -bench=. -benchmem ./...
	@echo "=========================================="
	@echo "✅ Benchmarks completed!"

# Run specific package tests
test-models: ## Run model tests
	@echo "🧪 Running model tests..."
	@echo "=========================================="
	go test -v ./internal/models
	@echo "=========================================="
	@echo "✅ Model tests completed!"

test-middleware: ## Run middleware tests
	@echo "🧪 Running middleware tests..."
	@echo "=========================================="
	go test -v ./internal/middleware
	@echo "=========================================="
	@echo "✅ Middleware tests completed!"

test-repository: ## Run repository tests
	@echo "🧪 Running repository tests..."
	@echo "=========================================="
	go test -v ./internal/repository
	@echo "=========================================="
	@echo "✅ Repository tests completed!"

test-handlers: ## Run handler tests
	@echo "🧪 Running handler tests..."
	@echo "=========================================="
	go test -v ./internal/handlers
	@echo "=========================================="
	@echo "✅ Handler tests completed!"

test-services: ## Run service tests
	@echo "🧪 Running service tests..."
	@echo "=========================================="
	go test -v ./internal/services
	@echo "=========================================="
	@echo "✅ Service tests completed!"

# Code quality
lint: ## Run linter
	@echo "🔍 Running linter..."
	@echo "=========================================="
	golangci-lint run
	@echo "=========================================="
	@echo "✅ Linting completed!"

fmt: ## Format code
	@echo "🎨 Formatting code..."
	@echo "=========================================="
	go fmt ./...
	@echo "=========================================="
	@echo "✅ Code formatting completed!"

vet: ## Run go vet
	@echo "🔍 Running go vet..."
	@echo "=========================================="
	go vet ./...
	@echo "=========================================="
	@echo "✅ Go vet completed!"

# Clean up
clean: ## Clean build artifacts
	@echo "Cleaning up..."
	rm -rf bin/
	rm -f coverage.out coverage.html

clean-port: ## Clean port 8080
	@echo "🧹 Cleaning port 8080..."
	@echo "=========================================="
	@lsof -ti:8080 | xargs kill -9 2>/dev/null || echo "No processes found on port 8080"
	@echo "=========================================="
	@echo "✅ Port 8080 cleaned!"

# Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t ip-geolocation-service .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 \
		-e DATABASE_FILE_PATH=/app/data/ip_locations.csv \
		-v $(PWD)/data:/app/data:ro \
		ip-geolocation-service

docker-compose-up: ## Start services with Docker Compose
	@echo "Starting services with Docker Compose..."
	docker-compose up --build

docker-compose-down: ## Stop services with Docker Compose
	@echo "Stopping services with Docker Compose..."
	docker-compose down

docker-build-run: ## Build and run Docker container
	@echo "🔨 Building and running Docker container..."
	@echo "=========================================="
	@$(MAKE) docker-build
	@echo "=========================================="
	@echo "🚀 Starting Docker container..."
	@$(MAKE) docker-run

docker-restart: ## Stop, build, and run Docker container
	@echo "🔄 Restarting Docker container..."
	@echo "=========================================="
	@echo "🛑 Stopping existing containers..."
	@docker stop $$(docker ps -q --filter "ancestor=ip-geolocation-service") 2>/dev/null || true
	@echo "=========================================="
	@$(MAKE) docker-build-run

# Development setup
setup: ## Setup development environment
	@echo "Setting up development environment..."
	go mod download
	go mod tidy
	@echo "Development environment ready!"

# Install tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run with different configurations
run-dev: ## Run with development configuration
	@echo "Running in development mode..."
	PORT=8080 DATABASE_TYPE=csv DATABASE_FILE_PATH=./data/ip_locations.csv RATE_LIMIT_RPS=5 RATE_LIMIT_BURST=10 LOG_LEVEL=debug go run cmd/server/main.go

run-prod: ## Run with production configuration
	@echo "Running in production mode..."
	PORT=8080 DATABASE_TYPE=csv DATABASE_FILE_PATH=./data/ip_locations.csv RATE_LIMIT_RPS=20 RATE_LIMIT_BURST=20 LOG_LEVEL=info LOG_FORMAT=json go run cmd/server/main.go

# API testing
test-api: ## Test API endpoints
	@echo "Testing API endpoints..."
	@echo "Testing health endpoint..."
	curl -s http://localhost:8080/health | jq .
	@echo ""
	@echo "Testing find-country endpoint..."
	curl -s "http://localhost:8080/v1/find-country?ip=8.8.8.8" | jq .

# Load testing
load-test: ## Run load tests (requires hey tool)
	@echo "Running load tests..."
	@which hey > /dev/null || (echo "Please install hey: go install github.com/rakyll/hey@latest" && exit 1)
	hey -n 1000 -c 10 "http://localhost:8080/v1/find-country?ip=8.8.8.8"

# Rate limiter testing
test-3-clients: ## Test rate limiter with 3 clients (50 requests each)
	@echo "Testing rate limiter with 3 clients..."
	@chmod +x scripts/test_3_clients.sh
	@./scripts/test_3_clients.sh

test-rate-limit-single: ## Test rate limiter with single client (25 requests)
	@echo "Testing rate limiter with single client..."
	@for i in {1..25}; do \
		echo "Request $$i:"; \
		curl -s -H "X-Forwarded-For: 192.168.1.200" "http://localhost:8080/v1/find-country?ip=8.8.8.8" | jq -r '.country // "RATE LIMITED"'; \
	done

# All-in-one commands
dev: clean fmt vet test build ## Full development cycle
	@echo "Development cycle complete!"

ci: clean fmt vet test test-coverage ## CI pipeline
	@echo "CI pipeline complete!"

# Default target
.DEFAULT_GOAL := help
