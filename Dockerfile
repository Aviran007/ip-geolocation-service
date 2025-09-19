# Multi-stage build for production-ready Go application
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod first for better caching
COPY go.mod ./

# Download dependencies (if any)
RUN go mod download

# Copy source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

# Final stage
FROM alpine:latest
WORKDIR /app

# Install only what we need
RUN apk --no-cache add ca-certificates tzdata wget

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copy files with correct ownership
COPY --from=builder --chown=appuser:appgroup /app/main .
COPY --from=builder --chown=appuser:appgroup /app/data ./data

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health >/dev/null 2>&1 || exit 1

EXPOSE 8080
CMD ["./main"]
