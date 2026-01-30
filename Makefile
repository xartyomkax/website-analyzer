# Makefile for Web Page Analyzer

# Variables
BINARY_NAME=webpage-analyzer
MAIN_PATH=./cmd
BUILD_DIR=./bin
DOCKER_IMAGE=webpage-analyzer
DOCKER_TAG=latest

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOFMT=$(GOCMD) fmt
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet


.PHONY: all build test clean fmt tidy run help docker-build docker-run docker-clean test-coverage vet

# Default target
all: clean fmt tidy test build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Run the application
run: build
	@echo "Starting $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME)

# Run with delve debugger
debug:
	@echo "Starting debugger on port 2345..."
	dlv debug $(MAIN_PATH) --headless --listen=:2345 --api-version=2 --accept-multiclient

# Build Docker image
docker-build:
	@echo "Building Docker image: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built successfully"

# Build Docker debug image
docker-build-debug:
	@echo "Building Docker debug image: $(DOCKER_IMAGE):debug"
	docker build --target debug -t $(DOCKER_IMAGE):debug .
	@echo "Docker debug image built successfully"

APP_PORT ?= 8080
DEBUG_PORT ?= 2345

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p $(APP_PORT):$(APP_PORT) --name $(DOCKER_IMAGE) --rm $(DOCKER_IMAGE):$(DOCKER_TAG)

# Run Docker debug container
docker-run-debug:
	@echo "Running Docker debug container..."
	docker run -p $(APP_PORT):$(APP_PORT) -p $(DEBUG_PORT):$(DEBUG_PORT) --name $(DOCKER_IMAGE)-debug --rm $(DOCKER_IMAGE):debug

# Stop and remove Docker container
docker-clean:
	@echo "Cleaning Docker resources..."
	@docker stop $(DOCKER_IMAGE) 2>/dev/null || true
	@docker rm $(DOCKER_IMAGE) 2>/dev/null || true
	@docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true
	@docker rmi $(DOCKER_IMAGE):debug 2>/dev/null || true
	@echo "Docker cleanup complete"

# Show help
help:
	@echo "Web Page Analyzer - Makefile Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all              - Run clean, fmt, tidy, test, and build"
	@echo "  build            - Build the application binary"
	@echo "  test             - Run all tests"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  vet              - Run go vet"
	@echo "  fmt              - Format Go code"
	@echo "  tidy             - Tidy Go module dependencies"
	@echo "  clean            - Remove build artifacts"
	@echo "  run              - Build and run the application"
	@echo "  debug            - Run with delve debugger (port 2345)"
	@echo "  docker-build     - Build Docker production image"
	@echo "  docker-build-debug - Build Docker debug image"
	@echo "  docker-run       - Run Docker production container"
	@echo "  docker-run-debug - Run Docker debug container"
	@echo "  docker-clean     - Remove Docker containers and images"
	@echo "  help             - Show this help message"
