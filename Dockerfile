# Multi-stage Dockerfile for Web Page Analyzer
# Supports both production and debug builds

# Stage 1: Builder
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /build/bin/webpage-analyzer \
    ./cmd

# Stage 2: Debug (with delve)
FROM golang:1.24-alpine AS debug

# Install delve debugger
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# Set working directory
WORKDIR /app

# Copy source code for debugging
COPY --from=builder /build /app

# Expose application and debugger ports
EXPOSE 8080 2345

# Run with delve
CMD ["dlv", "debug", "./cmd", "--headless", "--listen=:2345", "--api-version=2", "--accept-multiclient"]

# Stage 3: Production (minimal runtime)
FROM alpine:latest AS production

# Install CA certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bin/webpage-analyzer .

# Copy web assets
COPY --from=builder /build/web ./web

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose application port
EXPOSE 8080

# Run the application
CMD ["./webpage-analyzer"]
