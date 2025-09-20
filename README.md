# IP Geolocation Service

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/Aviran007/ip-geolocation-service)
[![Coverage](https://img.shields.io/badge/coverage-85.7%25-green.svg)](https://github.com/Aviran007/ip-geolocation-service)

A production-ready Go microservice that provides IP-to-country/city geolocation data via REST API with rate limiting, clean architecture, and comprehensive testing.

## 🚀 Features

- **REST API**: Clean RESTful endpoints for IP geolocation lookup
- **Rate Limiting**: Custom token bucket implementation with configurable limits
- **Clean Architecture**: Well-structured code following Go best practices
- **CSV Data Source**: Currently supports CSV file-based data storage
- **Production Ready**: Health checks, graceful shutdown, Docker support
- **Comprehensive Testing**: Unit tests, integration tests, and benchmarks
- **Rate Limiter Testing**: Built-in scripts for testing rate limiting behavior
- **Observability**: Structured logging with configurable levels and formats
- **Docker Support**: Multi-stage Docker build with Docker Compose
- **Development Tools**: Comprehensive Makefile with development commands

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [API Documentation](#api-documentation)
- [Configuration](#configuration)
- [Architecture](#architecture)
- [Development](#development)
- [Testing](#testing)
- [Deployment](#deployment)
- [Contributing](#contributing)
- [License](#license)

## 🏃 Quick Start

### Prerequisites

- Go 1.21 or later
- Docker (optional, for containerized deployment)

### Using Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/Aviran007/ip-geolocation-service.git
cd ip-geolocation-service

# Build and run with Docker Compose
make docker-compose-up

# The service will be available at http://localhost:8080

# Stop the service
make docker-compose-down
```

### Local Development

```bash
# Clone the repository
git clone https://github.com/Aviran007/ip-geolocation-service.git
cd ip-geolocation-service

# Setup development environment
make setup

# Run the service
make run

# Run in development mode (with debug logging)
make run-dev

# Run in production mode
make run-prod
```

## 📚 API Documentation

### Find Country by IP

```bash
# Get location for an IP address
curl "http://localhost:8080/v1/find-country?ip=8.8.8.8"

# Response
{
  "country": "United States",
  "city": "Mountain View"
}
```

### Health Check

```bash
# Check service health
curl "http://localhost:8080/health"

# Response
{
  "status": "healthy"
}
```


### Error Responses

```bash
# Invalid IP address
curl "http://localhost:8080/v1/find-country?ip=invalid-ip"

# Response (400 Bad Request)
{
  "error": "Invalid IP address format"
}

# Rate limit exceeded
curl "http://localhost:8080/v1/find-country?ip=8.8.8.8"
# After exceeding rate limit

# Response (429 Too Many Requests)
{
  "error": "Rate limit exceeded. Try again later."
}
```

## ⚙️ Configuration

The service can be configured using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `DATABASE_TYPE` | `csv` | Database type (currently only csv supported) |
| `DATABASE_FILE_PATH` | `./data/ip_locations.csv` | Path to CSV data file |
| `RATE_LIMIT_RPS` | `20` | Requests per second limit |
| `RATE_LIMIT_BURST` | `20` | Burst size for rate limiting |
| `RATE_LIMIT_CLEANUP_INTERVAL` | `1m` | Rate limiter cleanup interval |
| `RATE_LIMIT_INACTIVE_THRESHOLD` | `5m` | Inactive client cleanup threshold |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `LOG_FORMAT` | `json` | Log format (json, text) |
| `READ_TIMEOUT` | `30s` | HTTP read timeout |
| `WRITE_TIMEOUT` | `30s` | HTTP write timeout |
| `IDLE_TIMEOUT` | `120s` | HTTP idle timeout |

## 🏗️ Architecture

The service follows Clean Architecture principles with clear separation of concerns:

```
cmd/server/          # Application entry point
internal/
├── config/          # Configuration management
├── handlers/        # HTTP handlers
├── services/        # Business logic
├── models/          # Data models and validation
├── middleware/      # HTTP middleware
└── repository/      # Data access layer
data/               # Sample data files
scripts/            # Testing and utility scripts
```

### Architecture Diagram

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Client    │───▶│Load Balancer│───▶│ API Server  │
└─────────────┘    └─────────────┘    └─────────────┘
                                              │
                                              ▼
                                    ┌─────────────────┐
                                    │ Middleware Stack│
                                    │ ┌─────────────┐ │
                                    │ │  Recovery   │ │
                                    │ │  Logging    │ │
                                    │ │ Rate Limit  │ │
                                    │ │    CORS     │ │
                                    │ └─────────────┘ │
                                    └─────────────────┘
                                              │
                                              ▼
                                    ┌─────────────────┐
                                    │  Handler Layer  │
                                    │ ┌─────────────┐ │
                                    │ │ IP Handler  │ │
                                    │ │Health Check │ │
                                    │ │Debug Handler│ │
                                    │ └─────────────┘ │
                                    └─────────────────┘
                                              │
                                              ▼
                                    ┌─────────────────┐
                                    │  Service Layer  │
                                    │ ┌─────────────┐ │
                                    │ │ IP Service  │ │
                                    │ │IP Validator │ │
                                    │ └─────────────┘ │
                                    └─────────────────┘
                                              │
                                              ▼
                                    ┌─────────────────┐
                                    │Repository Layer │
                                    │ ┌─────────────┐ │
                                    │ │Repository   │ │
                                    │ │Interface    │ │
                                    │ │File Repo    │ │
                                    │ └─────────────┘ │
                                    └─────────────────┘
                                              │
                                              ▼
                                    ┌─────────────────┐
                                    │   Data Layer    │
                                    │ ┌─────────────┐ │
                                    │ │CSV Data File│ │
                                    │ └─────────────┘ │
                                    └─────────────────┘
```

For detailed Mermaid diagrams with request flow and component details, see the [Architecture Diagram](docs/architecture-diagram.md).

### Key Components

- **Handlers**: HTTP request/response handling with proper error handling
- **Services**: Business logic and orchestration
- **Repositories**: Data access abstraction with interface-based design
- **Middleware**: Rate limiting, logging, security, and recovery
- **Models**: Data structures, validation, and serialization

### Design Patterns

- **Dependency Injection**: Constructor-based dependency injection
- **Interface Segregation**: Small, focused interfaces
- **Repository Pattern**: Data access abstraction
- **Middleware Pattern**: Cross-cutting concerns
- **Factory Pattern**: Repository creation

## 🛠️ Development

### Project Structure

```
ip-geolocation-service/
├── cmd/
│   └── server/          # Application entry point
│       └── main.go
├── internal/
│   ├── config/          # Configuration management
│   │   ├── config.go
│   │   └── config_test.go
│   ├── handlers/        # HTTP handlers
│   │   ├── ip_handler.go
│   │   ├── ip_handler_test.go
│   │   ├── router.go
│   │   ├── router_test.go
│   │   └── mocks.go
│   ├── services/        # Business logic
│   │   ├── ip_service.go
│   │   ├── ip_service_test.go
│   │   └── mocks.go
│   ├── models/          # Data models
│   │   ├── location.go
│   │   └── location_test.go
│   ├── middleware/      # HTTP middleware
│   │   ├── logging.go
│   │   ├── logging_test.go
│   │   ├── rate_limiter.go
│   │   └── rate_limiter_test.go
│   └── repository/      # Data access layer
│       ├── interfaces.go
│       ├── factory.go
│       ├── factory_test.go
│       ├── file_repository.go
│       └── file_repository_test.go
├── data/                # Sample data files
│   └── ip_locations.csv
├── scripts/             # Testing and utility scripts
│   ├── README.md
│   └── test_3_clients.sh
├── Dockerfile           # Multi-stage Docker build
├── docker-compose.yml   # Docker Compose configuration
├── env.example          # Environment variables example
├── Makefile            # Build automation
├── go.mod              # Go module definition
└── README.md           # This file
```

### Code Quality

- **Go Modules**: Modern dependency management
- **Go Lint**: Code quality enforcement (`make lint`)
- **Go Vet**: Static analysis (`make vet`)
- **Go Fmt**: Code formatting (`make fmt`)
- **Go Test**: Comprehensive testing (`make test`)
- **Go Bench**: Performance benchmarking (`make benchmark`)

### Build Commands

```bash
# Show all available commands
make help

# Build the application
make build

# Run the application
make run

# Run tests
make test

# Run tests with coverage
make test-coverage

# Run specific test packages
make test-models
make test-middleware
make test-repository
make test-handlers
make test-services

# Code quality
make lint
make fmt
make vet

# Development setup
make setup
make install-tools

# Run with different configurations
make run-dev    # Development mode
make run-prod   # Production mode

# Clean up
make clean
make clean-port # Clean port 8080
```

### Development Workflow

```bash
# Complete development cycle
make dev        # Clean, format, vet, test, and build
make ci         # Clean, format, vet, test, and coverage
```

## 🧪 Testing

The project includes comprehensive testing:

- **Unit Tests**: Individual component testing
- **Integration Tests**: Component interaction testing
- **Benchmark Tests**: Performance testing
- **Table-Driven Tests**: Parameterized testing
- **Parallel Tests**: Concurrent test execution
- **Mock Testing**: Interface-based mocking
- **Rate Limiter Testing**: Built-in scripts for testing rate limiting behavior
- **API Testing**: Automated API endpoint testing
- **Load Testing**: Performance testing with multiple concurrent requests

### Test Coverage

Current test coverage:
- **Models**: 100.0% coverage
- **Repository**: 88.0% coverage
- **Services**: 85.7% coverage
- **Config**: 89.5% coverage
- **Handlers**: 79.8% coverage
- **Middleware**: 73.3% coverage
- **Overall**: 85.7% coverage

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
make test-models
make test-middleware
make test-repository
make test-handlers
make test-services

# Run tests in parallel
go test -parallel 4 ./...

# Run benchmarks
make benchmark
```

### Testing Scripts

The project includes testing scripts in the `scripts/` directory:
- **`test_3_clients.sh`**: Tests rate limiting with 3 concurrent clients (50 requests each)
- **Rate Limiter Tests**: Built-in Makefile targets for testing rate limiting behavior
- **API Tests**: Automated testing of all API endpoints
- **Load Tests**: Performance testing with configurable load

Available test commands:
```bash
make test-3-clients      # Test with 3 concurrent clients
make test-rate-limit-single  # Test with single client
make test-api            # Test API endpoints
make load-test           # Load testing
```

## 🚀 Deployment

### Docker Deployment

```bash
# Build Docker image
make docker-build

# Run container
make docker-run

# Using Docker Compose
make docker-compose-up
make docker-compose-down

# Advanced Docker commands
make docker-build-run    # Build and run in one command
make docker-restart      # Stop, build, and run
```

### Docker Compose Configuration

- **Main Service**: IP geolocation service with health checks
- **Nginx**: Optional reverse proxy for production (profile: production)
- **Health Checks**: Built-in health monitoring
- **Volume Mounting**: Data directory mounted as read-only

### Production Considerations

- **Health Checks**: Built-in health check endpoint
- **Graceful Shutdown**: Proper signal handling
- **Resource Limits**: Memory and CPU limits
- **Logging**: Structured JSON logging
- **Security**: Security headers and CORS

## 📊 Performance

### Benchmarks

- **IP Validation**: 7,000 IPs validated in 491µs
- **Rate Limiting**: 1,000 requests processed in 435µs
- **JSON Serialization**: High-performance JSON marshaling
- **Concurrent Access**: Thread-safe operations

### Rate Limiting

- **Token Bucket Algorithm**: Smooth rate limiting with burst capacity
- **Per-Client Limiting**: Based on client IP address
- **Configurable**: RPS and burst size via environment variables
- **Cleanup**: Automatic cleanup of inactive clients
- **Headers**: Rate limit information in response headers

### Performance Testing

```bash
# Run benchmarks
make benchmark

# Load testing with hey tool
make load-test
```

### Memory Management

- **Efficient Data Structures**: Optimized for memory usage
- **Garbage Collection**: Proper resource cleanup
- **Rate Limiter Cleanup**: Automatic cleanup of inactive clients

