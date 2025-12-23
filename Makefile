.PHONY: build test test-coverage clean install help

# Variables
BINARY_NAME=lamp
BUILD_DIR=bin
CMD_DIR=cmd
GO=go
GOTEST=$(GO) test
GOBUILD=$(GO) build

# Default target
all: clean build

# Build the binary for current platform
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build with optimizations and stripped symbols
build-release:
	@echo "Building release binary..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Release build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Install binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install ./$(CMD_DIR)
	@echo "Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

# Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run protocol tests only
test-protocol:
	@echo "Running protocol tests..."
	$(GOTEST) -v ./pkg/protocol/...

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GO) mod tidy

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build binary for current platform"
	@echo "  build-release  - Build optimized release binary"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-protocol  - Run protocol tests only"
	@echo "  bench          - Run benchmarks"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter (requires golangci-lint)"
	@echo "  tidy           - Tidy dependencies"
	@echo "  deps           - Download dependencies"
	@echo "  clean          - Clean build artifacts"
	@echo "  help           - Show this help message"
