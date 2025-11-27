.PHONY: all build test bench clean fmt vet lint run-demo run-interactive help

# Binary name and paths
BINARY_NAME=meclient
CMD_DIR=./cmd/example
BUILD_DIR=./bin
PKG_DIR=./pkg/meclient

# Go commands
GO=go
GOFLAGS=-v

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Build successful: $(BUILD_DIR)/$(BINARY_NAME)"

# Build with optimizations (smaller binary, no debug symbols)
build-release:
	@echo "Building $(BINARY_NAME) (release)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Build successful: $(BUILD_DIR)/$(BINARY_NAME)"

# Run all tests
test:
	@echo "Running tests..."
	$(GO) test $(GOFLAGS) ./...
	@echo "All tests passed!"

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	$(GO) test -race ./...
	@echo "No race conditions detected!"

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem $(PKG_DIR)

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Done!"

# Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...
	@echo "No issues found!"

# Run staticcheck (install with: go install honnef.co/go/tools/cmd/staticcheck@latest)
lint:
	@echo "Running staticcheck..."
	@which staticcheck > /dev/null || (echo "Install staticcheck: go install honnef.co/go/tools/cmd/staticcheck@latest" && exit 1)
	staticcheck ./...
	@echo "No issues found!"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Done!"

# Run demo mode (requires server running)
run-demo: build
	@echo "Running demo (connect to localhost:1234)..."
	$(BUILD_DIR)/$(BINARY_NAME) -demo

# Run interactive mode (requires server running)
run-interactive: build
	@echo "Running interactive mode (connect to localhost:1234)..."
	$(BUILD_DIR)/$(BINARY_NAME) -interactive

# Tidy up go.mod
tidy:
	@echo "Tidying go.mod..."
	$(GO) mod tidy
	@echo "Done!"

# Show help
help:
	@echo "Available targets:"
	@echo "  make build           - Build the binary"
	@echo "  make build-release   - Build optimized binary (smaller, no debug)"
	@echo "  make test            - Run all tests"
	@echo "  make test-coverage   - Run tests with coverage report"
	@echo "  make test-race       - Run tests with race detection"
	@echo "  make bench           - Run benchmarks"
	@echo "  make fmt             - Format code"
	@echo "  make vet             - Run go vet"
	@echo "  make lint            - Run staticcheck"
	@echo "  make clean           - Remove build artifacts"
	@echo "  make run-demo        - Build and run demo mode"
	@echo "  make run-interactive - Build and run interactive mode"
	@echo "  make tidy            - Tidy go.mod"
	@echo "  make help            - Show this help"
