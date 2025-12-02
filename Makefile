# Full path: Makefile

.PHONY: all build test bench clean fmt vet lint run run-i list scenario help

# Binary name and paths
BINARY_NAME=meclient
CMD_DIR=./cmd/meclient
BUILD_DIR=./bin
PKG_DIR=./pkg/meclient

# Server settings (override with: make run HOST=192.168.1.10 PORT=5000)
HOST ?= localhost
PORT ?= 1234

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
	$(GO) test -bench=. -benchmem $(PKG_DIR)/...

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

# Run staticcheck
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

# List available scenarios
list: build
	@$(BUILD_DIR)/$(BINARY_NAME) -list

# Run client (connects and waits for messages)
# Usage: make run [HOST=localhost] [PORT=1234]
run: build
	@echo "Connecting to $(HOST):$(PORT)..."
	$(BUILD_DIR)/$(BINARY_NAME) $(HOST) $(PORT)

# Run in interactive mode
# Usage: make run-i [HOST=localhost] [PORT=1234]
run-i: build
	@echo "Interactive mode - connecting to $(HOST):$(PORT)..."
	$(BUILD_DIR)/$(BINARY_NAME) $(HOST) $(PORT) -i

# Run specific scenario
# Usage: make scenario S=1 [HOST=localhost] [PORT=1234]
scenario: build
ifndef S
	@echo "Usage: make scenario S=<scenario_id>"
	@echo "Example: make scenario S=1"
	@echo ""
	@$(BUILD_DIR)/$(BINARY_NAME) -list
else
	@echo "Running scenario $(S) against $(HOST):$(PORT)..."
	$(BUILD_DIR)/$(BINARY_NAME) $(HOST) $(PORT) $(S) -v
endif

# Run with UDP transport
# Usage: make run-udp [HOST=localhost] [PORT=1234]
run-udp: build
	@echo "Connecting via UDP to $(HOST):$(PORT)..."
	$(BUILD_DIR)/$(BINARY_NAME) $(HOST) $(PORT) -udp

# Run with binary protocol (forced)
# Usage: make run-binary [HOST=localhost] [PORT=1234]
run-binary: build
	@echo "Connecting with binary protocol to $(HOST):$(PORT)..."
	$(BUILD_DIR)/$(BINARY_NAME) $(HOST) $(PORT) -binary

# Stress test shortcuts
stress-1k: build
	$(BUILD_DIR)/$(BINARY_NAME) $(HOST) $(PORT) 10

stress-10k: build
	$(BUILD_DIR)/$(BINARY_NAME) $(HOST) $(PORT) 11

stress-100k: build
	$(BUILD_DIR)/$(BINARY_NAME) $(HOST) $(PORT) 12

# Tidy up go.mod
tidy:
	@echo "Tidying go.mod..."
	$(GO) mod tidy
	@echo "Done!"

# Show help
help:
	@echo "Matching Engine Go Client"
	@echo "========================="
	@echo ""
	@echo "Build targets:"
	@echo "  make build           - Build the binary"
	@echo "  make build-release   - Build optimized binary"
	@echo "  make clean           - Remove build artifacts"
	@echo ""
	@echo "Test targets:"
	@echo "  make test            - Run all tests"
	@echo "  make test-coverage   - Run tests with coverage report"
	@echo "  make test-race       - Run tests with race detection"
	@echo "  make bench           - Run benchmarks"
	@echo ""
	@echo "Run targets (default: HOST=localhost PORT=1234):"
	@echo "  make run             - Connect and listen for messages"
	@echo "  make run-i           - Interactive mode"
	@echo "  make run-udp         - Connect via UDP"
	@echo "  make run-binary      - Force binary protocol"
	@echo "  make scenario S=N    - Run scenario N"
	@echo "  make list            - List available scenarios"
	@echo ""
	@echo "Stress tests:"
	@echo "  make stress-1k       - Run 1K order stress test"
	@echo "  make stress-10k      - Run 10K order stress test"
	@echo "  make stress-100k     - Run 100K order stress test"
	@echo ""
	@echo "Code quality:"
	@echo "  make fmt             - Format code"
	@echo "  make vet             - Run go vet"
	@echo "  make lint            - Run staticcheck"
	@echo "  make tidy            - Tidy go.mod"
	@echo ""
	@echo "Override server: make run HOST=192.168.1.10 PORT=5000"
