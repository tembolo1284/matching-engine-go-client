# Matching Engine Go Client

A high-performance Go client library for interacting with the [matching-engine-c](../matching-engine-c) TCP server. Designed for low-latency order submission and real-time market data reception.

## Features

- **Channel-based API**: Idiomatic Go design using channels for async message delivery
- **Zero dependencies**: Uses only the Go standard library
- **Automatic reconnection**: Configurable exponential backoff with reconnection events
- **Stateless design**: Client is a pure transport layer; server is source of truth
- **Thread-safe**: Safe to call `Send*()` methods from multiple goroutines
- **Cache-line padded stats**: Prevents false sharing in concurrent access
- **Input validation**: All orders validated before sending
- **Pre-allocated buffers**: Minimizes allocations in the hot path

## Installation

```bash
go get github.com/tembolo1284/matching-engine-go-client
```

Or clone and build locally:

```bash
git clone https://github.com/pauljunsukhan/matching-engine-go.git
cd matching-engine-go
go build ./...
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/pauljunsukhan/matching-engine-go/pkg/meclient"
)

func main() {
    // Create client
    cfg := meclient.DefaultConfig("localhost:12345")
    client, err := meclient.New(cfg)
    if err != nil {
        panic(err)
    }

    // Connect
    if err := client.Connect(); err != nil {
        panic(err)
    }
    defer client.Close()

    // Receive messages in background
    go func() {
        for ack := range client.Acks() {
            fmt.Printf("Order %d acknowledged\n", ack.OrderID)
        }
    }()

    go func() {
        for trade := range client.Trades() {
            fmt.Printf("Trade: %d @ %d\n", trade.Qty, trade.Price)
        }
    }()

    // Send an order
    err = client.SendOrder(meclient.NewOrder{
        UserID:  1,
        Symbol:  "IBM",
        Price:   150,
        Qty:     100,
        Side:    meclient.SideBuy,
        OrderID: 1001,
    })
    if err != nil {
        panic(err)
    }

    select {} // Keep running
}
```

## API Reference

### Configuration

```go
cfg := meclient.Config{
    Address:           "localhost:12345",
    ChannelBuffer:     2048,
    ReconnectMinDelay: 50 * time.Millisecond,
    ReconnectMaxDelay: 10 * time.Second,
    ConnectTimeout:    3 * time.Second,
    AutoReconnect:     true,
}
client, err := meclient.New(cfg)
```

### Sending Messages

```go
// New order (price=0 for market order)
err := client.SendOrder(meclient.NewOrder{
    UserID:  1,
    Symbol:  "AAPL",
    Price:   175,
    Qty:     100,
    Side:    meclient.SideBuy,
    OrderID: 1001,
})

// Cancel order
err := client.SendCancel(meclient.CancelOrder{
    Symbol:  "AAPL",
    UserID:  1,
    OrderID: 1001,
})

// Flush all books
err := client.SendFlush()
```

### Receiving Messages

```go
for ack := range client.Acks() { ... }
for trade := range client.Trades() { ... }
for update := range client.BookUpdates() { ... }
for cancelAck := range client.CancelAcks() { ... }
for err := range client.Errors() { ... }
for event := range client.Reconnects() { ... }
```

### Statistics

```go
stats := client.Stats()
fmt.Printf("Sent: %d, Received: %d, Errors: %d\n",
    stats.MessagesSent, stats.MessagesReceived, stats.ErrorCount)
```

## Example CLI

```bash
cd cmd/example
go build

# Listen for messages
./example -addr localhost:12345

# Run demo sequence
./example -addr localhost:12345 -demo

# Interactive mode
./example -addr localhost:12345 -interactive
```

## Testing

```bash
go test ./...                    # Run all tests
go test -v ./pkg/meclient        # Verbose
go test -race ./pkg/meclient     # Race detection
go test -bench=. ./pkg/meclient  # Benchmarks
go test -cover ./pkg/meclient    # Coverage
```

## Documentation

- [Quick Start Guide](docs/QUICK_START.md)
- [Build Guide](docs/BUILD.md)
- [Architecture](docs/ARCHITECTURE.md)

## License

MIT License
