# Build Guide

## Requirements

- Go 1.21+
- Git

## Project Structure

```
matching-engine-go/
├── cmd/
│   └── example/
│       └── main.go
├── pkg/
│   └── meclient/
│       ├── client.go
│       ├── config.go
│       ├── decoder.go
│       ├── encoder.go
│       ├── messages.go
│       ├── pool.go
│       ├── stats.go
│       ├── validation.go
│       └── *_test.go
├── docs/
│   ├── ARCHITECTURE.md
│   ├── BUILD.md
│   └── QUICK_START.md
├── go.mod
└── README.md
```

## Building

```bash
# Build everything
go build ./...

# Build example CLI
cd cmd/example
go build -o meclient-example

# Build with optimizations (smaller binary)
go build -ldflags="-s -w" ./cmd/example
```

## Testing

```bash
# All tests
go test ./...

# Verbose
go test -v ./pkg/meclient

# Specific test
go test -v -run TestClient_SendOrder ./pkg/meclient

# Race detection
go test -race ./pkg/meclient

# Coverage
go test -coverprofile=coverage.out ./pkg/meclient
go tool cover -html=coverage.out
```

## Benchmarks

```bash
# Run benchmarks
go test -bench=. ./pkg/meclient

# With memory stats
go test -bench=. -benchmem ./pkg/meclient

# Specific benchmark
go test -bench=BenchmarkEncoder -benchmem ./pkg/meclient
```

## Static Analysis

```bash
# Built-in vet
go vet ./...

# Install and run staticcheck
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...
```

## Cross-Compilation

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o meclient-linux ./cmd/example

# Windows
GOOS=windows GOARCH=amd64 go build -o meclient.exe ./cmd/example

# ARM64
GOOS=linux GOARCH=arm64 go build -o meclient-arm64 ./cmd/example
```

## Integration Testing

Terminal 1:
```bash
cd matching-engine-c/build
./matching_engine --mode tcp --port 12345
```

Terminal 2:
```bash
cd matching-engine-go/cmd/example
./example -addr localhost:12345 -demo
```
