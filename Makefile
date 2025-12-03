# Full path: Makefile

.PHONY: all build test bench clean fmt vet lint run run-tcp run-udp run-udp-binary list scenario help

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

#
# INTERACTIVE MODE TARGETS
#

# Interactive mode (auto-detect: TCP -> UDP fallback)
run: build
	@$(BUILD_DIR)/$(BINARY_NAME) $(HOST) $(PORT) -i

# Interactive via TCP only
run-tcp: build
	@$(BUILD_DIR)/$(BINARY_NAME) $(HOST) $(PORT) -i -tcp

# Interactive via UDP
run-udp: build
	@$(BUILD_DIR)/$(BINARY_NAME) $(HOST) $(PORT) -i -udp

# Interactive via UDP with binary protocol
run-udp-binary: build
	@$(BUILD_DIR)/$(BINARY_NAME) $(HOST) $(PORT) -i -udp -binary

#
# SCENARIO TARGETS
#

# Run specific scenario (auto-detect transport)
scenario: build
ifndef S
	@echo "Usage: make scenario S=<scenario_id>"
	@echo "Example: make scenario S=1"
	@echo ""
	@$(BUILD_DIR)/$(BINARY_NAME) -list
else
	@$(BUILD_DIR)/$(BINARY_NAME) $(HOST) $(PORT) $(S) -v
endif

# Run scenario via UDP
scenario-udp: build
ifndef S
	@echo "Usage: make scenario-udp S=<scenario_id>"
	@echo "Example: make scenario-udp S=1"
else
	@$(BUILD_DIR)/$(BINARY_NAME) $(HOST) $(PORT) $(S) -v -udp
endif

# Run scenario via UDP with binary protocol
scenario-udp-binary: build
ifndef S
	@echo "Usage: make scenario-udp-binary S=<scenario_id>"
	@echo "Example: make scenario-udp-binary S=1"
else
	@$(BUILD_DIR)/$(BINARY_NAME) $(HOST) $(PORT) $(S) -v -udp -binary
endif

#
# STRESS TEST TARGETS
#

stress-1k: build
	@$(BUILD_DIR)/$(BINARY_NAME) $(HOST) $(PORT) 10

stress-10k: build
	@$(BUILD_DIR)/$(BINARY_NAME) $(HOST) $(PORT) 11

stress-1k-udp: build
	@$(BUILD_DIR)/$(BINARY_NAME) $(HOST) $(PORT) 10 -udp

stress-10k-udp: build
	@$(BUILD_DIR)/$(BINARY_NAME) $(HOST) $(PORT) 11 -udp

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
	@echo "Build:"
	@echo "  make build           - Build the binary"
	@echo "  make build-release   - Build optimized binary (smaller)"
	@echo "  make clean           - Remove build artifacts"
	@echo ""
	@echo "Test:"
	@echo "  make test            - Run all tests"
	@echo "  make test-coverage   - Run tests with coverage report"
	@echo "  make test-race       - Run tests with race detection"
	@echo "  make bench           - Run benchmarks"
	@echo ""
	@echo "Interactive Mode (default: HOST=localhost PORT=1234):"
	@echo "  make run             - Interactive (auto-detect transport)"
	@echo "  make run-tcp         - Interactive via TCP only"
	@echo "  make run-udp         - Interactive via UDP"
	@echo "  make run-udp-binary  - Interactive via UDP with binary"
	@echo ""
	@echo "Scenarios:"
	@echo "  make list            - List available scenarios"
	@echo "  make scenario S=N    - Run scenario N (auto-detect)"
	@echo "  make scenario-udp S=N - Run scenario N via UDP"
	@echo ""
	@echo "Stress Tests:"
	@echo "  make stress-1k       - 1K orders (auto-detect)"
	@echo "  make stress-10k      - 10K orders (auto-detect)"
	@echo "  make stress-1k-udp   - 1K orders via UDP"
	@echo "  make stress-10k-udp  - 10K orders via UDP"
	@echo ""
	@echo "Code Quality:"
	@echo "  make fmt             - Format code"
	@echo "  make vet             - Run go vet"
	@echo "  make lint            - Run staticcheck"
	@echo "  make tidy            - Tidy go.mod"
	@echo ""
	@echo "Examples:"
	@echo "  make run                          # Interactive mode"
	@echo "  make run-udp                      # Interactive via UDP"
	@echo "  make run HOST=192.168.1.10        # Different host"
	@echo "  make scenario S=1                 # Run scenario 1"
	@echo "  make scenario-udp S=2             # Scenario 2 via UDP"
